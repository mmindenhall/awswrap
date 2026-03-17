package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

// CLIArgs holds all parsed command-line arguments.
type CLIArgs struct {
	Export      bool
	Exec        string
	Profile     string
	ShowVersion bool
	Command     []string
}

func parseArgs(args []string) (*CLIArgs, error) {
	fs := flag.NewFlagSet("awswrap", flag.ContinueOnError)
	fs.SetInterspersed(false)

	cli := &CLIArgs{}
	fs.BoolVar(&cli.Export, "export", false, "export credentials as environment variables")
	fs.StringVar(&cli.Exec, "exec", "", "execute command via system shell")

	profileDefault := os.Getenv("AWS_PROFILE")
	if profileDefault == "" {
		profileDefault = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	if profileDefault == "" {
		profileDefault = "default"
	}
	fs.StringVar(&cli.Profile, "profile", profileDefault, "the source profile to use for creating credentials")
	fs.BoolVarP(&cli.ShowVersion, "version", "v", false, "print version")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	cli.Command = fs.Args()

	// Validate mutual exclusivity
	if cli.Export && cli.Exec != "" {
		return nil, fmt.Errorf("arguments --export, --exec are mutually exclusive")
	}
	if cli.Exec != "" && len(cli.Command) > 0 {
		return nil, fmt.Errorf("cannot use --exec with positional command arguments")
	}
	if !cli.Export && !cli.ShowVersion && cli.Exec == "" && len(cli.Command) == 0 {
		return nil, fmt.Errorf("no command specified; use --export, --exec, or provide a command")
	}

	return cli, nil
}

func run(args []string) int {
	cli, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	if cli.ShowVersion {
		fmt.Printf("awswrap %s\n", Version)
		return 0
	}

	profile, err := retrieveProfile(cli.Profile, "profile")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var creds *RoleCredentials
	if profile.SourceProfile != nil {
		creds, err = getAssumedRoleCredentials(profile, "")
	} else {
		creds, err = getRoleCredentials(profile, "")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	switch {
	case cli.Export:
		if err := exportCredentials(creds, profile); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	default:
		exitCode, err := runCommand(creds, profile, cli)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return exitCode
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:]))
}
