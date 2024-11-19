package bytocol

// MessageInfo is a description object describing how to encode a message. This
// is used by the [bytocol.Message] interface to indicate options for a specific
// message type.
type MessageInfo struct {
	// TypeIndicator is required and must be a unique number within the entire
	// protocol. This is used by the indicator to distinguish between different
	// message formats during serialization/deserialization. Use any value above
	// zero as zero is reserved for error messages.
	TypeIndicator byte

	// DebugName is a helper string to name this message when debugging and for
	// stringification purposes.
	DebugName string
}
