package commands

import (
	"github.com/spf13/cobra"
)

func newReviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review actions of a ticket.",
	}

	cmd.AddCommand(newReviewChecklistCommand())
	cmd.AddCommand(newReviewFetchCommand())

	return cmd
}
