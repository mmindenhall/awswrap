package main

import (
	"testing"
)

func TestParseArgs_Defaults(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_DEFAULT_PROFILE", "")

	cli, err := parseArgs([]string{"--export"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Profile != "default" {
		t.Errorf("expected profile 'default', got %q", cli.Profile)
	}
	if cli.ShowVersion {
		t.Error("expected ShowVersion to be false")
	}
}

func TestParseArgs_ProfileFromEnv(t *testing.T) {
	t.Setenv("AWS_PROFILE", "my-profile")

	cli, err := parseArgs([]string{"--export"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Profile != "my-profile" {
		t.Errorf("expected profile 'my-profile', got %q", cli.Profile)
	}
}

func TestParseArgs_ProfileFromDefaultEnv(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_DEFAULT_PROFILE", "fallback-profile")

	cli, err := parseArgs([]string{"--export"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Profile != "fallback-profile" {
		t.Errorf("expected profile 'fallback-profile', got %q", cli.Profile)
	}
}

func TestParseArgs_ExplicitFlags(t *testing.T) {
	cli, err := parseArgs([]string{"--export", "--profile", "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cli.Export {
		t.Error("expected --export to be true")
	}
	if cli.Profile != "prod" {
		t.Errorf("expected profile 'prod', got %q", cli.Profile)
	}
}

func TestParseArgs_MutualExclusion(t *testing.T) {
	_, err := parseArgs([]string{"--export", "--exec", "echo hi"})
	if err == nil {
		t.Error("expected error for mutually exclusive flags")
	}
}

func TestParseArgs_ExecFlag(t *testing.T) {
	cli, err := parseArgs([]string{"--exec", "echo hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Exec != "echo hello" {
		t.Errorf("expected exec 'echo hello', got %q", cli.Exec)
	}
}

func TestParseArgs_PositionalCommand(t *testing.T) {
	cli, err := parseArgs([]string{"--profile", "dev", "terraform", "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cli.Command) != 2 || cli.Command[0] != "terraform" || cli.Command[1] != "plan" {
		t.Errorf("expected command [terraform plan], got %v", cli.Command)
	}
}

func TestParseArgs_Version(t *testing.T) {
	cli, err := parseArgs([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cli.ShowVersion {
		t.Error("expected --version to be true")
	}
}

func TestParseArgs_ShortVersion(t *testing.T) {
	cli, err := parseArgs([]string{"-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cli.ShowVersion {
		t.Error("expected -v to set ShowVersion true")
	}
}

func TestParseArgs_ExecWithPositionalArgs(t *testing.T) {
	_, err := parseArgs([]string{"--exec", "echo hi", "ls"})
	if err == nil {
		t.Error("expected error for --exec with positional args")
	}
}

func TestParseArgs_NoCommand(t *testing.T) {
	_, err := parseArgs([]string{"--profile", "dev"})
	if err == nil {
		t.Error("expected error when no command, --export, or --exec specified")
	}
}
