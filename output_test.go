package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("cannot create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestExportCredentials_POSIX(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "")
	creds := &RoleCredentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI",
		SessionToken:    "FwoGZXIvYXdz",
	}
	profile := &Profile{
		Name: "test",
		Raw:  map[string]string{"region": "us-west-2"},
	}

	output := captureStdout(t, func() {
		exportCredentials(creds, profile)
	})

	if !strings.Contains(output, "export AWS_ACCESS_KEY_ID='AKIAIOSFODNN7EXAMPLE'") {
		t.Errorf("missing or incorrectly quoted access key in output:\n%s", output)
	}
	if !strings.Contains(output, "export AWS_SECRET_ACCESS_KEY='wJalrXUtnFEMI'") {
		t.Errorf("missing or incorrectly quoted secret key in output:\n%s", output)
	}
	if !strings.Contains(output, "export AWS_SESSION_TOKEN='FwoGZXIvYXdz'") {
		t.Errorf("missing or incorrectly quoted session token in output:\n%s", output)
	}
	if !strings.Contains(output, "export AWS_DEFAULT_REGION='us-west-2'") {
		t.Errorf("missing region in output:\n%s", output)
	}
}

func TestExportCredentials_RegionOmittedWhenSet(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "eu-west-1")
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{
		Name: "test",
		Raw:  map[string]string{"region": "us-west-2"},
	}

	output := captureStdout(t, func() {
		exportCredentials(creds, profile)
	})

	if strings.Contains(output, "AWS_DEFAULT_REGION") {
		t.Errorf("should not export region when AWS_DEFAULT_REGION is already set:\n%s", output)
	}
}

func TestExportCredentials_NoRegionInProfile(t *testing.T) {
	t.Setenv("AWS_DEFAULT_REGION", "")
	creds := &RoleCredentials{
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		SessionToken:    "TOKEN",
	}
	profile := &Profile{
		Name: "test",
		Raw:  map[string]string{},
	}

	output := captureStdout(t, func() {
		exportCredentials(creds, profile)
	})

	if strings.Contains(output, "AWS_DEFAULT_REGION") {
		t.Errorf("should not export region when profile has no region:\n%s", output)
	}
}
