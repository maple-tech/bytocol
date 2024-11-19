package bytocol

import "testing"

func TestMakePlan(t *testing.T) {
	// Catch non-struct objects
	_, err := makePlan([]byte{})
	if err == nil {
		t.Error("expected error on non-struct")
	}

	type Good struct {
		NonPlanned  int
		Exported    int    `bytocol:"1"`
		Str         string `bytocol:"0,length-prefix=8"`
		nonExported int    `bytocol:"2"`
	}
	good := Good{
		NonPlanned: 3,
		Exported:   10,
		Str:        "hello",
	}

	// Ensure accepts pointers
	plan, err := makePlan(&good)
	if err != nil {
		t.Error(err)
	} else if len(plan) != 2 {
		t.Errorf("expected 2 plan entries, got %d", len(plan))
	} else {
		// Check plan
		if plan[0].FieldIndex != 2 {
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
	_, err = makePlan(Duplicated{})
	if err == nil {
		t.Error("expected error for duplicate order")
	}

	// Catch bad tags
	type BadTags struct {
		Foo int `bytocol:"-1"`
	}
	_, err = makePlan(BadTags{})
	if err == nil {
		t.Error("expected error for bad tag")
	}
}
