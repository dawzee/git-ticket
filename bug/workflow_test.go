package bug

import (
	"testing"
)

func TestNextStates(t *testing.T) {

	// Test the default state transitions
	currentState := [NumStatuses]Status{
		ProposedStatus,
		VettedStatus,
		InProgressStatus,
		InReviewStatus,
		ReviewedStatus,
		AcceptedStatus,
		MergedStatus,
	}
	nextStates := [NumStatuses][]Status{
		{VettedStatus},
		{ProposedStatus, InProgressStatus},
		{InReviewStatus},
		{InProgressStatus, ReviewedStatus},
		{AcceptedStatus},
		{MergedStatus},
		{},
	}

	defaultWf, err := FindWorkflow("workflow:default")
	if err != nil {
		t.Fatal("No default workflow defined")
	}

	for test := ProposedStatus; test < NumStatuses; test++ {
		next, err := defaultWf.NextStates(currentState[test])
		if err != nil || len(next) != len(nextStates[test]) {
			t.Fatal("Invalid default state transition", currentState[test], ">", next, "(error", err, ")")
		}
		for tr, _ := range next {
			if next[tr] != nextStates[test][tr] {
				t.Fatal("Invalid default state transition", currentState[test], ">", next, "(error", err, ")")
			}
		}
	}

	// Test validation of state transition
	if err := defaultWf.ValidateTransition(ProposedStatus, VettedStatus); err != nil {
		t.Fatal("Default state transition proposed > vetted flagged invalid when it isn't")
	}
	if err := defaultWf.ValidateTransition(ProposedStatus, MergedStatus); err == nil {
		t.Fatal("Default state transition proposed > merged flagged valid when it isn't")
	}
}
