package bytocol

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type planEntry struct {
	FieldIndex int
	Field      reflect.StructField
	Order      uint
	Size       uint
	VarLength  bool
	LengthBits byte
}

func (pe planEntry) String() string {
	var str strings.Builder

	str.WriteString(strconv.FormatUint(uint64(pe.Order), 10))
	str.WriteByte(' ')
	str.WriteString(pe.Field.Name)
	str.WriteByte(' ')
	str.WriteString(pe.Field.Type.String())
	str.WriteByte(' ')
	str.WriteString(strconv.FormatUint(uint64(pe.Size), 10))
	if pe.VarLength {
		str.WriteByte('+')
	}

	return str.String()
}

// TypePlan is a cached plan for how to encode/decode a given Message type.
type TypePlan struct {
	typeOf        reflect.Type
	typeIndicator byte
	debugName     string
	entries       []planEntry
	size          uint
	varLength     bool
}

// IsValid returns true if this [TypePlan] is considered valid. It is valid if
// there is a type assigned and number of fields is not zero.
func (ep TypePlan) IsValid() bool {
	return ep.typeOf != nil && len(ep.entries) > 0
}

// String returns a human-friendly multi-line string outlining the fields, their
// sizes, and types.
func (ep TypePlan) String() string {
	if !ep.IsValid() {
		return "Invalid Plan"
	}

	var str strings.Builder

	// Header
	str.WriteString("Type plan for ")
	str.WriteString(ep.typeOf.Name())
	str.WriteString(" (type=")
	str.WriteString(strconv.FormatUint(uint64(ep.typeIndicator), 10))
	str.WriteString(", name=")
	str.WriteString(ep.debugName)

	if ep.varLength {
		str.WriteString(", min-size=")
		str.WriteString(strconv.FormatUint(uint64(ep.size), 10))
		str.WriteString(", varlen")
	} else {
		str.WriteString(", size=")
		str.WriteString(strconv.FormatUint(uint64(ep.size), 10))
	}

	str.WriteString("):")
	str.WriteByte('\n')

	// Loop through entries
	for i, entry := range ep.entries {
		str.WriteString("| ")
		str.WriteString(entry.String())

		if i < len(ep.entries)-1 {
			str.WriteByte(' ')
		} else {
			str.WriteString(" |")
		}
	}

	return str.String()
}

// Type returns the underlying [reflect.Type] that this plan represents.
func (ep TypePlan) Type() reflect.Type {
	return ep.typeOf
}

// TypeIndicator returns the byte for the message type indicator that will
// be used during transmission. This is equivalent to the [Message] type indicator.
func (ep TypePlan) TypeIndicator() byte {
	return ep.typeIndicator
}

// Name returns the debug name declared by the [Message] when the plan was
// made.
func (ep TypePlan) Name() string {
	return ep.debugName
}

// Size returns the total byte size of a message encoded.
func (ep TypePlan) Size() uint {
	return ep.size
}

// Explain is used to debug byte data that should represent the type and group it
// against the debug string diagram of the plan. It returns a new human-readable
// explanation as a string that can be printed.
func (ep TypePlan) Explain(data []byte) string {
	if len(data) < int(ep.size) {
		return fmt.Sprintf("Invalid data size %d compared to minimum size %d", len(data), ep.size)
	}

	var str strings.Builder

	typeIndicator := data[0]
	str.WriteString("Type Indicator = ")
	str.WriteString(fmt.Sprintf("%03d", typeIndicator))
	str.WriteByte('\n')

	// TODO: Replace this with a tabular table builder.
	// Probably something with a []column kind of type

	// Break down the bytes into their groups, start at 1 because
	// the first is the type indicator
	offset := 1
	for _, entry := range ep.entries {
		str.WriteString(entry.String())
		str.WriteString(" = ")

		byteLength := int(entry.Size)

		// Check if this is a variable length entry
		if entry.VarLength {
			// Protect against overflow
			if (offset + byteLength) >= len(data) {
				str.WriteString("DATA OVERFLOW")
				break
			}

			// The size indicates the length prefix, decode that now
			var length uint64
			if entry.Size == 1 {
				length = uint64(data[offset])
				offset++
			} else if entry.Size == 2 {
				raw, err := bytesToNumber[uint16](data[offset : offset+2])
				if err != nil {
					str.WriteString("ERROR PARSING LENGTH")
				} else {
					length = uint64(raw)
				}
				offset += 2
			} else if entry.Size == 4 {
				raw, err := bytesToNumber[uint32](data[offset : offset+4])
				if err != nil {
					str.WriteString("ERROR PARSING LENGTH")
				} else {
					length = uint64(raw)
				}
				offset += 4
			} else if entry.Size == 8 {
				raw, err := bytesToNumber[uint64](data[offset : offset+8])
				if err != nil {
					str.WriteString("ERROR PARSING LENGTH")
				} else {
					length = uint64(raw)
				}
				offset += 8
			}

			// Print the length first
			str.WriteString("Length=")
			str.WriteString(strconv.FormatUint(length, 10))
			str.WriteString(", ")

			byteLength = int(length)
		}

		// Print out the bytes
		maxOffset := (offset + byteLength)
		for j := offset; j < maxOffset; j++ {
			if j >= len(data) {
				str.WriteString("EOD")
				break
			}

			str.WriteString(fmt.Sprintf("%03d", data[j]))
			if j < maxOffset-1 {
				str.WriteByte(' ')
			}
		}

		// If it was a string print it now
		if entry.Field.Type.Kind() == reflect.String {
			str.WriteString(` "`)
			str.WriteString(string(data[offset : offset+byteLength]))
			str.WriteByte('"')
		}

		offset += byteLength

		str.WriteByte('\n')
	}

	if offset < len(data) {
		str.WriteString(fmt.Sprintf("\n%d bytes remaining", len(data)-offset))
	} else {
		str.WriteString(fmt.Sprintf("All bytes accounted for, total length %d bytes", offset))
	}

	return str.String()
}

