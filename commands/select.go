package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
)

func runSelect(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("You must provide a ticket id")
	}

	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	prefix := args[0]

	b, err := backend.ResolveBugPrefix(prefix)
	if err != nil {
		return err
	}

	err = _select.Select(backend, b.Id())
	if err != nil {
		return err
	}

	fmt.Printf("selected ticket %s: %s\n", b.Id().Human(), b.Snapshot().Title)

	return nil
}

var selectCmd = &cobra.Command{
	Use:   "select <id>",
	Short: "Select a ticket for implicit use in future commands.",
	Example: `git ticket select 2f15
git ticket comment
git ticket status
`,
	Long: `Select a ticket for implicit use in future commands.

This command allows you to omit any ticket <id> argument, for example:
  git ticket show
instead of
  git ticket show 2f153ca

The complementary command is "git ticket deselect" performing the opposite operation.
`,
	PreRunE: loadRepo,
	RunE:    runSelect,
}

func init() {
	RootCmd.AddCommand(selectCmd)
	selectCmd.Flags().SortFlags = false
}
