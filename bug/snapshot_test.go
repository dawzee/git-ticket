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
		Checklists: map[Label]map[entity.Id]ChecklistSnapshot{
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

	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[Label]ChecklistState{"checklist:XYZ": Passed})

	// one review has left an answer TBD, should still be overall pass
	snapshot.Checklists["checklist:XYZ"]["456"].Checklist.Sections[0].Questions[1].State = TBD
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[Label]ChecklistState{"checklist:XYZ": Passed})

	// both reviewers have left an answer TBD, should be overall TBD
	snapshot.Checklists["checklist:XYZ"]["123"].Checklist.Sections[1].Questions[1].State = TBD
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[Label]ChecklistState{"checklist:XYZ": TBD})

	// one review has left an answer failed, should be overall fail
	snapshot.Checklists["checklist:XYZ"]["456"].Checklist.Sections[0].Questions[1].State = Failed
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[Label]ChecklistState{"checklist:XYZ": Failed})

	// the default state for an unreviewed checklist is TBD
	snapshot.Labels = append(snapshot.Labels, "checklist:ABC")
	assert.Equal(t, snapshot.GetChecklistCompoundStates(), map[Label]ChecklistState{"checklist:XYZ": Failed, "checklist:ABC": TBD})
}
