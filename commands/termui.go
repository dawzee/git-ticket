package commands

import (
	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/termui"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runTermUI(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	return termui.Run(backend)
}

var termUICmd = &cobra.Command{
	Use:     "termui",
	Aliases: []string{"tui"},
	Short:   "Launch the terminal UI.",
	PreRunE: loadRepoEnsureUser,
	RunE:    runTermUI,
}

func init() {
	RootCmd.AddCommand(termUICmd)
}
