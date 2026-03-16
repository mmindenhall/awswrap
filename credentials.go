package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// RoleCredentials holds the AWS credential values.
type RoleCredentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      string `json:"expiration"`
}

// getRoleCredentials fetches SSO role credentials for a profile.
func getRoleCredentials(profile *Profile, parentProfileName string) (*RoleCredentials, error) {
	profileName := profile.Name
	startURL, err := profile.getAttribute("sso_start_url")
	if err != nil {
		return nil, err
	}
	ssoRegion, err := profile.getAttribute("sso_region")
	if err != nil {
		return nil, err
	}
	accountID, err := profile.getAttribute("sso_account_id")
	if err != nil {
		return nil, err
	}
	roleName, err := profile.getAttribute("sso_role_name")
	if err != nil {
		return nil, err
	}

	refreshProfile := chooseRefreshableProfile(parentProfileName, profile)
	accessToken, err := retrieveToken(startURL, ssoRegion, profileName, refreshProfile)
	if err != nil {
		return nil, err
	}

	result, err := runAWSCLI([]string{
		"sso", "get-role-credentials",
		"--role-name", roleName,
		"--account-id", accountID,
		"--access-token", accessToken,
		"--region", ssoRegion,
	}, profileName)
	if err != nil {
		return nil, err
	}

	var output struct {
		RoleCredentials struct {
			AccessKeyID     string      `json:"accessKeyId"`
			SecretAccessKey string      `json:"secretAccessKey"`
			SessionToken    string      `json:"sessionToken"`
			Expiration      json.Number `json:"expiration"`
		} `json:"roleCredentials"`
	}
	if err := json.Unmarshal(result, &output); err != nil {
		return nil, fmt.Errorf("failed to parse role credentials: %w", err)
	}

	// Convert epoch-ms expiration to ISO 8601
	epochMs, err := output.RoleCredentials.Expiration.Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to parse expiration: %w", err)
	}
	expTime := time.Unix(epochMs/1000, (epochMs%1000)*int64(time.Millisecond)).UTC()

	return &RoleCredentials{
		AccessKeyID:     output.RoleCredentials.AccessKeyID,
		SecretAccessKey: output.RoleCredentials.SecretAccessKey,
		SessionToken:    output.RoleCredentials.SessionToken,
		Expiration:      expTime.Format(time.RFC3339),
	}, nil
}

// chooseRefreshableProfile determines which profile name to use for token refresh.
func chooseRefreshableProfile(parentProfileName string, profile *Profile) string {
	if profile.SSOSession == nil {
		return "" // not refreshable
	}
	if parentProfileName != "" {
		return parentProfileName
	}
	return profile.Name
}

// getAssumedRoleCredentials recursively resolves source_profile chains and assumes roles.
func getAssumedRoleCredentials(profile *Profile, parentProfileName string) (*RoleCredentials, error) {
	if profile.SourceProfile == nil {
		return getRoleCredentials(profile, parentProfileName)
	}

	// Get source credentials recursively
	sourceCreds, err := getAssumedRoleCredentials(profile.SourceProfile, profile.Name)
	if err != nil {
		return nil, err
	}

	// Build env with source credentials, replacing any existing values
	env := filterEnv(os.Environ(),
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN",
	)
	env = append(env,
		"AWS_ACCESS_KEY_ID="+sourceCreds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY="+sourceCreds.SecretAccessKey,
		"AWS_SESSION_TOKEN="+sourceCreds.SessionToken,
	)

	// Determine role_session_name
	roleSessionName, err := profile.getAttribute("role_session_name")
	if err != nil {
		roleSessionName = fmt.Sprintf("botocore-session-%d", time.Now().Unix())
	}

	roleARN, err := profile.getAttribute("role_arn")
	if err != nil {
		return nil, err
	}

	result, err := runAWSCLI(
		[]string{
			"sts", "assume-role",
			"--role-arn", roleARN,
			"--role-session-name", roleSessionName,
		},
		profile.Name,
		withEnv(env),
		withErrorMessage(fmt.Sprintf("failed to assume-role %q", roleARN)),
	)
	if err != nil {
		return nil, err
	}

	var output struct {
		Credentials struct {
			AccessKeyID     string `json:"AccessKeyId"`
			SecretAccessKey string `json:"SecretAccessKey"`
			SessionToken    string `json:"SessionToken"`
			Expiration      string `json:"Expiration"`
		} `json:"Credentials"`
	}
	if err := json.Unmarshal(result, &output); err != nil {
		return nil, fmt.Errorf("failed to parse assume-role response: %w", err)
	}

	return &RoleCredentials{
		AccessKeyID:     output.Credentials.AccessKeyID,
		SecretAccessKey: output.Credentials.SecretAccessKey,
		SessionToken:    output.Credentials.SessionToken,
		Expiration:      output.Credentials.Expiration,
	}, nil
}
