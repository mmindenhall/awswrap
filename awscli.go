package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// cliOption is a functional option for callAWSCLI.
type cliOption struct {
	appendProfile bool
	env           []string
	errorMessage  string
}

type cliOptFunc func(*cliOption)

func withProfile() cliOptFunc {
	return func(o *cliOption) {
		o.appendProfile = true
	}
}

func withEnv(env []string) cliOptFunc {
	return func(o *cliOption) {
		o.env = env
	}
}

func withErrorMessage(msg string) cliOptFunc {
	return func(o *cliOption) {
		o.errorMessage = msg
	}
}

// runAWSCLI is a package-level variable to allow mocking in tests.
var runAWSCLI = callAWSCLI

// callAWSCLI executes the aws CLI with the given arguments and options.
func callAWSCLI(args []string, profileName string, opts ...cliOptFunc) ([]byte, error) {
	o := &cliOption{}
	for _, fn := range opts {
		fn(o)
	}

	finalArgs := append([]string{}, args...)
	if o.appendProfile {
		finalArgs = append(finalArgs, "--profile", profileName)
	}
	finalArgs = append(finalArgs, "--output", "json", "--no-cli-auto-prompt")

	cmd := exec.Command("aws", finalArgs...)
	if o.env != nil {
		cmd.Env = o.env
	}

	out, err := cmd.Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("cannot execute AWS CLI: %w (is 'aws' installed and on your PATH?)", execErr)
		}
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			fmt.Fprint(os.Stderr, string(exitErr.Stderr))
		}
		errMsg := o.errorMessage
		if errMsg == "" {
			errMsg = fmt.Sprintf("please login with 'aws sso login --profile=%s'", profileName)
		}
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}
	return out, nil
}
