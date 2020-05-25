package commands

import (
	"fmt"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/commands/select"
	"github.com/MichaelMure/git-bug/input"
	"github.com/MichaelMure/git-bug/util/interrupt"
	"github.com/manifoldco/promptui"
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

	id, err := backend.GetUserIdentity()
	if err != nil {
		return err
	}

	ticketChecklists, err := b.Snapshot().GetUserChecklists(id.Id())
	if err != nil {
		return err
	}

	if len(ticketChecklists) == 0 {
		fmt.Println("No checklists associated with ticket")
		return nil
	}

	// Collect checklist labels
	ticketChecklistLabels := make([]string, 0, len(ticketChecklists))

	for k := range ticketChecklists {
		ticketChecklistLabels = append(ticketChecklistLabels, k)
	}

	// If there are multiple checklists associated with the ticket then give the
	// user the option to choose which to edit rather than editing every one

	var selectedChecklistLabel string

	if len(ticketChecklistLabels) > 1 {
		prompt := promptui.Select{
			Label: "Select Checklist",
			Items: ticketChecklistLabels,
		}

		_, selectedChecklistLabel, err = prompt.Run()

		if err != nil {
			return err
		}
	} else {
		selectedChecklistLabel = ticketChecklistLabels[0]
	}

	// Use the editor to edit the checklist, if it changed then create an update
	// operation and commit
	clChange, err := input.ChecklistEditorInput(repo, ticketChecklists[selectedChecklistLabel])
	if err != nil {
		return err
	}

	if clChange {
		_, err = b.SetChecklist(ticketChecklists[selectedChecklistLabel])
		if err != nil {
			return err
		}

		return b.Commit()
	}

	fmt.Println("Checklists unchanged")
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
