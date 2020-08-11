package bug

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/daedaleanai/git-ticket/entity"
)

func TestSnapshot_GetChecklistCompoundStates(t *testing.T) {
	// Create an initial snapshot, one checklist reviewed by two people, all passed.
	snapshot := Snapshot{
		Labels: []Label{"checklist:XYZ"},
		Checklists: map[string]map[entity.Id]ChecklistSnapshot{
			"checklist:XYZ": map[entity.Id]ChecklistSnapshot{
				"123": ChecklistSnapshot{
					Checklist: Checklist{
						Sections: []ChecklistSection{
							ChecklistSection{
								Questions: []ChecklistQuestion{
									ChecklistQuestion{State: Passed},
									ChecklistQuestion{State: Passed},
								},
							},
							ChecklistSection{
								Questions: []ChecklistQuestion{
									ChecklistQuestion{State: Passed},
									ChecklistQuestion{State: Passed},
								},
							},
						},
					},
					LastEdit: time.Time{},
				},
				"456": ChecklistSnapshot{
					Checklist: Checklist{
						Sections: []ChecklistSection{
							ChecklistSection{
								Questions: []ChecklistQuestion{
									ChecklistQuestion{State: Passed},
									ChecklistQuestion{State: Passed},
								},
							},
							ChecklistSection{
								Questions: []ChecklistQuestion{
									ChecklistQuestion{State: Passed},
									ChecklistQuestion{State: Passed},
								},
							},
						},
					},
					LastEdit: time.Time{},
				},
			},
		},
	}

	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[string]ChecklistState{"checklist:XYZ": Passed})

	// one review has left an answer pending, should still be overall pass
	snapshot.Checklists["checklist:XYZ"]["456"].Checklist.Sections[0].Questions[1].State = Pending
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[string]ChecklistState{"checklist:XYZ": Passed})

	// both reviewers have left an answer pending, should be overall pending
	snapshot.Checklists["checklist:XYZ"]["123"].Checklist.Sections[1].Questions[1].State = Pending
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[string]ChecklistState{"checklist:XYZ": Pending})

	// one review has left an answer failed, should be overall fail
	snapshot.Checklists["checklist:XYZ"]["456"].Checklist.Sections[0].Questions[1].State = Failed
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[string]ChecklistState{"checklist:XYZ": Failed})

	// the default state for an unreviewed checklist is pending
	snapshot.Labels = append(snapshot.Labels, "checklist:ABC")
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[string]ChecklistState{"checklist:XYZ": Failed, "checklist:ABC": Pending})
}
