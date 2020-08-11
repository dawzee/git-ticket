package commands

import (
	"fmt"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/input"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var configFile string

func runConfigSet(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	if len(args) != 1 {
		return fmt.Errorf("only one config can be set at a time")
	}

	var configData string

	if configFile != "" {
		configData, err = input.ConfigFileInput(configFile)
		if err != nil {
			return err
		}
	} else {
		currentConfig, err := backend.GetConfig(args[0])
		if err != nil {
			configData, err = input.ConfigEditorInput(backend, "")
		} else {
			configData, err = input.ConfigEditorInput(backend, string(currentConfig))
		}
		if err == input.ErrEmptyMessage {
			fmt.Println("Empty config, aborting.")
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to get config data from the editor: %s", err)
		}
	}

	// Validate json
	var tmp map[string]interface{}
	if err := yaml.Unmarshal([]byte(configData), &tmp); err != nil {
		return fmt.Errorf("the config data you specified is not properly formatted: %s", err)
	}

	return backend.SetConfig(args[0], []byte(configData))
}

var configSetCmd = &cobra.Command{
	Use:     "set <name>",
	Short:   "Set the named configuration data.",
	PreRunE: loadRepo,
	RunE:    runConfigSet,
}

func init() {
	configCmd.AddCommand(configSetCmd)

	configSetCmd.Flags().SortFlags = false

	configSetCmd.Flags().StringVarP(&configFile, "file", "F", "",
		"Take the config from the given file. Use - to read the config from the standard input",
	)
}
