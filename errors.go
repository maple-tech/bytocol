package bytocol

import "errors"

var (
	// Error indicating that a read succeeded, but did not fill the expected content
	// size based on the length provided.
	ErrReadInvariance = errors.New("reading from pipe did not fill expected content size")

	// Error indicating that something might have been written to the [io.Writer],
	// but that the bytes written do not match the original byte length of the
	// data sent.
	ErrWriteInvariance = errors.New("writing to pipe reported incorrect number of bytes compared to data length")

	// Error indicating a type is not valid for encoding/decoding
	ErrNonStruct = errors.New("cannot encode/decode non-struct type")

	// Error indicating the object provided for a plan does not match the type
	// that was used to create the plan.
	ErrNonMatchingType = errors.New("type of object does not match plan type")

	// Error indicating the object provided does not implement the [Message]
	// interface.
	ErrNonMessageType = errors.New("type of object does not implement Message interface")

	// Error indicating that a nil pointer was supplied for unmarshaling
	ErrNilTarget = errors.New("target for unmarshaling is nil")
)

// ErrorMessage is a provided message type built-in for bytocol that wraps a
// standard error message. It uses the reserved type-indicator of 0. The error
// message is transmitted as a 16-bit length-prefixed string allowing for a
// maximum error message size of 65,535 bytes. It uses the debug name "error".
type ErrorMessage struct {
	err error `bytocol:"0,length-prefix=2"`
}

func (e ErrorMessage) BytocolMessage() MessageInfo {
	return MessageInfo{0, "error"}
}

// Error wraps a standard error into an [ErrorMessage] for transmission as a
// message.
func Error(err error) ErrorMessage {
	return ErrorMessage{err}
}

// NewError constructs a new [ErrorMessage] and as such an [error] using the
// given message contents.
func NewError(msg string) ErrorMessage {
	return ErrorMessage{errors.New(msg)}
}
