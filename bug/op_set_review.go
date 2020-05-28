package bug

import (
	"encoding/json"

	"github.com/daedaleanai/git-ticket/entity"
	"github.com/daedaleanai/git-ticket/identity"
	"github.com/daedaleanai/git-ticket/util/timestamp"
)

var _ Operation = &SetReviewOperation{}

// SetReviewOperation will update the review associated with a ticket
type SetReviewOperation struct {
	OpBase
	Review ReviewInfo `json:"review"`
}

//Sign-post method for gqlgen
func (op *SetReviewOperation) IsOperation() {}

func (op *SetReviewOperation) base() *OpBase {
	return &op.OpBase
}

func (op *SetReviewOperation) Id() entity.Id {
	return idOperation(op)
}

func (op *SetReviewOperation) Apply(snapshot *Snapshot) {
	// If there is already review data in the snapshot then append the comments
	// and status data, else just set it
	if r, present := snapshot.Reviews[op.Review.Id]; present {
		r.Comments = append(r.Comments, op.Review.Comments...)
		r.Status = append(r.Status, op.Review.Status...)
		r.LastTransaction = op.Review.LastTransaction
		snapshot.Reviews[op.Review.Id] = r
	} else {
		snapshot.Reviews[op.Review.Id] = op.Review
	}
	snapshot.addActor(op.Author)

	item := &SetReviewTimelineItem{
		id:       op.Id(),
		Author:   op.Author,
		UnixTime: timestamp.Timestamp(op.UnixTime),
		Review:   op.Review,
	}

	snapshot.Timeline = append(snapshot.Timeline, item)
}

func (op *SetReviewOperation) Validate() error {
	if err := opBaseValidate(op, SetReviewOp); err != nil {
		return err
	}

	return nil
}

// UnmarshalJSON is a two step JSON unmarshaling
// This workaround is necessary to avoid the inner OpBase.MarshalJSON
// overriding the outer op's MarshalJSON
func (op *SetReviewOperation) UnmarshalJSON(data []byte) error {
	// Unmarshal OpBase and the op separately

	base := OpBase{}
	err := json.Unmarshal(data, &base)
	if err != nil {
		return err
	}

	aux := struct {
		Review ReviewInfo `json:"review"`
	}{}

	err = json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	op.OpBase = base
	op.Review = aux.Review

	return nil
}

// Sign post method for gqlgen
func (op *SetReviewOperation) IsAuthored() {}

func NewSetReviewOp(author identity.Interface, unixTime int64, review *ReviewInfo) *SetReviewOperation {
	return &SetReviewOperation{
		OpBase: newOpBase(SetReviewOp, author, unixTime),
		Review: *review,
	}
}

type SetReviewTimelineItem struct {
	id       entity.Id
	Author   identity.Interface
	UnixTime timestamp.Timestamp
	Review   ReviewInfo
}

func (s SetReviewTimelineItem) Id() entity.Id {
	return s.id
}

// Sign post method for gqlgen
func (s *SetReviewTimelineItem) IsAuthored() {}

// Convenience function to apply the operation
func SetReview(b Interface, author identity.Interface, unixTime int64, review *ReviewInfo) (*SetReviewOperation, error) {
	setReviewOp := NewSetReviewOp(author, unixTime, review)

	if err := setReviewOp.Validate(); err != nil {
		return nil, err
	}

	b.Append(setReviewOp)
	return setReviewOp, nil
}
