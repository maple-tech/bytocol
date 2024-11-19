package bytocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type blob interface {
	string | []byte
}

// Blob to bytes is for encoding, it adds the length prefix.
func blobToBytes[T blob](data T, lenBits byte) []byte {
	// Ensure raw byte data
	var bytData []byte
	if asByts, ok := any(data).([]byte); ok {
		bytData = asByts
	} else if asStr, ok := any(data).(string); ok {
		bytData = []byte(asStr)
	}

	// Setup the buffer
	lenBytes := int(lenBits / 8)
	output := make([]byte, lenBytes+len(bytData))

	switch lenBytes {
	case 1:
		output[0] = byte(len(bytData))
	case 2:
		binary.BigEndian.PutUint16(output[:2], uint16(len(bytData)))
	case 4:
		binary.BigEndian.PutUint32(output[:4], uint32(len(bytData)))
	case 8:
		binary.BigEndian.PutUint64(output[:8], uint64(len(bytData)))
	default:
		panic(fmt.Sprintf("unsupported length bits %d", lenBits))
	}

	copy(output[lenBytes:], bytData)

	return output
}

func writeBlob[T blob](data T, lenBits byte, w io.Writer) (err error) {
	defer func() {
		recErr := recover()
		if recErr != nil {
			err = recErr.(error)
		}
	}()
	byts := blobToBytes(data, lenBits)

	_, wErr := w.Write(byts)
	return wErr
}
