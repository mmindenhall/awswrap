package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestParseExpiresAt_ZFormat(t *testing.T) {
	ts, err := parseExpiresAt("2021-01-21T23:30:56Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2021, 1, 21, 23, 30, 56, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, ts)
	}
}

func TestParseExpiresAt_UTCFormat(t *testing.T) {
	ts, err := parseExpiresAt("2020-03-26T13:28:35UTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2020, 3, 26, 13, 28, 35, 0, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, ts)
	}
}

func TestParseExpiresAt_Microseconds(t *testing.T) {
	ts, err := parseExpiresAt("2021-02-18T18:13:41.632177Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2021, 2, 18, 18, 13, 41, 632177000, time.UTC)
	if !ts.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, ts)
	}
}

func TestParseExpiresAt_Invalid(t *testing.T) {
	_, err := parseExpiresAt("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestRetrieveTokenFromFile_Valid(t *testing.T) {
	path, _ := filepath.Abs("testdata/sso_cache_valid.json")
	token, err := retrieveTokenFromFile(path, "https://my-sso-portal.awsapps.com/start", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "valid-test-token-12345" {
		t.Errorf("expected 'valid-test-token-12345', got %q", token)
	}
}

func TestRetrieveTokenFromFile_Expired(t *testing.T) {
	path, _ := filepath.Abs("testdata/sso_cache_expired.json")
	token, err := retrieveTokenFromFile(path, "https://my-sso-portal.awsapps.com/start", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token for expired cache, got %q", token)
	}
}

func TestRetrieveTokenFromFile_WrongURL(t *testing.T) {
	path, _ := filepath.Abs("testdata/sso_cache_valid.json")
	token, err := retrieveTokenFromFile(path, "https://other-portal.awsapps.com/start", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token for wrong URL, got %q", token)
	}
}

func TestRetrieveTokenFromFile_UTCFormat(t *testing.T) {
	path, _ := filepath.Abs("testdata/sso_cache_utc_format.json")
	token, err := retrieveTokenFromFile(path, "https://my-sso-portal.awsapps.com/start", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "utc-format-token" {
		t.Errorf("expected 'utc-format-token', got %q", token)
	}
}

func TestRetrieveTokenFromFile_Microseconds(t *testing.T) {
	path, _ := filepath.Abs("testdata/sso_cache_microseconds.json")
	token, err := retrieveTokenFromFile(path, "https://my-sso-portal.awsapps.com/start", "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "microseconds-token" {
		t.Errorf("expected 'microseconds-token', got %q", token)
	}
}

func TestRetrieveTokenFromFile_NonExistent(t *testing.T) {
	_, err := retrieveTokenFromFile("/nonexistent/path.json", "url", "region")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
