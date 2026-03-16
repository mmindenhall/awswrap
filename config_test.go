package main

import (
	"path/filepath"
	"testing"
)

func testConfigPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("testdata/config")
	if err != nil {
		t.Fatalf("cannot resolve testdata/config: %v", err)
	}
	return abs
}

func setupTestConfig(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_CONFIG_FILE", testConfigPath(t))
}

func TestRetrieveProfile_SSOProfile(t *testing.T) {
	setupTestConfig(t)

	p, err := retrieveProfile("test-sso", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "test-sso" {
		t.Errorf("expected name 'test-sso', got %q", p.Name)
	}
	url, _ := p.getAttribute("sso_start_url")
	if url != "https://my-sso-portal.awsapps.com/start" {
		t.Errorf("unexpected sso_start_url: %q", url)
	}
	region, _ := p.getAttribute("sso_region")
	if region != "us-east-1" {
		t.Errorf("unexpected sso_region: %q", region)
	}
	if p.SourceProfile != nil {
		t.Error("expected no source_profile")
	}
}

func TestRetrieveProfile_AssumeRole(t *testing.T) {
	setupTestConfig(t)

	p, err := retrieveProfile("test-assume", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.SourceProfile == nil {
		t.Fatal("expected source_profile to be resolved")
	}
	if p.SourceProfile.Name != "test-sso" {
		t.Errorf("expected source_profile name 'test-sso', got %q", p.SourceProfile.Name)
	}
	arn, _ := p.getAttribute("role_arn")
	if arn != "arn:aws:iam::123456789012:role/AssumedRole" {
		t.Errorf("unexpected role_arn: %q", arn)
	}
}

func TestRetrieveProfile_SSOSession(t *testing.T) {
	setupTestConfig(t)

	p, err := retrieveProfile("test-session", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.SSOSession == nil {
		t.Fatal("expected sso_session to be resolved")
	}
	if p.SSOSession.Name != "my-session" {
		t.Errorf("expected sso_session name 'my-session', got %q", p.SSOSession.Name)
	}
	// getAttribute should fall back to sso_session
	url, err := p.getAttribute("sso_start_url")
	if err != nil {
		t.Fatalf("getAttribute failed: %v", err)
	}
	if url != "https://my-sso-portal.awsapps.com/start" {
		t.Errorf("unexpected sso_start_url from sso_session fallback: %q", url)
	}
}

func TestRetrieveProfile_NotFound(t *testing.T) {
	setupTestConfig(t)

	_, err := retrieveProfile("nonexistent", "profile")
	if err == nil {
		t.Error("expected error for missing profile")
	}
}

func TestRetrieveProfile_Default(t *testing.T) {
	setupTestConfig(t)

	p, err := retrieveProfile("default", "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	region, _ := p.getAttribute("region")
	if region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %q", region)
	}
}

func TestGetAttribute_Missing(t *testing.T) {
	p := &Profile{Name: "test", Raw: map[string]string{}}
	_, err := p.getAttribute("nonexistent")
	if err == nil {
		t.Error("expected error for missing attribute")
	}
}
