package bug

import "fmt"

type Transition struct {
	start Status
	end   Status
}

type Workflow struct {
	label        string
	initialState Status
	transitions  []Transition
}

var workflowStore []Workflow

// FindWorkflow searches the workflow store by name and returns a pointer to it
func FindWorkflow(name string) *Workflow {

	for wf := range workflowStore {
		if workflowStore[wf].label == name {
			return &workflowStore[wf]
		}
	}
	return nil
}

// NextStates returns a slice of next possible states in the workflow
// for the given one
func (w *Workflow) NextStates(s Status) ([]Status, error) {
	var validStates []Status
	for _, t := range w.transitions {
		if t.start == s {
			validStates = append(validStates, t.end)
		}
	}
	return validStates, nil
}

// ValidateTransition checks if the transition is valid for a given start and end
func (w *Workflow) ValidateTransition(from, to Status) error {
	for _, t := range w.transitions {
		if t.start == from && t.end == to {
			return nil
		}
	}
	return fmt.Errorf("Invalid transition %s -> %s", from, to)
}

func init() {

	// Initialise list of worflows with the default one
	workflowStore = []Workflow{
		Workflow{label: "workflow:default",
			initialState: ProposedStatus,
			transitions: []Transition{
				Transition{start: ProposedStatus, end: VettedStatus},
				Transition{start: VettedStatus, end: ProposedStatus},
				Transition{start: VettedStatus, end: InProgressStatus},
				Transition{start: InProgressStatus, end: InReviewStatus},
				Transition{start: InReviewStatus, end: InProgressStatus},
				Transition{start: InReviewStatus, end: ReviewedStatus},
				Transition{start: ReviewedStatus, end: AcceptedStatus},
				Transition{start: AcceptedStatus, end: MergedStatus},
			},
		},
	}
}
