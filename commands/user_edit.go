package commands

import (
	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/cache"
	"github.com/daedaleanai/git-ticket/input"
)

func newUserEditCommand() *cobra.Command {
	env := newEnv()

	cmd := &cobra.Command{
		Use:   "edit USER-ID",
		Short: "Edit a user identity.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserEdit(env, args)
		},
	}

	return cmd
}

func runUserEdit(env *Env, args []string) error {
	var id *cache.IdentityCache
	var err error
	if len(args) == 1 {
		id, err = env.backend.ResolveIdentityPrefix(args[0])
	} else {
		id, err = env.backend.GetUserIdentity()
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

	return env.backend.UpdateIdentity(id, name, email, "", avatarURL)
}
