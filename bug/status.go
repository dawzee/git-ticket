package bug

import (
	"fmt"
	"strings"
)

type Status int

const (
	_ Status = iota
	ProposedStatus
	VettedStatus
	InProgressStatus
	InReviewStatus
	ReviewedStatus
	AcceptedStatus
	MergedStatus
)

const FirstStatus = ProposedStatus
const LastStatus = MergedStatus
const NumStatuses = 7

func (s Status) String() string {
	switch s {
	case ProposedStatus:
		return "proposed"
	case VettedStatus:
		return "vetted"
	case InProgressStatus:
		return "inprogress"
	case InReviewStatus:
		return "inreview"
	case ReviewedStatus:
		return "reviewed"
	case AcceptedStatus:
		return "accepted"
	case MergedStatus:
		return "merged"
	default:
		return "unknown status"
	}
}

func (s Status) Action() string {
	switch s {
	case ProposedStatus:
		return "proposed"
	case VettedStatus:
		return "vetted"
	case InProgressStatus:
		return "set to in progress"
	case InReviewStatus:
		return "set to in review"
	case ReviewedStatus:
		return "reviewed"
	case AcceptedStatus:
		return "accepted"
	case MergedStatus:
		return "merged"
	default:
		return "unknown status"
	}
}

func StatusFromString(str string) (Status, error) {
	cleaned := strings.ToLower(strings.TrimSpace(str))

	switch cleaned {
	case "proposed":
		return ProposedStatus, nil
	case "vetted":
		return VettedStatus, nil
	case "inprogress":
		return InProgressStatus, nil
	case "inreview":
		return InReviewStatus, nil
	case "reviewed":
		return ReviewedStatus, nil
	case "accepted":
		return AcceptedStatus, nil
	case "merged":
		return MergedStatus, nil
	default:
		return 0, fmt.Errorf("unknown status")
	}
}

func (s Status) Validate() error {
	if s < ProposedStatus || s > MergedStatus {
		return fmt.Errorf("invalid")
	}

	return nil
}
