package commands

import (
	"fmt"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runReviewFetch(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no DiffID supplied")
	}

	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	diffId := args[0]
	args = args[1:]

	b, args, err := _select.ResolveBug(backend, args)
	if err != nil {
		return err
	}

	// We need an API token to access Phabricator through conduit
	var apiToken string
	if apiToken, err = backend.LocalConfig().ReadString("daedalean.taskmgr-api-token"); err != nil {
		if apiToken, err = backend.GlobalConfig().ReadString("daedalean.taskmgr-api-token"); err != nil {
			msg := `No Phabricator API token set. Please go to
	https://p.daedalean.ai/settings/user/<YOUR_USERNAME_HERE>/page/apitokens/
click on <Generate API Token>, and then paste the token into this command
	git config --global --replace-all daedalean.taskmgr-api-token <PASTE_TOKEN_HERE>`
			return fmt.Errorf(msg)
		}
	}

	// If we already have review data for this Differential then just get any updates
	// since then
	var lastUpdate string
	if existingReview, ok := b.Snapshot().Reviews[diffId]; ok {
		lastUpdate = existingReview.LastTransaction
	}

	review, err := bug.FetchReviewInfo(apiToken, diffId, lastUpdate)
	if err != nil {
		return err
	}

	if len(review.Comments) == 0 && len(review.Status) == 0 {
		fmt.Printf("No updates to save for %s, aborting\n", diffId)
		return nil
	}

	_, err = b.SetReview(review)
	if err != nil {
		return err
	}

	return b.Commit()
}

var reviewFetchCmd = &cobra.Command{
	Use:     "fetch <DiffID> [<id>]",
	Short:   "Fetch review data for a ticket.",
	PreRunE: loadRepoEnsureUser,
	RunE:    runReviewFetch,
}

func init() {
	reviewCmd.AddCommand(reviewFetchCmd)

	reviewFetchCmd.Flags().SortFlags = false
}
