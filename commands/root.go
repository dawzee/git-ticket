// Package commands contains the CLI commands
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/bug"
	"github.com/daedaleanai/git-ticket/identity"
	"github.com/daedaleanai/git-ticket/repository"
)

const rootCommandName = "git-ticket"

// package scoped var to hold the repo after the PreRun execution
var repo repository.ClockedRepo

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   rootCommandName,
	Short: "A ticket tracker embedded in Git.",
	Long: `git-ticket is a ticket tracker embedded in git.

git-ticket use git objects to store the ticket tracking separated from the files
history. As tickets are regular git objects, they can be pushed and pulled from/to
the same git remote your are already using to collaborate with other peoples.

`,

	// For the root command, force the execution of the PreRun
	// even if we just display the help. This is to make sure that we check
	// the repository and give the user early feedback.
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			os.Exit(1)
		}
	},

	SilenceUsage:      true,
	DisableAutoGenTag: true,

	// Custom bash code to connect the git completion for "git ticket" to the
	// git-ticket completion for "git-ticket"
	BashCompletionFunction: `
_git_ticket() {
    __start_git-ticket "$@"
}
`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadRepo is a pre-run function that load the repository for use in a command
func loadRepo(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get the current working directory: %q", err)
	}

	repo, err = repository.NewGitRepo(cwd, bug.Witnesser)
	if err == repository.ErrNotARepo {
		return fmt.Errorf("%s must be run from within a git repo", rootCommandName)
	}

	if err != nil {
		return err
	}

	return nil
}

// loadRepoEnsureUser is the same as loadRepo, but also ensure that the user has configured
// an identity. Use this pre-run function when an error after using the configured user won't
// do.
func loadRepoEnsureUser(cmd *cobra.Command, args []string) error {
	err := loadRepo(cmd, args)
	if err != nil {
		return err
	}

	_, err = identity.GetUserIdentity(repo)
	if err != nil {
		return err
	}

	return nil
}
