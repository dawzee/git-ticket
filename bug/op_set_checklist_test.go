package bug

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/MichaelMure/git-bug/identity"
	"github.com/stretchr/testify/assert"
)

func TestOpSetChecklist_SetChecklist(t *testing.T) {
	var rene = identity.NewBare("René Descarte", "rene@descartes.fr")
	unix := time.Now().Unix()
	bug1 := NewBug()

	before, err := SetChecklist(bug1, rene, unix, Checklist{Label: "123",
		Title: "123 Checklist",
		Questions: []ChecklistQuestion{
			ChecklistQuestion{Question: "1?"},
			ChecklistQuestion{Question: "2?"},
			ChecklistQuestion{Question: "3?"},
		},
	})
	assert.NoError(t, err)

	data, err := json.Marshal(before)
	assert.NoError(t, err)

	var after SetChecklistOperation
	err = json.Unmarshal(data, &after)
	assert.NoError(t, err)

	// enforce creating the IDs
	before.Id()
	rene.Id()

	assert.Equal(t, before, &after)
}
