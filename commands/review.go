package commands

import (
	"fmt"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/commands/select"
	"github.com/MichaelMure/git-bug/input"
	"github.com/MichaelMure/git-bug/util/interrupt"
	"github.com/spf13/cobra"
)

func runReview(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	b, args, err := _select.ResolveBug(backend, args)
	if err != nil {
		return err
	}

	snap := b.Snapshot()

	currentChecklists, err := snap.GetChecklists()
	if err != nil {
		return err
	}

	// TODO if there are multiple checklists associated with the ticket then give the
	// user the option to choose which to edit rather than forcing them to edit every one

	update := false

	for _, cl := range currentChecklists {
		clChange, err := input.ChecklistEditorInput(repo, cl)
		if err != nil {
			return err
		}

		if clChange {
			update = true
			fmt.Println(cl.Title, "updated")
			_, err = b.SetChecklist(cl)
			if err != nil {
				return err
			}
		}
	}

	if update {
		return b.Commit()
	}
	fmt.Println("checklist unchanged")
	return nil
}

var reviewCmd = &cobra.Command{
	Use:     "review [<id>]",
	Short:   "Review a ticket.",
	PreRunE: loadRepoEnsureUser,
	RunE:    runReview,
}

func init() {
	RootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().SortFlags = false
}
