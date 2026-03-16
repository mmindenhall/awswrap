package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SSOCacheEntry represents a cached SSO token file.
type SSOCacheEntry struct {
	StartURL    string `json:"startUrl"`
	Region      string `json:"region"`
	ExpiresAt   string `json:"expiresAt"`
	AccessToken string `json:"accessToken"`
}

// retrieveToken gets the access token from the SSO cache, optionally refreshing if expired.
func retrieveToken(startURL, region, profileName string, refreshProfile string) (string, error) {
	token, err := retrieveTokenFromCache(startURL, region, profileName)
	if err == nil {
		return token, nil
	}
	if refreshProfile == "" {
		return "", err
	}
	if err := tryRefreshingTokens(refreshProfile); err != nil {
		return "", err
	}
	return retrieveTokenFromCache(startURL, region, profileName)
}

// retrieveTokenFromCache scans ~/.aws/sso/cache for a matching, non-expired token.
func retrieveTokenFromCache(startURL, region, profileName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	cacheDir := filepath.Join(home, ".aws", "sso", "cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", fmt.Errorf("cannot read SSO cache directory: %w", err)
	}

	var fileErrors []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		token, err := retrieveTokenFromFile(filepath.Join(cacheDir, entry.Name()), startURL, region)
		if err != nil {
			fileErrors = append(fileErrors, fmt.Sprintf("%s: %v", entry.Name(), err))
			continue
		}
		if token != "" {
			return token, nil
		}
	}
	if len(fileErrors) > 0 {
		return "", fmt.Errorf("no valid SSO token found (%d cache file(s) unreadable: %s); please login with 'aws sso login --profile=%s'",
			len(fileErrors), strings.Join(fileErrors, "; "), profileName)
	}
	return "", fmt.Errorf("please login with 'aws sso login --profile=%s'", profileName)
}

// retrieveTokenFromFile checks a single SSO cache file for a valid token.
func retrieveTokenFromFile(filename, startURL, region string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var entry SSOCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", err
	}

	if entry.StartURL != startURL || entry.Region != region {
		return "", nil
	}

	expiresAt := entry.ExpiresAt
	if expiresAt == "" {
		return "", nil
	}

	expireTime, err := parseExpiresAt(expiresAt)
	if err != nil {
		return "", fmt.Errorf("cannot parse expiresAt %q: %w", expiresAt, err)
	}

	if expireTime.Before(time.Now().UTC()) {
		return "", nil // expired
	}

	return entry.AccessToken, nil
}

// parseExpiresAt handles the three formats AWS uses:
// "2020-03-26T13:28:35UTC", "2021-01-21T23:30:56Z", "2021-02-18T18:13:41.632177Z"
func parseExpiresAt(s string) (time.Time, error) {
	// Normalize: replace trailing "UTC" with "+0000", replace "Z" with "+0000"
	normalized := s
	if strings.HasSuffix(normalized, "UTC") {
		normalized = strings.TrimSuffix(normalized, "UTC") + "+0000"
	} else if strings.HasSuffix(normalized, "Z") {
		normalized = strings.TrimSuffix(normalized, "Z") + "+0000"
	}

	// Try with microseconds first, then without
	formats := []string{
		"2006-01-02T15:04:05.999999+0000",
		"2006-01-02T15:04:05+0000",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, normalized); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %s", s)
}

// tryRefreshingTokens refreshes the SSO token by making a quick STS call.
func tryRefreshingTokens(profileName string) error {
	_, err := runAWSCLI([]string{"sts", "get-caller-identity"}, profileName, withProfile())
	return err
}
