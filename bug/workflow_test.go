package bug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the default state transitions
var nextStates [][]Status = [][]Status{
	nil, // first status is 1
	{VettedStatus},
	{ProposedStatus, InProgressStatus},
	{InReviewStatus},
	{InProgressStatus, ReviewedStatus},
	{AcceptedStatus},
	{MergedStatus},
	nil,
}

func TestWorkflow_NextStates(t *testing.T) {

	defaultWf := FindWorkflow("workflow:default")
	if defaultWf == nil {
		t.Fatal("No default workflow defined")
	}

	for currentState := FirstStatus; currentState <= LastStatus; currentState++ {
		next, err := defaultWf.NextStates(currentState)
		if err != nil {
			t.Fatal("Invalid default state transition", currentState, ">", next, "(error", err, ")")
		}
		assert.Equal(t, nextStates[currentState], next)
	}

}

func TestWorkflow_ValidateTransition(t *testing.T) {

	defaultWf := FindWorkflow("workflow:default")
	if defaultWf == nil {
		t.Fatal("No default workflow defined")
	}

	// Test validation of state transition
	for from := FirstStatus; from <= LastStatus; from++ {
		for _, to := range nextStates[from] {
			if err := defaultWf.ValidateTransition(from, to); err != nil {
				t.Fatal("Default state transition " + from.String() + " > " + to.String() + " flagged invalid when it isn't")
			}
		}
	}

	if err := defaultWf.ValidateTransition(ProposedStatus, MergedStatus); err == nil {
		t.Fatal("Default state transition proposed > merged flagged valid when it isn't")
	}

}
