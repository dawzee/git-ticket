package commands

import (
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:     "review",
	Short:   "Review actions of a ticket.",
	PreRunE: loadRepo,
}

func init() {
	RootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().SortFlags = false
}
