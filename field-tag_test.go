package bytocol

import "testing"

func TestParseFieldTag(t *testing.T) {
	// Order only
	tag, err := parseFieldTag("3")
	if err != nil {
		t.Error(err)
	} else if tag.Order != 3 {
		t.Errorf("expected order to be 3, got %d", tag.Order)
	}

	// Catch negative values
	_, err = parseFieldTag("-5")
	if err == nil {
		t.Error("expected error for negative order")
	}

	// Catch non-number parsable order
	_, err = parseFieldTag("foo")
	if err == nil {
		t.Error("expected error for non-number order")
	}

	// With null-terminated, space escaped
	tag, err = parseFieldTag("3, null-terminated")
	if err != nil {
		t.Error(err)
	} else if tag.StringLengthPrefix {
		t.Errorf("expected null-terminated string")
	}

	// With length-prefix, space escaped
	tag, err = parseFieldTag("3, length-prefix = 8")
	if err != nil {
		t.Error(err)
	} else if !tag.StringLengthPrefix {
		t.Error("expected to be length-prefixed")
	} else if tag.StringLengthSize != 8 {
		t.Errorf("expected length size to be 8, instead got %d", tag.StringLengthSize)
	}

	// Catch length-prefix non-number
	_, err = parseFieldTag("3, length-prefix=foo")
	if err == nil {
		t.Error("expected error for non-number length-prefix")
	}

	// Catch zero length-prefix
	_, err = parseFieldTag("4, length-prefix=0")
	if err == nil {
		t.Error("expected error for zero length-prefix")
	}

	// Catch incorrect length-prefix
	_, err = parseFieldTag("4,length-prefix=12")
	if err == nil {
		t.Error("expected error on length-prefix size")
	}

}
