package bytocol

import (
	"fmt"
	"reflect"
	"slices"
)

type planEntry struct {
	FieldIndex int
	ValueOf    reflect.Value
	TypeOf     reflect.StructField
	Order      uint
	LengthBits byte
}
type encodingPlan []planEntry

func makePlan(obj any) (encodingPlan, error) {
	plan := make(encodingPlan, 0)

	valOf := reflect.ValueOf(obj)
	typOf := valOf.Type()
	kind := typOf.Kind()

	if kind == reflect.Pointer {
		valOf = valOf.Elem()
		typOf = valOf.Type()
		kind = typOf.Kind()
	}

	if kind != reflect.Struct {
		return plan, ErrNonStruct
	}

	var entry planEntry
	for i := 0; i < typOf.NumField(); i++ {
		entry = planEntry{
			FieldIndex: i,
			ValueOf:    valOf.Field(i),
			TypeOf:     typOf.Field(i),
			LengthBits: 64,
		}

		if !entry.TypeOf.IsExported() {
			continue
		}

		// Parse the field tag if it exists
		tag, hasTag := entry.TypeOf.Tag.Lookup("bytocol")
		if !hasTag {
			continue
		}
		tagInfo, err := parseFieldTag(tag)
		if err != nil {
			return plan, err
		}

		// Apply the order from the tag
		entry.Order = tagInfo.Order

		// Save the plan entry
		plan = append(plan, entry)
	}

	// Sort the plan by the field order
	slices.SortFunc(plan, func(a planEntry, b planEntry) int {
		return int(a.Order) - int(b.Order)
	})

	// Ensure no duplicate-orders
	for i, entry := range plan {
		if i == 0 {
			continue
		}

		if entry.Order == plan[i-1].Order {
			return plan, fmt.Errorf("duplicate field order %d on field #%d", entry.Order, entry.FieldIndex)
		}
	}

	return plan, nil
}
