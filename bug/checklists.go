package bug

import (
	"fmt"
	"strings"
	"time"

	"github.com/daedaleanai/git-ticket/util/colors"
)

type ChecklistState int

const (
	Pending ChecklistState = iota
	Passed
	Failed
	NotApplicable
)

type ChecklistQuestion struct {
	Question string
	Comment  string
	State    ChecklistState
}
type ChecklistSection struct {
	Title     string
	Questions []ChecklistQuestion
}
type Checklist struct {
	Label    string
	Title    string
	Sections []ChecklistSection
}
type ChecklistSnapshot struct {
	Checklist
	LastEdit time.Time
}

var ChecklistStore map[string]Checklist

func (s ChecklistState) String() string {
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

func (s ChecklistState) ColorString() string {
	switch s {
	case Pending:
		return colors.Blue("PENDING")
	case Passed:
		return colors.Green("PASSED")
	case Failed:
		return colors.Red("FAILED")
	case NotApplicable:
		return "NOT APPLICABLE"
	default:
		return "UNKNOWN"
	}
}

func StateFromString(str string) (ChecklistState, error) {
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

func (s ChecklistState) Validate() error {
	if s < Pending || s > NotApplicable {
		return fmt.Errorf("invalid")
	}

	return nil
}

// CompoundState returns an overall state for the checklist given the state of
// each of the questions. If any of the questions are Failed then the checklist
// Failed, else if any are Pending it's Pending, else it's Passed
func (c Checklist) CompoundState() ChecklistState {
	var pendingCount, failedCount int
	for _, s := range c.Sections {
		for _, q := range s.Questions {
			switch q.State {
			case Pending:
				pendingCount++
			case Failed:
				failedCount++
			}
		}
	}
	// If at least one question has Failed then return that state
	if failedCount > 0 {
		return Failed
	}
	// None have Failed, but if any are still Pending return that
	if pendingCount > 0 {
		return Pending
	}
	// None Failed or Pending, all questions are NotApplicable or Passed, return Passed
	return Passed
}

func (c Checklist) String() string {
	result := fmt.Sprintf("%s [%s]\n", c.Title, c.CompoundState().ColorString())

	for sn, s := range c.Sections {
		result = result + fmt.Sprintf("#### %s ####\n", s.Title)
		for qn, q := range s.Questions {
			result = result + fmt.Sprintf("(%d.%d) %s [%s]\n", sn+1, qn+1, q.Question, q.State.ColorString())
			if q.Comment != "" {
				result = result + fmt.Sprintf("# %s\n", strings.Replace(q.Comment, "\n", "\n# ", -1))
			}
		}
	}
	return result
}

func init() {
	// TODO put proper checklists here

	// Initialise map of checklists
	ChecklistStore = make(map[string]Checklist)

	ChecklistStore["checklist:dummy-code"] =
		Checklist{Label: "checklist:dummy-code",
			Title: "Dummy Code Review Checklist",
			Sections: []ChecklistSection{
				ChecklistSection{Title: "Code review section 1",
					Questions: []ChecklistQuestion{
						ChecklistQuestion{Question: "Code review question 1?"},
						ChecklistQuestion{Question: "Code review question 2?"},
						ChecklistQuestion{Question: "Code review question 3?"},
					},
				},
				ChecklistSection{Title: "Code review section 2",
					Questions: []ChecklistQuestion{
						ChecklistQuestion{Question: "Code review question 4?"},
						ChecklistQuestion{Question: "Code review question 5?"},
						ChecklistQuestion{Question: "Code review question 6?"},
					},
				},
			},
		}

	ChecklistStore["checklist:dummy-test"] =
		Checklist{Label: "checklist:dummy-test",
			Title: "Dummy Test Review Checklist",
			Sections: []ChecklistSection{
				ChecklistSection{Title: "Test review section 1",
					Questions: []ChecklistQuestion{
						ChecklistQuestion{Question: "Test review question 1?"},
						ChecklistQuestion{Question: "Test review question 2?"},
						ChecklistQuestion{Question: "Test review question 3?"},
					},
				},
				ChecklistSection{Title: "Test review section 2",
					Questions: []ChecklistQuestion{
						ChecklistQuestion{Question: "Test review question 4?"},
						ChecklistQuestion{Question: "Test review question 5?"},
						ChecklistQuestion{Question: "Test review question 6?"},
					},
				},
			},
		}
}
