package bytocol

import "testing"

func TestBlobToBytes(t *testing.T) {
	str := "Hello, World"

	// 8-bit
	data := blobToBytes(str, 8)
	if len(data) != (len(str) + 1) {
		t.Errorf("unexpected length %d", len(data))
	} else if string(data[1:]) != str {
		t.Errorf("unexpected body %s", data[1:])
	}

	// 16-bit
	data = blobToBytes(str, 16)
	if len(data) != (len(str) + 2) {
		t.Errorf("unexpected length %d", len(data))
	} else if string(data[2:]) != str {
		t.Errorf("unexpected body %s", data[2:])
	}

	// 32-bit
	data = blobToBytes(str, 32)
	if len(data) != (len(str) + 4) {
		t.Errorf("unexpected length %d", len(data))
	} else if string(data[4:]) != str {
		t.Errorf("unexpected body %s", data[4:])
	}

	// 64-bit
	data = blobToBytes(str, 64)
	if len(data) != (len(str) + 8) {
		t.Errorf("unexpected length %d", len(data))
	} else if string(data[8:]) != str {
		t.Errorf("unexpected body %s", data[8:])
	}

	{
		// Check panic recovery
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected error")
			}
		}()

		_ = blobToBytes([]byte{}, 4)
	}
}
