package main

import (
	"encoding/json"
	"testing"
)

func mockAWSCLI(response interface{}) func() {
	original := runAWSCLI
	data, _ := json.Marshal(response)
	runAWSCLI = func(args []string, profileName string, opts ...cliOptFunc) ([]byte, error) {
		return data, nil
	}
	return func() { runAWSCLI = original }
}

func TestGetRoleCredentials(t *testing.T) {
	// Mock the AWS CLI to return role credentials
	restore := mockAWSCLI(map[string]interface{}{
		"roleCredentials": map[string]interface{}{
			"accessKeyId":     "AKIAIOSFODNN7EXAMPLE",
			"secretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"sessionToken":    "FwoGZXIvYXdzEBY...",
			"expiration":      float64(4102444800000), // 2099-12-31T00:00:00Z in epoch ms
		},
	})
	defer restore()

	// Mock token retrieval by setting up a temp cache dir
	// Instead, we'll mock the full chain by also mocking retrieveToken indirectly
	// For this test, we need a profile with SSO attributes and a valid token
	// Let's test with a direct mock approach

	profile := &Profile{
		Name: "test",
		Raw: map[string]string{
			"sso_start_url":  "https://my-sso-portal.awsapps.com/start",
			"sso_region":     "us-east-1",
			"sso_account_id": "123456789012",
			"sso_role_name":  "MyRole",
		},
	}

	// We need to also handle the token retrieval; let's create a temp SSO cache
	setupTempSSOCache(t)

	creds, err := getRoleCredentials(profile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("unexpected access key: %q", creds.AccessKeyID)
	}
	if creds.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("unexpected secret key: %q", creds.SecretAccessKey)
	}
	if creds.SessionToken != "FwoGZXIvYXdzEBY..." {
		t.Errorf("unexpected session token: %q", creds.SessionToken)
	}
	if creds.Expiration != "2100-01-01T00:00:00Z" {
		t.Errorf("unexpected expiration: %q", creds.Expiration)
	}
}

func TestGetAssumedRoleCredentials(t *testing.T) {
	callCount := 0
	original := runAWSCLI
	runAWSCLI = func(args []string, profileName string, opts ...cliOptFunc) ([]byte, error) {
		callCount++
		if callCount == 1 {
			// First call: get-role-credentials for source profile
			resp := map[string]interface{}{
				"roleCredentials": map[string]interface{}{
					"accessKeyId":     "SOURCE_KEY",
					"secretAccessKey": "SOURCE_SECRET",
					"sessionToken":    "SOURCE_TOKEN",
					"expiration":      float64(4102444800000),
				},
			}
			data, _ := json.Marshal(resp)
			return data, nil
		}
		// Second call: assume-role
		resp := map[string]interface{}{
			"Credentials": map[string]interface{}{
				"AccessKeyId":     "ASSUMED_KEY",
				"SecretAccessKey": "ASSUMED_SECRET",
				"SessionToken":    "ASSUMED_TOKEN",
				"Expiration":      "2099-12-31T00:00:00Z",
			},
		}
		data, _ := json.Marshal(resp)
		return data, nil
	}
	defer func() { runAWSCLI = original }()

	setupTempSSOCache(t)

	sourceProfile := &Profile{
		Name: "source",
		Raw: map[string]string{
			"sso_start_url":  "https://my-sso-portal.awsapps.com/start",
			"sso_region":     "us-east-1",
			"sso_account_id": "123456789012",
			"sso_role_name":  "MyRole",
		},
	}
	profile := &Profile{
		Name:          "assumed",
		SourceProfile: sourceProfile,
		Raw: map[string]string{
			"role_arn":       "arn:aws:iam::123456789012:role/AssumedRole",
			"source_profile": "source",
		},
	}

	creds, err := getAssumedRoleCredentials(profile, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessKeyID != "ASSUMED_KEY" {
		t.Errorf("expected 'ASSUMED_KEY', got %q", creds.AccessKeyID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 AWS CLI calls, got %d", callCount)
	}
}

func TestChooseRefreshableProfile(t *testing.T) {
	tests := []struct {
		name       string
		parent     string
		ssoSession *Profile
		expected   string
	}{
		{"no session", "", nil, ""},
		{"with session, no parent", "", &Profile{Name: "session"}, "my-profile"},
		{"with session and parent", "parent-profile", &Profile{Name: "session"}, "parent-profile"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{Name: "my-profile", SSOSession: tt.ssoSession, Raw: map[string]string{}}
			result := chooseRefreshableProfile(tt.parent, p)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// setupTempSSOCache creates a temporary SSO cache directory with a valid token file.
func setupTempSSOCache(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	cacheDir := tmpDir + "/.aws/sso/cache"
	if err := mkdirAll(cacheDir); err != nil {
		t.Fatalf("cannot create temp cache dir: %v", err)
	}
	tokenData := `{
		"startUrl": "https://my-sso-portal.awsapps.com/start",
		"region": "us-east-1",
		"accessToken": "test-token",
		"expiresAt": "2099-12-31T23:59:59Z"
	}`
	if err := writeFile(cacheDir+"/test.json", tokenData); err != nil {
		t.Fatalf("cannot write temp token file: %v", err)
	}
	t.Setenv("HOME", tmpDir)
}

func mkdirAll(path string) error {
	return __mkdirAll(path)
}

func writeFile(path, content string) error {
	return __writeFile(path, content)
}
