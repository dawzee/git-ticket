package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/commands/select"
	"github.com/MichaelMure/git-bug/entity"
	"github.com/MichaelMure/git-bug/util/interrupt"
	"github.com/spf13/cobra"
)

func runAssign(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	// TODO allow the user to clear the assignee field
	user := args[0]

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
				fmt.Printf("Multiple users matching %s\n", user)
				return nil
			}
			assigneeId = i.Id
		}
	}

	if assigneeId == "" {
		fmt.Printf("No users matching %s\n", user)
		return nil
	}

	// Check the ticket is not already assigned to the new assignee
	currentAssignee, err := backend.ResolveIdentityExcerpt(b.Snapshot().Assignee.Id())
	if err != nil {
		return err
	}
	if assigneeId == currentAssignee.Id {
		fmt.Printf("Ticket already assigned to %s\n", currentAssignee.Name)
		return nil
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

	_, _ = fmt.Fprintf(os.Stderr, "Assigning ticket to: %s\n", i.DisplayName())

	return b.Commit()
}

var assignCmd = &cobra.Command{
	Use:     "assign <user>",
	Short:   "Assign a user.",
	PreRunE: loadRepo,
	RunE:    runAssign,
}

func init() {
	RootCmd.AddCommand(assignCmd)
	assignCmd.Flags().SortFlags = false
}
