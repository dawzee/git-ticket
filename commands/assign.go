package commands

import (
	"fmt"
	"strings"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/entity"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runAssign(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return fmt.Errorf("no user supplied")
	}

	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	// TODO allow the user to clear the assignee field
	user := args[0]
	args = args[1:]

	b, args, err := _select.ResolveBug(backend, args)
	if err != nil {
		return err
	}

	// Search through all known users looking for and Id that matches or Name that
	// contains the supplied string

	var assigneeId entity.Id

	for _, id := range backend.AllIdentityIds() {
		i, err := backend.ResolveIdentityExcerpt(id)
		if err != nil {
			return err
		}

		if i.Id.HasPrefix(user) || strings.Contains(i.Name, user) {
			if assigneeId != "" {
				// TODO instead of doing this we could allow the user to select from a list
				return fmt.Errorf("multiple users matching %s", user)
			}
			assigneeId = i.Id
		}
	}

	if assigneeId == "" {
		return fmt.Errorf("no users matching %s", user)
	}

	// Check the ticket is not already assigned to the new assignee
	if b.Snapshot().Assignee != nil {
		currentAssignee, err := backend.ResolveIdentityExcerpt(b.Snapshot().Assignee.Id())
		if err != nil {
			return err
		}
		if assigneeId == currentAssignee.Id {
			return fmt.Errorf("ticket already assigned to %s", currentAssignee.Name)
		}
	}

	// Looks good, get the full identitiy and update the ticket
	i, err := backend.ResolveIdentity(assigneeId)
	if err != nil {
		return err
	}

	_, err = b.SetAssignee(i)
	if err != nil {
		return err
	}

	fmt.Printf("Assigning ticket %s to %s\n", b.Id().Human(), i.DisplayName())

	return b.Commit()
}

var assignCmd = &cobra.Command{
	Use:     "assign <user> [<id>]",
	Short:   "Assign a user to a ticket.",
	PreRunE: loadRepo,
	RunE:    runAssign,
}

func init() {
	RootCmd.AddCommand(assignCmd)
	assignCmd.Flags().SortFlags = false
}
