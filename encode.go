package bytocol

import (
	"bytes"
	"fmt"
	"io"
)

// Encode accepts an incoming object and encodes it into a byte slice for
// transmission. It returns the byte data, and an error if one occurred. If the
// error is non-nil the data is considered garbage in most cases but still
// contains the data up until the error was encountered.
//
// The object must implement [bytocol.Message] interface.
//
// NOTE: This will run the reflection processes on every object entered.
func Marshal(obj Message) ([]byte, error) {
	var buf bytes.Buffer

	// Write the type-indicator byte from the message info
	msgInfo := obj.BytocolMessage()
	buf.WriteByte(msgInfo.TypeIndicator)

	// Build an encoding plan containing values in order with their
	// values, types, and encoding options.
	plan, err := PlanObject(obj)
	if err != nil {
		return buf.Bytes(), fmt.Errorf("bytocol: cannot encode %s, %s", msgInfo.DebugName, err)
	} else if !plan.IsValid() {
		return buf.Bytes(), fmt.Errorf("bytocol: cannot encode %s, no exported fields", msgInfo.DebugName)
	}

	// Run through the plan, encoding each based on the type.
	data, err := plan.Marshal(obj)
	if err != nil {
		return buf.Bytes(), err
	}
	buf.Write(data)

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
	data, err := Marshal(obj)
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
