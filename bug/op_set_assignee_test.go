package bug

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/daedaleanai/git-ticket/identity"
	"github.com/stretchr/testify/assert"
)

func TestSetAssigneeSerialize(t *testing.T) {
	var rene = identity.NewBare("René Descartes", "rene@descartes.fr")
	var mickey = identity.NewBare("Mickey Mouse", "mm@disney.com")
	unix := time.Now().Unix()
	before := NewSetAssigneeOp(rene, unix, mickey)

	data, err := json.Marshal(before)
	assert.NoError(t, err)

	var after SetAssigneeOperation
	err = json.Unmarshal(data, &after)
	assert.NoError(t, err)

	// enforce creating the IDs
	before.Id()
	rene.Id()
	mickey.Id()

	assert.Equal(t, before, &after)
}