// fillTopLevel grabs the [MessageInfo] data from the [Message] interface
// for the object provided. If the object does not implement the interface,
// an [ErrNonMessageType] error is returned.
func (ep *TypePlan) fillTopLevel(obj any) error {
	// Extract the top-level information
	if asMessage, ok := obj.(Message); ok {
		msgInfo := asMessage.BytocolMessage()
		ep.typeIndicator = msgInfo.TypeIndicator
		ep.debugName = msgInfo.DebugName
	} else {
		return ErrNonMessageType
	}
	return nil
}

// planObject internally sets all the variables for this [TypePlan] based on
// the object provided. The object can be a zero value.
func (ep *TypePlan) planObject(obj any) error {
	ep.entries = make([]planEntry, 0)

	ep.typeOf = reflect.TypeOf(obj)
	if ep.typeOf.Kind() == reflect.Pointer {
		// Indirect pointer to type
		ep.typeOf = ep.typeOf.Elem()
	}

	if ep.typeOf.Kind() != reflect.Struct {
		return ErrNonStruct
	}

	ep.size = 0

	// Iterate over all the fields and save them to the plan entries
	var entry planEntry
	for i := 0; i < ep.typeOf.NumField(); i++ {
		entry = planEntry{
			FieldIndex: i,
			Field:      ep.typeOf.Field(i),
			LengthBits: 64,
		}

		if !entry.Field.IsExported() {
			continue
		}

		// Parse the field tag if it exists
		tag, hasTag := entry.Field.Tag.Lookup("bytocol")
		if !hasTag {
			continue
		}
		tagInfo, err := parseFieldTag(tag)
		if err != nil {
			return err
		}

		// Apply the order from the tag
		entry.Order = tagInfo.Order

		// Figure out the encoding size and type
		switch entry.Field.Type.Kind() {
		case reflect.Bool, reflect.Uint8:
			entry.Size = 1
		case reflect.Uint16, reflect.Int16:
			entry.Size = 2
		case reflect.Uint32, reflect.Int32, reflect.Float32:
			entry.Size = 4
		case reflect.Uint64, reflect.Uint, reflect.Int64, reflect.Int, reflect.Float64:
			entry.Size = 8
		case reflect.String:
			entry.Size = uint(entry.LengthBits / 8)
			entry.VarLength = true
			ep.varLength = true
		case reflect.Slice:
			elem := entry.Field.Type.Elem()
			if elem.Kind() == reflect.Uint8 {
				entry.Size = uint(entry.LengthBits / 8)
				entry.VarLength = true
			} else {
				// UNIMPLEMENTED
				err = fmt.Errorf("bytocol: unsupported slice type %s", elem.String())
			}
			ep.varLength = true
		default:
			err = fmt.Errorf("bytocol: unsupported encode type %s", entry.Field.Type.String())
		}

		if err != nil {
			return err
		}
		ep.size += entry.Size

		// Save the plan entry
		ep.entries = append(ep.entries, entry)
	}

	// Sort the plan by the field order
	slices.SortFunc(ep.entries, func(a planEntry, b planEntry) int {
		return int(a.Order) - int(b.Order)
	})

	// Ensure no duplicate-orders
	for i, entry := range ep.entries {
		if i == 0 {
			continue
		}

		if entry.Order == ep.entries[i-1].Order {
			return fmt.Errorf("duplicate field order %d on field #%d", entry.Order, entry.FieldIndex)
		}
	}

	return nil
}

