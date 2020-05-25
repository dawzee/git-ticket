package commands

import (
	"fmt"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runLabelAdd(cmd *cobra.Command, args []string) error {
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

	changes, _, err := b.ChangeLabels(args, nil)

	for _, change := range changes {
		fmt.Println(change)
	}

	if err != nil {
		return err
	}

	return b.Commit()
}

var labelAddCmd = &cobra.Command{
	Use:     "add [<id>] <label>[...]",
	Short:   "Add a label to a ticket.",
	PreRunE: loadRepoEnsureUser,
	RunE:    runLabelAdd,
}

func init() {
	labelCmd.AddCommand(labelAddCmd)
}
