package bytocol

import (
	"bytes"
	"fmt"
	"io"
)

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
	// Grab the message info for the error messages, the plan will
	// write the actual type indicator
	msgInfo := obj.BytocolMessage()

	// Build an encoding plan containing values in order with their
	// values, types, and encoding options.
	plan, err := PlanObject(obj)
	if err != nil {
		return fmt.Errorf("bytocol: cannot encode %s, %s", msgInfo.DebugName, err)
	} else if !plan.IsValid() {
		return fmt.Errorf("bytocol: cannot encode %s, no exported fields", msgInfo.DebugName)
	}

	// Encode the plan onto the Writer
	return plan.Write(obj, w)
}

// Marshal accepts an incoming object and encodes it into a byte slice for
// transmission. It returns the byte data, and an error if one occurred. If the
// error is non-nil the data is considered garbage in most cases but still
// contains the data up until the error was encountered.
//
// The object must implement [bytocol.Message] interface.
//
// NOTE: This will run the reflection processes on every object entered.
func Marshal(obj Message) ([]byte, error) {
	var buf bytes.Buffer
	err := Write(obj, &buf)
	return buf.Bytes(), err
}
