package commands

import (
	"github.com/spf13/cobra"

	_select "github.com/daedaleanai/git-ticket/commands/select"
)

func newDeselectCommand() *cobra.Command {
	env := newEnv()

	var cmd = &cobra.Command{
		Use:   "deselect",
		Short: "Clear the implicitly selected ticket.",
		Example: `git ticket select 2f15
git ticket comment
git ticket status
git ticket deselect
`,
		PreRunE:  loadBackend(env),
		PostRunE: closeBackend(env),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeselect(env)
		},
	}

	return cmd
}

func runDeselect(env *Env) error {
	err := _select.Clear(env.backend)
	if err != nil {
		return err
	}

	return nil
}
