package commands

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/entity"
	"github.com/daedaleanai/git-ticket/util/interrupt"
)

func runPull(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("Only pulling from one remote at a time is supported")
	}

	remote := "origin"
	if len(args) == 1 {
		remote = args[0]
	}

	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	fmt.Println("Fetching remote ...")

	stdout, err := backend.Fetch(remote)
	if err != nil {
		return err
	}

	fmt.Print(stdout)

	fmt.Println("Merging data ...")

	for result := range backend.MergeAll(remote) {
		if result.Err != nil {
			fmt.Println(result.Err)
		}

		if result.Status != entity.MergeStatusNothing {
			fmt.Printf("%s: %s\n", result.Id.Human(), result)
		}
	}

	fmt.Println("Updating configs ...")
	stdout, err = backend.UpdateConfigs(remote)
	fmt.Print(stdout)
	if err != nil {
		return err
	}

	return nil
}

// showCmd defines the "push" subcommand.
var pullCmd = &cobra.Command{
	Use:     "pull [<remote>]",
	Short:   "Pull tickets update from a git remote.",
	PreRunE: loadRepo,
	RunE:    runPull,
}

func init() {
	RootCmd.AddCommand(pullCmd)
}
