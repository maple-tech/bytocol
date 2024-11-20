package bytocol

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

func TestNumberToBytes(t *testing.T) {
	if test := numberToBytes(byte(0xF0)); !bytes.Equal(test, []byte{0xF0}) {
		t.Errorf("incorrect byte: %v", test)
	}
	if test := numberToBytes(uint16(0xF0A0)); !bytes.Equal(test, []byte{0xF0, 0xA0}) {
		t.Errorf("incorrect uint16: %v", test)
	}
	if test := numberToBytes(uint32(0xF0A05020)); !bytes.Equal(test, []byte{0xF0, 0xA0, 0x50, 0x20}) {
		t.Errorf("incorrect uint32: %v", test)
	}
	if test := numberToBytes(uint64(0xF0A05020EEDDCCBB)); !bytes.Equal(test, []byte{0xF0, 0xA0, 0x50, 0x20, 0xEE, 0xDD, 0xCC, 0xBB}) {
		t.Errorf("incorrect uint64: %v", test)
	}

	if test := numberToBytes(int8(-64)); !bytes.Equal(test, []byte{192}) {
		t.Errorf("incorrect int8: %v", test)
	}
	if test := numberToBytes(int16(-1000)); !bytes.Equal(test, []byte{252, 24}) {
		t.Errorf("incorrect int8: %v", test)
	}
	if test := numberToBytes(int32(-5000)); !bytes.Equal(test, []byte{255, 255, 236, 120}) {
		t.Errorf("incorrect int8: %v", test)
	}
	if test := numberToBytes(int64(-1_000_000)); !bytes.Equal(test, []byte{255, 255, 255, 255, 255, 240, 189, 192}) {
		t.Errorf("incorrect int8: %v", test)
	}

	if test := numberToBytes(float32(math.E)); !bytes.Equal(test, []byte{64, 45, 248, 84}) {
		t.Errorf("incorrect float32: %v", test)
	}
	if test := numberToBytes(float64(-math.Pi * 2)); !bytes.Equal(test, []byte{192, 25, 33, 251, 84, 68, 45, 24}) {
		t.Errorf("incorrect float64: %v", test)
	}
}

func testNumber[T number](t *testing.T, num T) {
	data := numberToBytes(num)
	value, err := bytesToNumber[T](data)
	if err != nil {
		t.Error(err)
	} else if value != num {
		t.Errorf("invalid value %v from input %v", value, num)
	}
}

func TestNumbers(t *testing.T) {
	testNumber(t, uint8(0xF0))
	testNumber(t, uint16(0xF0E0))
	testNumber(t, uint32(0xF0E0D0B0))
	testNumber(t, uint64(0xF0E0D0B0A0908070))

	testNumber(t, int8(-64))
	testNumber(t, int16(-5000))
	testNumber(t, int32(-1_000_000))
	testNumber(t, int64(-1_000_000_000))

	testNumber(t, float32(math.Pi))
	testNumber(t, float64(math.E))
}

func TestSetNumberFromBytes(t *testing.T) {
	data := numberToBytes(float32(math.Pi))

	var target float32
	valueOf := reflect.ValueOf(&target)
	if err := setNumberFromBytes[float32](data, valueOf); err != nil {
		t.Error(err)
	} else if target != float32(math.Pi) {
		t.Error("value to not equate")
	}
}
