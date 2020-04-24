package bug

import (
	"fmt"
	"strings"
)

type ChecklistQuestionState int

const (
	Pending ChecklistQuestionState = iota
	Passed
	Failed
	NotApplicable
)

type ChecklistQuestion struct {
	Question string
	Comment  string
	State    ChecklistQuestionState
}

type Checklist struct {
	Label     string
	Title     string
	Questions []ChecklistQuestion
}

var checklistStore []Checklist

func (s ChecklistQuestionState) String() string {
	switch s {
	case Pending:
		return "PENDING"
	case Passed:
		return "PASSED"
	case Failed:
		return "FAILED"
	case NotApplicable:
		return "NOT APPLICABLE"
	default:
		return "UNKNOWN"
	}
}

func StateFromString(str string) (ChecklistQuestionState, error) {
	cleaned := strings.ToLower(strings.TrimSpace(str))

	switch cleaned {
	case "pending":
		return Pending, nil
	case "passed":
		return Passed, nil
	case "failed":
		return Failed, nil
	case "not applicable":
		return NotApplicable, nil
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

// FindChecklist searches the checklist store by label and returns a pointer to it
func FindChecklist(label string) *Checklist {
	for cl := range checklistStore {
		if checklistStore[cl].Label == label {
			return &checklistStore[cl]
		}
	}
	return nil
}

func init() {
	// TODO put proper checklists here

	// Initialise list of checklists
	checklistStore = []Checklist{
		Checklist{Label: "checklist:code",
			Title: "Code Review Checklist",
			Questions: []ChecklistQuestion{
				ChecklistQuestion{Question: "Is it nice code?"},
				ChecklistQuestion{Question: "Does it compile?"},
				ChecklistQuestion{Question: "Are you sure?"},
			},
		},
		Checklist{Label: "checklist:test",
			Title: "Test Review Checklist",
			Questions: []ChecklistQuestion{
				ChecklistQuestion{Question: "Is it a nice test?"},
				ChecklistQuestion{Question: "Does it pass?"},
				ChecklistQuestion{Question: "Are you sure?"},
			},
		},
	}
}
