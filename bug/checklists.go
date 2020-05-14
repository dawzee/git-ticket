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

var ChecklistStore map[string]Checklist

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

func (s ChecklistQuestionState) Validate() error {
	if s < Pending || s > NotApplicable {
		return fmt.Errorf("invalid")
	}

	return nil
}

func init() {
	// TODO put proper checklists here

	// Initialise map of checklists
	ChecklistStore = make(map[string]Checklist)

	ChecklistStore["checklist:dummy-code"] = Checklist{Label: "checklist:dummy-code",
		Title: "Dummy Code Review Checklist",
		Questions: []ChecklistQuestion{
			ChecklistQuestion{Question: "Code review question 1?"},
			ChecklistQuestion{Question: "Code review question 2?"},
			ChecklistQuestion{Question: "Code review question 3?"},
		},
	}

	ChecklistStore["checklist:dummy-test"] = Checklist{Label: "checklist:dummy-test",
		Title: "Dummy Test Review Checklist",
		Questions: []ChecklistQuestion{
			ChecklistQuestion{Question: "Test review question 1?"},
			ChecklistQuestion{Question: "Test review question 2?"},
			ChecklistQuestion{Question: "Test review question 3?"},
		},
	}
}