// Write executes an encoding plan using the given object as a value for the
// plan and writes the bytes to the given [io.Writer]. The type of the object must
// match the type for the plan.
func (ep TypePlan) Write(obj Message, w io.Writer) error {
	var err error

	// Write the type indicator first
	_, err = w.Write([]byte{ep.typeIndicator})
	if err != nil {
		return err
	}

	// Figure out the object type
	valueOf := reflect.ValueOf(obj)
	if valueOf.Type().Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
	}

	typeOf := valueOf.Type()
	if typeOf != ep.typeOf {
		return ErrNonMatchingType
	}

	for _, entry := range ep.entries {
		fieldValue := valueOf.Field(entry.FieldIndex)

		switch entry.Field.Type.Kind() {
		case reflect.Bool:
			err = writeNumber(boolToByte(fieldValue.Bool()), w)

		case reflect.Uint8:
			casted, _ := fieldValue.Interface().(uint8)
			err = writeNumber(casted, w)
		case reflect.Uint16:
			casted, _ := fieldValue.Interface().(uint16)
			err = writeNumber(casted, w)
		case reflect.Uint32:
			casted, _ := fieldValue.Interface().(uint32)
			err = writeNumber(casted, w)
		case reflect.Uint64, reflect.Uint:
			casted, _ := fieldValue.Interface().(uint64)
			err = writeNumber(casted, w)

		case reflect.Int8:
			casted, _ := fieldValue.Interface().(int8)
			err = writeNumber(casted, w)
		case reflect.Int16:
			casted, _ := fieldValue.Interface().(int16)
			err = writeNumber(casted, w)
		case reflect.Int32:
			casted, _ := fieldValue.Interface().(int32)
			err = writeNumber(casted, w)
		case reflect.Int64, reflect.Int:
			casted, _ := fieldValue.Interface().(int64)
			err = writeNumber(casted, w)

		case reflect.Float32:
			casted, _ := fieldValue.Interface().(float32)
			err = writeNumber(casted, w)
		case reflect.Float64:
			casted, _ := fieldValue.Interface().(float64)
			err = writeNumber(casted, w)

		case reflect.String:
			err = writeBlob(fieldValue.String(), entry.LengthBits, w)
		case reflect.Slice:
			elem := entry.Field.Type.Elem()
			if elem.Kind() == reflect.Uint8 {
				// Byte slice, use the blob method
				err = writeBlob(fieldValue.Bytes(), entry.LengthBits, w)
			} else {
				// UNIMPLEMENTED
				err = fmt.Errorf("bytocol: unsupported slice type %s", elem.String())
			}
		default:
			err = fmt.Errorf("bytocol: unsupported encode type %s", entry.Field.Type.String())
		}

		if err != nil {
			break
		}
	}

	return err
}

// Marshal executes an encoding plan using the given object as a value for the
// plan and returns the byte representation of it. The type of the object must
// match the type for the plan.
func (ep TypePlan) Marshal(obj Message) ([]byte, error) {
	var buf bytes.Buffer
	err := ep.Write(obj, &buf)
	return buf.Bytes(), err
}

// PlanType creates a new [TypePlan] based on the generic argument provided.
// This will create a zero-value object of the type and run the reflection process
// to build an encoding/decoding plan for it. It returns nil, and an error if
// something fails in building a plan for the given type.
func PlanType[T Message]() (*TypePlan, error) {
	var zeroValue T
	plan := new(TypePlan)

	if err := plan.fillTopLevel(zeroValue); err != nil {
		return nil, err
	}

	if err := plan.planObject(zeroValue); err != nil {
		return nil, err
	}
	return plan, nil
}

// PlanObject creates a new [TypePlan] based on the object provided's type. If
// something fails in building a plan for the given type than nil and the error
// are returned.
func PlanObject(obj Message) (*TypePlan, error) {
	plan := new(TypePlan)

	if err := plan.fillTopLevel(obj); err != nil {
		return nil, err
	}

	if err := plan.planObject(obj); err != nil {
		return nil, err
	}
	return plan, nil
}
