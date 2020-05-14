package input

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/MichaelMure/git-bug/util/interrupt"
)

// PromptValidator is a validator for a user entry
// If complaint is "", value is considered valid, otherwise it's the error reported to the user
// If err != nil, a terminal error happened
type PromptValidator func(name string, value string) (complaint string, err error)

// Required is a validator preventing a "" value
func Required(name string, value string) (string, error) {
	if value == "" {
		return fmt.Sprintf("%s is empty", name), nil
	}
	return "", nil
}

// IsURL is a validator checking that the value is a fully formed URL
func IsURL(name string, value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return fmt.Sprintf("%s is invalid: %v", name, err), nil
	}
	if u.Scheme == "" {
		return fmt.Sprintf("%s is missing a scheme", name), nil
	}
	if u.Host == "" {
		return fmt.Sprintf("%s is missing a host", name), nil
	}
	return "", nil
}

// Prompts

// Prompt is a simple text input.
func Prompt(prompt, name string, validators ...PromptValidator) (string, error) {
	return PromptDefault(prompt, name, "", validators...)
}

// PromptDefault is a simple text input with a default value.
func PromptDefault(prompt, name, preValue string, validators ...PromptValidator) (string, error) {
loop:
	for {
		if preValue != "" {
			_, _ = fmt.Fprintf(os.Stderr, "%s [%s]: ", prompt, preValue)
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s: ", prompt)
		}

		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSpace(line)

		if preValue != "" && line == "" {
			line = preValue
		}

		for _, validator := range validators {
			complaint, err := validator(name, line)
			if err != nil {
				return "", err
			}
			if complaint != "" {
				_, _ = fmt.Fprintln(os.Stderr, complaint)
				continue loop
			}
		}

		return line, nil
	}
}

// PromptPassword is a specialized text input that doesn't display the characters entered.
func PromptPassword(prompt, name string, validators ...PromptValidator) (string, error) {
	termState, err := terminal.GetState(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	cancel := interrupt.RegisterCleaner(func() error {
		return terminal.Restore(int(syscall.Stdin), termState)
	})
	defer cancel()

loop:
	for {
		_, _ = fmt.Fprintf(os.Stderr, "%s: ", prompt)

		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		// new line for coherent formatting, ReadPassword clip the normal new line
		// entered by the user
		fmt.Println()

		if err != nil {
			return "", err
		}

		pass := string(bytePassword)

		for _, validator := range validators {
			complaint, err := validator(name, pass)
			if err != nil {
				return "", err
			}
			if complaint != "" {
				_, _ = fmt.Fprintln(os.Stderr, complaint)
				continue loop
			}
		}

		return pass, nil
	}
}

// PromptChoice is a prompt giving possible choices
// Return the index starting at zero of the choice selected.
func PromptChoice(prompt string, choices []string) (int, error) {
	for {
		for i, choice := range choices {
			_, _ = fmt.Fprintf(os.Stderr, "[%d]: %s\n", i+1, choice)
		}
		_, _ = fmt.Fprintf(os.Stderr, "%s: ", prompt)

		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		fmt.Println()
		if err != nil {
			return 0, err
		}

		line = strings.TrimSpace(line)

		index, err := strconv.Atoi(line)
		if err != nil || index < 1 || index > len(choices) {
			_, _ = fmt.Fprintln(os.Stderr, "invalid input")
			continue
		}

		return index - 1, nil
	}
}
