package commands

import (
	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runDeselect(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	err = _select.Clear(backend)
	if err != nil {
		return err
	}

	return nil
}

var deselectCmd = &cobra.Command{
	Use:   "deselect",
	Short: "Clear the implicitly selected ticket.",
	Example: `git ticket select 2f15
git ticket comment
git ticket status
git ticket deselect
`,
	PreRunE: loadRepo,
	RunE:    runDeselect,
}

func init() {
	RootCmd.AddCommand(deselectCmd)
	deselectCmd.Flags().SortFlags = false
}
