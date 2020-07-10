package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/input"
	"github.com/daedaleanai/git-ticket/util/interrupt"
)

func runUserEdit(cmd *cobra.Command, args []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	if len(args) > 1 {
		return errors.New("only one identity can be edited at a time")
	}

	var id *cache.IdentityCache
	if len(args) == 1 {
		id, err = backend.ResolveIdentityPrefix(args[0])
	} else {
		id, err = backend.GetUserIdentity()
	}

	if err != nil {
		return err
	}

	name, err := input.PromptDefault("Name", "name", id.DisplayName(), input.Required)
	if err != nil {
		return err
	}

	email, err := input.PromptDefault("Email", "email", id.Email(), input.Required)
	if err != nil {
		return err
	}

	avatarURL, err := input.PromptDefault("Avatar URL", "avatar", id.AvatarUrl())
	if err != nil {
		return err
	}

	return backend.UpdateIdentity(id, name, email, "", avatarURL)
}

var userEditCmd = &cobra.Command{
	Use:     "edit [<user-id>]",
	Short:   "Edit a user identity.",
	PreRunE: loadRepo,
	RunE:    runUserEdit,
}

func init() {
	userCmd.AddCommand(userEditCmd)
	userEditCmd.Flags().SortFlags = false
}
