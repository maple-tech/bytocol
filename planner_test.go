package bytocol

import (
	"bytes"
	"testing"
)

func TestPlanEntries(t *testing.T) {
	// Catch non-struct objects
	plan := new(TypePlan)
	if err := plan.planObject([]byte{}); err == nil {
		t.Error("expected error on non-struct")
	}

	type Good struct {
		NonPlanned int
		Exported   int    `bytocol:"1"`
		Str        string `bytocol:"0,length-prefix=8"`

		//lint:ignore U1000 This is a test file
		nonExported int `bytocol:"2"`
	}
	good := Good{
		NonPlanned: 3,
		Exported:   10,
		Str:        "hello",
	}

	// Ensure accepts pointers
	err := plan.planObject(&good)
	if err != nil {
		t.Error(err)
	} else if len(plan.entries) != 2 {
		t.Errorf("expected 2 plan entries, got %d", len(plan.entries))
	} else {
		// Check plan
		if plan.entries[0].FieldIndex != 2 {
			t.Error("plan is out of order")
			return
		}
	}

	// Check duplicate orders
	type Duplicated struct {
		First  int `bytocol:"0"`
		Second int `bytocol:"1"`
		Third  int `bytocol:"1"`
		Fourth int `bytocol:"2"`
	}
	err = plan.planObject(Duplicated{})
	if err == nil {
		t.Error("expected error for duplicate order")
	}

	// Catch bad tags
	type BadTags struct {
		Foo int `bytocol:"-1"`
	}
	err = plan.planObject(BadTags{})
	if err == nil {
		t.Error("expected error for bad tag")
	}
}

func TestPlanMarshal(t *testing.T) {
	plan, err := PlanObject(testMessageObj)
	if err != nil {
		t.Error(err)
		return
	}

	data, err := plan.Marshal(&testMessageObj)
	if err != nil {
		t.Error(err)
		return
	}

	if len(data) != testMessageLength {
		t.Errorf("unexpected length %d: %v", len(data), data)
		t.Log(plan.String())
		t.Log(plan.Explain(data))
	}
}

func TestPlanUnmarshal(t *testing.T) {
	// Marshal the test object to get byte data and proove
	// cyclical behaviour
	plan, err := PlanObject(testMessageObj)
	if err != nil {
		t.Error(err)
		return
	}

	data, err := plan.Marshal(testMessageObj)
	if err != nil {
		t.Error(err)
		return
	}

	// Unmarshal and check equivalence
	var result testMessage
	if err = plan.Unmarshal(data[1:], &result); err != nil {
		t.Error(err)
		t.Log(plan.Explain(data))
		return
	}

	if result.Bool != testMessageObj.Bool {
		t.Error("incorrect Bool")
	}
	if result.Int != testMessageObj.Int {
		t.Errorf("incorrect Int %d", result.Int)
	}
	if result.Uint != testMessageObj.Uint {
		t.Error("incorrect Uint")
	}
	if result.Float != testMessageObj.Float {
		t.Error("incorrect Float")
	}
	if result.String != testMessageObj.String {
		t.Error("incorrect String")
	}
	if !bytes.Equal(result.Bytes, testMessageObj.Bytes) {
		t.Error("incorrect Bytes")
	}

	if t.Failed() {
		t.Log(plan.Explain(data))
	}
}
