package commands

import (
	"github.com/MichaelMure/git-bug/bug"
	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/commands/select"
	"github.com/MichaelMure/git-bug/util/interrupt"
	"github.com/spf13/cobra"
)

func runStatusSet(cmd *cobra.Command, args []string) error {

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

	s, _ := bug.StatusFromString(cmd.Annotations["status"])

	_, err = b.SetStatus(s)
	if err != nil {
		return err
	}

	return b.Commit()
}

var setStatusCmds [bug.NumStatuses]*cobra.Command

func init() {

	s := bug.FirstStatus

	for c := 0; c < bug.NumStatuses; c++ {

		setStatusCmds[c] = &cobra.Command{
			Use:     s.String() + " [<id>]",
			Short:   "Ticket is " + s.Action() + ".",
			PreRunE: loadRepoEnsureUser,
			RunE:    runStatusSet,
		}

		// Use the Annotations map to store new status
		setStatusCmds[c].Annotations = make(map[string]string)
		setStatusCmds[c].Annotations["status"] = s.String()

		statusCmd.AddCommand(setStatusCmds[c])

		s++
	}
}
