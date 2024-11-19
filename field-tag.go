package bytocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type fieldTag struct {
	Order              uint
	StringLengthPrefix bool
	StringLengthSize   byte
}

func parseFieldTag(tag string) (fieldTag, error) {
	var err error
	var info fieldTag

	firstComma := strings.IndexRune(tag, ',')
	if firstComma == -1 {
		// Only the order
		u64, err := strconv.ParseUint(tag, 10, 32)
		if err != nil {
			return info, err
		}
		info.Order = uint(u64)
	} else {
		// Contains options, recursively parse the options
		var optionKey string
		var optionValue string
		for _, rawOption := range strings.Split(tag[firstComma+1:], ",") {
			// Check if it has value
			equalInd := strings.IndexRune(rawOption, '=')
			if equalInd == -1 {
				// No value, just the option
				optionKey = strings.TrimSpace(rawOption)
			} else {
				// Has value probably
				optionKey = strings.TrimSpace(rawOption[:equalInd])
				optionValue = strings.TrimSpace(rawOption[equalInd+1:])
			}

			switch optionKey {
			case "null-terminated":
				info.StringLengthPrefix = false
				info.StringLengthSize = 0
			case "length-prefix":
				info.StringLengthPrefix = true
				u64, err := strconv.ParseUint(optionValue, 10, 8)
				if err != nil {
					return info, fmt.Errorf("invalid length-prefix value: %s", err)
				} else if u64 == 0 {
					return info, errors.New("cannot have 0 length-prefix value")
				} else if u64 != 8 && u64 != 16 && u64 != 32 && u64 != 64 {
					return info, fmt.Errorf("length-prefix bit-size %d is invalid, must be 8|16|32|64", u64)
				}
				info.StringLengthSize = byte(u64)
			default:
				return info, fmt.Errorf("invalid option %s in bytocol struct tag", optionKey)
			}
		}
	}

	return info, err
}
