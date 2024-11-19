package bytocol

// Message interface defines any struct that can be encoded using Bytocol and is
// required for any and all messages. It defines the `BytocolMessage` method
// which describes the top-level information for encoding the message.
type Message interface {
	// BytocolMessage indicates that this struct can be encoded as a message. It
	// returns the [bytocol.MessageInfo] type to indicate top-level options for
	// the message. The required value for the [MessageInfo] is the Type-Indicator
	// byte which must be unique for this message across all messages in the
	// protocol.
	BytocolMessage() MessageInfo
}
