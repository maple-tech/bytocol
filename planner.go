package bytocol

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
)

type planEntry struct {
	FieldIndex int
	Field      reflect.StructField
	Order      uint
	LengthBits byte
}

// TypePlan is a cached plan for how to encode/decode a given Message type.
type TypePlan struct {
	typeOf        reflect.Type
	typeIndicator byte
	entries       []planEntry
}

// IsValid returns true if this [TypePlan] is considered valid. It is valid if
// there is a type assigned and number of fields is not zero.
func (ep TypePlan) IsValid() bool {
	return ep.typeOf != nil && len(ep.entries) > 0
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

	// Extract the top-level information
	// TODO

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

// Marshal executes an encoding plan using the given object as a value for the
// plan and returns the byte representation of it. The type of the object must
// match the type for the plan.
func (ep TypePlan) Marshal(obj Message) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	valueOf := reflect.ValueOf(obj)
	if valueOf.Type().Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
	}

	typeOf := valueOf.Type()
	if typeOf != ep.typeOf {
		return nil, ErrNonMatchingType
	}

	for _, entry := range ep.entries {
		fieldValue := valueOf.Field(entry.FieldIndex)

		switch entry.Field.Type.Kind() {
		case reflect.Bool:
			err = writeNumber(boolToByte(fieldValue.Bool()), &buf)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			err = writeNumber(fieldValue.Uint(), &buf)
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			err = writeNumber(fieldValue.Int(), &buf)
		case reflect.Float32, reflect.Float64:
			err = writeNumber(fieldValue.Float(), &buf)
		case reflect.String:
			err = writeBlob(fieldValue.String(), entry.LengthBits, &buf)
		case reflect.Slice:
			elem := entry.Field.Type.Elem()
			if elem.Kind() == reflect.Uint8 {
				// Byte slice, use the blob method
				err = writeBlob(fieldValue.Bytes(), entry.LengthBits, &buf)
			} else {
				// UNIMPLEMENTED
				err = fmt.Errorf("bytocol: unsupported slice type %s", elem.String())
			}
		default:
			err = fmt.Errorf("bytocol: unsupported encode type %s", entry.Field.Type.String())
		}

		if err != nil {
			return buf.Bytes(), err
		}
	}

	return buf.Bytes(), nil
}

// PlanType creates a new [TypePlan] based on the generic argument provided.
// This will create a zero-value object of the type and run the reflection process
// to build an encoding/decoding plan for it. It returns nil, and an error if
// something fails in building a plan for the given type.
func PlanType[T Message]() (*TypePlan, error) {
	var zeroValue T
	plan := new(TypePlan)
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
	if err := plan.planObject(obj); err != nil {
		return nil, err
	}
	return plan, nil
}
