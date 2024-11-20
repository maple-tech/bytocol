package bytocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

type number interface {
	byte | uint16 | uint32 | uint64 | uint |
		int8 | int16 | int32 | int64 | int |
		float32 | float64
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func numberToBytes[T number](num T) []byte {
	var data []byte

	switch any(num).(type) {
	case uint8, int8:
		data = []byte{byte(num)}
	case uint16, int16:
		data = make([]byte, 2)
		binary.BigEndian.PutUint16(data, uint16(num))
	case uint32, int32:
		data = make([]byte, 4)
		binary.BigEndian.PutUint32(data, uint32(num))
	case uint64, int64, uint, int:
		data = make([]byte, 8)
		binary.BigEndian.PutUint64(data, uint64(num))
	case float32:
		data = make([]byte, 4)
		binary.BigEndian.PutUint32(data, math.Float32bits(float32(num)))
	case float64:
		data = make([]byte, 8)
		binary.BigEndian.PutUint64(data, math.Float64bits(float64(num)))
	}

	return data
}

func writeNumber[T number](num T, w io.Writer) error {
	data := numberToBytes(num)
	_, err := w.Write(data)
	return err
}

func bytesToNumber[T number](data []byte) (T, error) {
	var value T

	switch any(value).(type) {
	case uint8, int8:
		if len(data) != 1 {
			return value, fmt.Errorf("cannot convert %v bytes to byte", data)
		}
		value = T(data[0])
	case uint16, int16:
		if len(data) != 2 {
			return value, fmt.Errorf("cannot convert %v bytes to uint16", data)
		}
		value = T(binary.BigEndian.Uint16(data))
	case uint32, int32:
		if len(data) != 4 {
			return value, fmt.Errorf("cannot convert %v bytes to uint32", data)
		}
		value = T(binary.BigEndian.Uint32(data))
	case uint64, int64:
		if len(data) != 8 {
			return value, fmt.Errorf("cannot convert %v bytes to uint64", data)
		}
		value = T(binary.BigEndian.Uint64(data))
	case float32:
		if len(data) != 4 {
			return value, fmt.Errorf("cannot convert %v bytes to float32", data)
		}
		ui := binary.BigEndian.Uint32(data)
		value = T(math.Float32frombits(ui))
	case float64:
		if len(data) != 8 {
			return value, fmt.Errorf("cannot convert %v bytes to float64", data)
		}
		ui := binary.BigEndian.Uint64(data)
		value = T(math.Float64frombits(ui))
	}

	return value, nil
}

func readNumber[T number](r io.Reader) (T, error) {
	var value T

	length := 1
	switch any(value).(type) {
	case uint8, int8:
		length = 1
	case uint16, int16:
		length = 2
	case uint32, int32, float32:
		length = 4
	case uint64, int64, float64:
		length = 8
	}

	buf := make([]byte, length)
	n, err := r.Read(buf)

	if n == length {
		value, e := bytesToNumber[T](buf)
		if e != nil {
			return value, e
		}
	}
	return value, err
}

func setNumberFromBytes[T number](data []byte, target reflect.Value) error {
	num, err := bytesToNumber[T](data)
	if err != nil {
		return err
	}

	if target.Type().Kind() == reflect.Pointer {
		target = target.Elem()
	}

	if !target.CanSet() {
		return errors.New("cannot set value, un-addressable")
	}

	switch any(num).(type) {
	case uint8, uint16, uint32, uint64, uint:
		target.SetUint(uint64(num))
	case int8, int16, int32, int64, int:
		target.SetInt(int64(num))
	case float32, float64:
		target.SetFloat(float64(num))
	}

	return nil
}
