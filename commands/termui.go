package commands

import (
	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/termui"
)

func newTermUICommand() *cobra.Command {
	env := newEnv()

	cmd := &cobra.Command{
		Use:      "termui",
		Aliases:  []string{"tui"},
		Short:    "Launch the terminal UI.",
		PreRunE:  loadBackendEnsureUser(env),
		PostRunE: closeBackend(env),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTermUI(env)
		},
	}

	return cmd
}

func runTermUI(env *Env) error {
	return termui.Run(env.backend)
}
