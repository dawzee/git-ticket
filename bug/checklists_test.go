package bug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecklists_ChecklistCompoundState(t *testing.T) {
	testChecklist := Checklist{Label: "XYZ",
		Title: "XYZ Checklist",
		Sections: []ChecklistSection{
			ChecklistSection{Title: "ABC",
				Questions: []ChecklistQuestion{
					ChecklistQuestion{Question: "1?", State: Passed},
					ChecklistQuestion{Question: "2?", State: Passed},
					ChecklistQuestion{Question: "3?", State: Passed},
				},
			},
			ChecklistSection{Title: "DEF",
				Questions: []ChecklistQuestion{
					ChecklistQuestion{Question: "4?", State: Passed},
					ChecklistQuestion{Question: "5?", State: Passed},
					ChecklistQuestion{Question: "6?", State: Passed},
				},
			},
		},
	}
	assert.Equal(t, testChecklist.CompoundState(), Passed)

	testChecklist.Sections[0].Questions[0].State = NotApplicable
	assert.Equal(t, testChecklist.CompoundState(), Passed)

	testChecklist.Sections[0].Questions[1].State = Pending
	assert.Equal(t, testChecklist.CompoundState(), Pending)

	testChecklist.Sections[0].Questions[2].State = Failed
	assert.Equal(t, testChecklist.CompoundState(), Failed)
}
