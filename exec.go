package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// filterEnv returns os.Environ() with the specified keys removed.
func filterEnv(env []string, keys ...string) []string {
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		exclude := false
		for _, key := range keys {
			if strings.HasPrefix(e, key+"=") {
				exclude = true
				break
			}
		}
		if !exclude {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// buildCredentialEnv returns the current environment with AWS credential vars set.
func buildCredentialEnv(creds *RoleCredentials, profile *Profile) []string {
	env := filterEnv(os.Environ(),
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
	)
	env = append(env,
		"AWS_ACCESS_KEY_ID="+creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY="+creds.SecretAccessKey,
		"AWS_SESSION_TOKEN="+creds.SessionToken,
	)
	if os.Getenv("AWS_DEFAULT_REGION") == "" {
		if region, err := profile.getAttribute("region"); err == nil {
			env = append(env, "AWS_DEFAULT_REGION="+region)
		}
	}
	return env
}

// runCommand executes the wrapped command with AWS credentials in the environment.
func runCommand(creds *RoleCredentials, profile *Profile, cli *CLIArgs) (int, error) {
	env := buildCredentialEnv(creds, profile)

	var cmd *exec.Cmd
	if cli.Exec != "" {
		// --exec: run through system shell
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cli.Exec)
		} else {
			cmd = exec.Command("sh", "-c", cli.Exec)
		}
	} else if len(cli.Command) > 0 {
		// Positional args: run directly
		cmd = exec.Command(cli.Command[0], cli.Command[1:]...)
	} else {
		return 0, nil
	}

	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

// shellQuote quotes arguments for shell execution.
func shellQuote(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t\"'\\$`!#&|;(){}[]<>?*~") {
			quoted[i] = "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
		} else {
			quoted[i] = arg
		}
	}
	return strings.Join(quoted, " ")
}
