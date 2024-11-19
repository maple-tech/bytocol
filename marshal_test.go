package bytocol

import (
	"math"
	"testing"
)

type testMessage struct {
	NonEncoded int
	Bool       bool    `bytocol:"0"`
	Uint       uint16  `bytocol:"1"`
	Int        int     `bytocol:"2"`
	Float      float32 `bytocol:"3"`
	String     string  `bytocol:"4"`
	Bytes      []byte  `bytocol:"5"`
}

func (m testMessage) BytocolMessage() MessageInfo {
	return MessageInfo{1, "test"}
}

var testMessageObj = testMessage{0, true, 1234, 4321, math.Pi, "Foo", []byte("Bar")}

const testMessageLength = 1 + (1 + 2 + 8 + 4 + (8 + 3) + (8 + 3))

func TestMarshal(t *testing.T) {
	data, err := Marshal(testMessageObj)
	if err != nil {
		t.Error(err)
		return
	} else if len(data) != testMessageLength {
		t.Errorf("unexpected length %d: %v", len(data), data)
	}
}
