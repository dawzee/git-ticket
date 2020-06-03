package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/util/interrupt"
)

func runConfig(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	if len(args) > 1 {
		return fmt.Errorf("only one config can be displayed at a time")
	}

	if len(args) == 1 {
		data, err := backend.GetConfig(args[0])

		if err != nil {
			return fmt.Errorf("failed to get config %s: %s", args[0], err)
		}

		fmt.Println(string(data))
	} else {
		configs, err := backend.ListConfigs()
		if err != nil {
			return fmt.Errorf("failed to list configs: %s", err)
		}

		for _, config := range configs {
			fmt.Printf("%s\n", config)
		}
	}

	return nil
}

var configCmd = &cobra.Command{
	Use:     "config <config-name>",
	Short:   "List configs or show the specified config",
	PreRunE: loadRepo,
	RunE:    runConfig,
}

func init() {
	RootCmd.AddCommand(configCmd)
}
