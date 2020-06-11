package commands

import (
	"errors"
	"fmt"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/commands/select"
	"github.com/daedaleanai/git-ticket/util/interrupt"
	"github.com/spf13/cobra"
)

func runReviewFetch(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	if len(args) < 1 {
		return errors.New("no DiffID supplied")
	}

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
			return errors.New(msg)
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

	if len(review.Comments) == 0 && len(review.Statuses) == 0 {
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
	Use:   "fetch <DiffID> [<id>]",
	Short: "Get Differential Revision data from Phabricator and store in a ticket.",
	Long: `fetch stores Phabricator Differential Revision data in a ticket.

The command takes a Phabricator Differential Revision ID (e.g. D1234) and queries the
Phabricator server for any associated comments or status changes, any resulting data
is stored with the selected ticket. Subsequent calls with the same ID will fetch and
store any updates since the previous call. Multiple Revisions can be stored with a
ticket by running the command with different IDs.

`,
	PreRunE: loadRepoEnsureUser,
	RunE:    runReviewFetch,
}

func init() {
	reviewCmd.AddCommand(reviewFetchCmd)

	reviewFetchCmd.Flags().SortFlags = false
}
