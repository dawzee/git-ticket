package commands

import (
	"fmt"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runStatus(cmd *cobra.Command, args []string) error {
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

	fmt.Println(snap.Status)

	return nil
}

var statusCmd = &cobra.Command{
	Use:     "status [<id>]",
	Short:   "Display or change a ticket status.",
	PreRunE: loadRepo,
	RunE:    runStatus,
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
