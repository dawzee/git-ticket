package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/input"
	"github.com/daedaleanai/git-ticket/util/interrupt"
)

func runReviewChecklist(cmd *cobra.Command, args []string) error {
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
		ticketChecklistLabels = append(ticketChecklistLabels, string(k))
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
	clChange, err := input.ChecklistEditorInput(repo, ticketChecklists[bug.Label(selectedChecklistLabel)])
	if err != nil {
		return err
	}

	if clChange {
		_, err = b.SetChecklist(ticketChecklists[bug.Label(selectedChecklistLabel)])
		if err != nil {
			return err
		}

		return b.Commit()
	}

	fmt.Println("Checklists unchanged")
	return nil
}

var reviewChecklistCmd = &cobra.Command{
	Use:     "checklist [<id>]",
	Short:   "Complete a checklist associated with a ticket.",
	PreRunE: loadRepoEnsureUser,
	RunE:    runReviewChecklist,
}

func init() {
	reviewCmd.AddCommand(reviewChecklistCmd)

	reviewChecklistCmd.Flags().SortFlags = false
}
