package bug

import (
	"fmt"
	"strings"
	"time"

	"github.com/MichaelMure/git-bug/entity"
	"github.com/MichaelMure/git-bug/identity"
)

// Snapshot is a compiled form of the Bug data structure used for storage and merge
type Snapshot struct {
	id entity.Id

	Status       Status
	Title        string
	Comments     []Comment
	Labels       []Label
	Checklists   map[string]Checklist
	Author       identity.Interface
	Actors       []identity.Interface
	Participants []identity.Interface
	CreatedAt    time.Time

	Timeline []TimelineItem

	Operations []Operation
}

// Return the Bug identifier
func (snap *Snapshot) Id() entity.Id {
	return snap.id
}

// Return the last time a bug was modified
func (snap *Snapshot) LastEditTime() time.Time {
	if len(snap.Operations) == 0 {
		return time.Unix(0, 0)
	}

	return snap.Operations[len(snap.Operations)-1].Time()
}

// Return the last timestamp a bug was modified
func (snap *Snapshot) LastEditUnix() int64 {
	if len(snap.Operations) == 0 {
		return 0
	}

	return snap.Operations[len(snap.Operations)-1].GetUnixTime()
}

// GetCreateMetadata return the creation metadata
func (snap *Snapshot) GetCreateMetadata(key string) (string, bool) {
	return snap.Operations[0].GetMetadata(key)
}

// SearchTimelineItem will search in the timeline for an item matching the given hash
func (snap *Snapshot) SearchTimelineItem(id entity.Id) (TimelineItem, error) {
	for i := range snap.Timeline {
		if snap.Timeline[i].Id() == id {
			return snap.Timeline[i], nil
		}
	}

	return nil, fmt.Errorf("timeline item not found")
}

// SearchComment will search for a comment matching the given hash
func (snap *Snapshot) SearchComment(id entity.Id) (*Comment, error) {
	for _, c := range snap.Comments {
		if c.id == id {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("comment item not found")
}

// append the operation author to the actors list
func (snap *Snapshot) addActor(actor identity.Interface) {
	for _, a := range snap.Actors {
		if actor.Id() == a.Id() {
			return
		}
	}

	snap.Actors = append(snap.Actors, actor)
}

// append the operation author to the participants list
func (snap *Snapshot) addParticipant(participant identity.Interface) {
	for _, p := range snap.Participants {
		if participant.Id() == p.Id() {
			return
		}
	}

	snap.Participants = append(snap.Participants, participant)
}

// HasParticipant return true if the id is a participant
func (snap *Snapshot) HasParticipant(id entity.Id) bool {
	for _, p := range snap.Participants {
		if p.Id() == id {
			return true
		}
	}
	return false
}

// HasAnyParticipant return true if one of the ids is a participant
func (snap *Snapshot) HasAnyParticipant(ids ...entity.Id) bool {
	for _, id := range ids {
		if snap.HasParticipant(id) {
			return true
		}
	}
	return false
}

// HasActor return true if the id is a actor
func (snap *Snapshot) HasActor(id entity.Id) bool {
	for _, p := range snap.Actors {
		if p.Id() == id {
			return true
		}
	}
	return false
}

// HasAnyActor return true if one of the ids is a actor
func (snap *Snapshot) HasAnyActor(ids ...entity.Id) bool {
	for _, id := range ids {
		if snap.HasActor(id) {
			return true
		}
	}
	return false
}

// Sign post method for gqlgen
func (snap *Snapshot) IsAuthored() {}

// GetChecklists returns a map of checklists associated with this snapshot
func (snap *Snapshot) GetChecklists() (map[string]Checklist, error) {
	checklists := make(map[string]Checklist)

	// Only checklists named in the labels list are currently valid
	for _, l := range snap.Labels {
		lblStr := string(l)
		if strings.HasPrefix(lblStr, "checklist:") {
			var present bool
			checklists[lblStr], present = snap.Checklists[lblStr]
			if !present {
				checklists[lblStr], present = ChecklistStore[lblStr]
				if !present {
					return nil, fmt.Errorf("invalid checklist %s", l)
				}
			}
		}
	}
	return checklists, nil
}

// NextStates returns a slice of next possible states for the assigned workflow
func (snap *Snapshot) NextStates() ([]Status, error) {
	for _, l := range snap.Labels {
		if strings.HasPrefix(string(l), "workflow:") {
			w := FindWorkflow(string(l))
			if w == nil {
				return nil, fmt.Errorf("invalid workflow %s", l)
			}
			return w.NextStates(snap.Status)
		}
	}
	return nil, fmt.Errorf("ticket has no associated workflow")
}

// ValidateTransition returns an error if the supplied state is an invalid
// destination from the current state for the assigned workflow
func (snap *Snapshot) ValidateTransition(newStatus Status) error {
	for _, l := range snap.Labels {
		if strings.HasPrefix(string(l), "workflow:") {
			w := FindWorkflow(string(l))
			if w == nil {
				return fmt.Errorf("invalid workflow %s", l)
			}
			return w.ValidateTransition(snap.Status, newStatus)
		}
	}
	return fmt.Errorf("ticket has no associated workflow")
}
