package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/daedaleanai/git-ticket/validate"
)

func newValidateCommand() *cobra.Command {
	env := newEnv()

	cmd := &cobra.Command{
		Use:      "validate COMMIT...",
		Short:    "Validate identities and commits signatures.",
		PreRunE:  loadBackend(env),
		PostRunE: closeBackend(env),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(env, args)
		},
	}

	return cmd
}

func runValidate(env *Env, args []string) error {
	validator, err := validate.NewValidator(env.backend)
	if err != nil {
		return err
	}

	fmt.Printf("first commit signed with key: %s\n", validator.FirstKey.Fingerprint())

	var refErr error
	for _, ref := range args {
		hash, err := env.backend.ResolveRef(ref)
		if err != nil {
			return err
		}

		_, err = validator.ValidateCommit(hash)
		if err != nil {
			refErr = errors.Wrapf(refErr, "ref %s check fail", ref)
			fmt.Printf("ref %s\tFAIL: %s\n", ref, err.Error())
		} else {
			fmt.Printf("ref %s\tOK\n", ref)
		}
	}
	if refErr != nil {
		return refErr
	}

	return nil
}
