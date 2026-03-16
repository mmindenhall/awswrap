package main

import (
	"strings"
	"testing"
)

func TestShellQuote_SimpleArgs(t *testing.T) {
	result := shellQuote([]string{"ls", "-la", "foo"})
	if result != "ls -la foo" {
		t.Errorf("expected 'ls -la foo', got %q", result)
	}
}

func TestShellQuote_ArgsWithSpaces(t *testing.T) {
	result := shellQuote([]string{"echo", "hello world"})
	if result != "echo 'hello world'" {
		t.Errorf("expected \"echo 'hello world'\", got %q", result)
	}
}

func TestShellQuote_ArgsWithSingleQuotes(t *testing.T) {
	result := shellQuote([]string{"echo", "it's"})
	expected := "echo 'it'\"'\"'s'"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestShellQuote_ArgsWithMetachars(t *testing.T) {
	result := shellQuote([]string{"echo", "$HOME"})
	if result != "echo '$HOME'" {
		t.Errorf("expected \"echo '$HOME'\", got %q", result)
	}
}

func TestShellQuote_EmptySlice(t *testing.T) {
	result := shellQuote([]string{})
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFilterEnv(t *testing.T) {
	env := []string{
		"AWS_ACCESS_KEY_ID=old",
		"AWS_SECRET_ACCESS_KEY=old",
		"PATH=/usr/bin",
		"HOME=/home/user",
	}
	filtered := filterEnv(env, "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(filtered), filtered)
	}
	if filtered[0] != "PATH=/usr/bin" || filtered[1] != "HOME=/home/user" {
		t.Errorf("unexpected filtered env: %v", filtered)
	}
}

func TestBuildCredentialEnv(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "old-key")
	t.Setenv("AWS_DEFAULT_REGION", "")

	creds := &RoleCredentials{
		AccessKeyID:     "new-key",
		SecretAccessKey: "new-secret",
		SessionToken:    "new-token",
	}
	profile := &Profile{
		Name: "test",
		Raw:  map[string]string{"region": "us-west-2"},
	}

	env := buildCredentialEnv(creds, profile)

	found := map[string]string{}
	for _, e := range env {
		for _, key := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_DEFAULT_REGION"} {
			if len(e) > len(key) && e[:len(key)+1] == key+"=" {
				found[key] = e[len(key)+1:]
			}
		}
	}

	if found["AWS_ACCESS_KEY_ID"] != "new-key" {
		t.Errorf("expected new-key, got %q", found["AWS_ACCESS_KEY_ID"])
	}
	if found["AWS_SECRET_ACCESS_KEY"] != "new-secret" {
		t.Errorf("expected new-secret, got %q", found["AWS_SECRET_ACCESS_KEY"])
	}
	if found["AWS_SESSION_TOKEN"] != "new-token" {
		t.Errorf("expected new-token, got %q", found["AWS_SESSION_TOKEN"])
	}
	if found["AWS_DEFAULT_REGION"] != "us-west-2" {
		t.Errorf("expected us-west-2, got %q", found["AWS_DEFAULT_REGION"])
	}

	// Verify no duplicates of AWS_ACCESS_KEY_ID
	count := 0
	for _, e := range env {
		if strings.HasPrefix(e, "AWS_ACCESS_KEY_ID=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 AWS_ACCESS_KEY_ID entry, got %d", count)
	}
}

func TestRunCommand_ExecMode(t *testing.T) {
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{Name: "test", Raw: map[string]string{}}
	cli := &CLIArgs{Exec: "true"}

	code, err := runCommand(creds, profile, cli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunCommand_PositionalCommand(t *testing.T) {
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{Name: "test", Raw: map[string]string{}}
	cli := &CLIArgs{Command: []string{"true"}}

	code, err := runCommand(creds, profile, cli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunCommand_FailingCommand(t *testing.T) {
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{Name: "test", Raw: map[string]string{}}
	cli := &CLIArgs{Command: []string{"false"}}

	code, err := runCommand(creds, profile, cli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == 0 {
		t.Error("expected non-zero exit code for 'false' command")
	}
}

func TestRunCommand_NoCommand(t *testing.T) {
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{Name: "test", Raw: map[string]string{}}
	cli := &CLIArgs{}

	code, err := runCommand(creds, profile, cli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}
