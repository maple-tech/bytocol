package bytocol

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

// Encode accepts an incoming object and encodes it into a byte slice for
// transmission. It returns the byte data, and an error if one occurred. If the
// error is non-nil the data is considered garbage in most cases but still
// contains the data up until the error was encountered.
//
// The object must implement [bytocol.Message] interface.
//
// NOTE: This will run the reflection processes on every object entered.
func Encode(obj Message) ([]byte, error) {
	var buf bytes.Buffer

	// Write the type-indicator byte from the message info
	msgInfo := obj.BytocolMessage()
	buf.WriteByte(msgInfo.TypeIndicator)

	// Build an encoding plan containing values in order with their
	// values, types, and encoding options.
	plan, err := makePlan(obj)
	if err != nil {
		return buf.Bytes(), fmt.Errorf("bytocol: cannot encode %s, %s", msgInfo.DebugName, err)
	} else if len(plan) == 0 {
		return buf.Bytes(), fmt.Errorf("bytocol: cannot encode %s, no exported fields", msgInfo.DebugName)
	}

	// Run through the plan, encoding each based on the type.
	for _, entry := range plan {
		switch entry.TypeOf.Type.Kind() {
		case reflect.Bool:
			err = writeNumber(boolToByte(entry.ValueOf.Bool()), &buf)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			err = writeNumber(entry.ValueOf.Uint(), &buf)
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			err = writeNumber(entry.ValueOf.Int(), &buf)
		case reflect.Float32, reflect.Float64:
			err = writeNumber(entry.ValueOf.Float(), &buf)
		case reflect.String:
			err = writeBlob(entry.ValueOf.String(), entry.LengthBits, &buf)
		case reflect.Slice:
			elem := entry.TypeOf.Type.Elem()
			if elem.Kind() == reflect.Uint8 {
				// Byte slice, use the blob method
				err = writeBlob(entry.ValueOf.Bytes(), entry.LengthBits, &buf)
			} else {
				// UNIMPLEMENTED
				err = fmt.Errorf("bytocol: unsupported slice type %s", elem.String())
			}
		default:
			err = fmt.Errorf("bytocol: unsupported encode type %s", entry.TypeOf.Type.String())
		}

		if err != nil {
			return buf.Bytes(), err
		}
	}

	return buf.Bytes(), nil
}

// Write accepts an incoming object, and an [io.Writer]. It encodes the object
// and writes it directly to the writer. It returns an error if one occurred
// during encoding or writing to the pipe. Depending on the error will indicate
// if the message might have been sent or not, but best to assume it did not.
//
// Just like [Encoder.Encode] the object must implement [bytocol.Message] interface,
// as this uses that first and then writes it to the pipe.
//
// NOTE: This will run the reflection process on each object and will not
// cache the encoding plan. Additionally, the byte data is allocated in-memory
// and not streamed directly.
func Write(obj Message, w io.Writer) error {
	data, err := Encode(obj)
	if err != nil {
		return err
	}

	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return ErrWriteInvariance
	}
	return nil
}
