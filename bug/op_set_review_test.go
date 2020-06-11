package bug

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/daedaleanai/git-ticket/identity"
	"github.com/stretchr/testify/assert"
)

var testStatuses = []ReviewStatus{
	ReviewStatus{
		PhabTransaction: PhabTransaction{
			Id:        "10000",
			User:      "USERID1",
			Timestamp: 0},
		Status: "in progress"},
	ReviewStatus{
		PhabTransaction: PhabTransaction{
			Id:        "10005",
			User:      "USERID1",
			Timestamp: 5},
		Status: "on review"},
	ReviewStatus{
		PhabTransaction: PhabTransaction{
			Id:        "10010",
			User:      "USERID1",
			Timestamp: 10},
		Status: "complete"},
}

var testComments = []ReviewComment{
	ReviewComment{
		PhabTransaction: PhabTransaction{
			Id:        "10001",
			User:      "USERID2",
			Timestamp: 1},
		Diff: 123,
		Path: "code/under_test.go",
		Line: 1,
		Text: "needs work"},
	ReviewComment{
		PhabTransaction: PhabTransaction{
			Id:        "10002",
			User:      "USERID2",
			Timestamp: 2},
		Diff: 124,
		Path: "code/under_test.go",
		Line: 1,
		Text: "LGTM"},
}

func TestOpSetReview_SetReview(t *testing.T) {
	var rene = identity.NewBare("René Descarte", "rene@descartes.fr")
	unix := time.Now().Unix()
	bug1 := NewBug()

	before, err := SetReview(bug1, rene, unix,
		&ReviewInfo{RevisionId: "D1234",
			LastTransaction: "12345",
			Statuses:        testStatuses,
			Comments:        testComments,
		})
	assert.NoError(t, err)

	data, err := json.Marshal(before)
	assert.NoError(t, err)

	var after SetReviewOperation
	err = json.Unmarshal(data, &after)
	assert.NoError(t, err)

	// enforce creating the IDs
	before.Id()
	rene.Id()

	assert.Equal(t, before, &after)
}

func TestOpSetReview_Apply(t *testing.T) {

	var rene = identity.NewBare("René Descarte", "rene@descartes.fr")
	unix := time.Now().Unix()
	snapshot := NewBug().Compile()

	// create an operation and apply to the snapshot
	setReviewOp := NewSetReviewOp(rene, unix, &ReviewInfo{RevisionId: "D1234",
		LastTransaction: "12345",
		Statuses:        []ReviewStatus{testStatuses[0]},
		Comments:        []ReviewComment{testComments[0]}})
	setReviewOp.Apply(&snapshot)

	// sumation holds a local copy of what should be in the snapshot
	sumation := ReviewInfo{RevisionId: "D1234",
		LastTransaction: "12345",
		Statuses:        []ReviewStatus{testStatuses[0]},
		Comments:        []ReviewComment{testComments[0]},
	}

	assert.Equal(t, sumation, snapshot.Reviews["D1234"])

	// add an extra comment
	setReviewOp2 := NewSetReviewOp(rene, unix, &ReviewInfo{RevisionId: "D1234",
		LastTransaction: "12346",
		Comments:        []ReviewComment{testComments[1]}})
	setReviewOp2.Apply(&snapshot)

	sumation.Comments = append(sumation.Comments, testComments[1])
	sumation.LastTransaction = "12346"

	assert.Equal(t, sumation, snapshot.Reviews["D1234"])

	// and a couple more status changes
	setReviewOp3 := NewSetReviewOp(rene, unix, &ReviewInfo{RevisionId: "D1234",
		LastTransaction: "12347",
		Statuses:        testStatuses[1:2]})
	setReviewOp3.Apply(&snapshot)

	sumation.Statuses = append(sumation.Statuses, testStatuses[1:2]...)
	sumation.LastTransaction = "12347"

	assert.Equal(t, sumation, snapshot.Reviews["D1234"])
}
