package bug

import (
	"testing"
)

func TestChecklists_FindChecklist(t *testing.T) {
	if cl := FindChecklist("checklist:code"); cl == nil || cl.Label != "checklist:code" {
		t.Fatal("Finding checklist:code failed")
	}

	if cl := FindChecklist("checklist:test"); cl == nil || cl.Label != "checklist:test" {
		t.Fatal("Finding checklist:test failed")
	}

	if FindChecklist("checklist:XYZGASH") != nil {
		t.Fatal("FindChecklist returned reference to non-existant checklist")
	}
}
