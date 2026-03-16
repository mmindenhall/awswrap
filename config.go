package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// Profile represents a resolved AWS config profile.
type Profile struct {
	Name          string
	SourceProfile *Profile
	SSOSession    *Profile
	Raw           map[string]string
}

// getAttribute returns the value for a key, falling back to the SSOSession if present.
func (p *Profile) getAttribute(key string) (string, error) {
	if v, ok := p.Raw[key]; ok {
		return v, nil
	}
	if p.SSOSession != nil {
		if v, ok := p.SSOSession.Raw[key]; ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("%q not found in profile %q", key, p.Name)
}

// readAWSConfig reads and parses the AWS config file.
func readAWSConfig() (*ini.File, string, error) {
	configPath := os.Getenv("AWS_CONFIG_FILE")
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		configPath = filepath.Join(home, ".aws", "config")
	}
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("cannot resolve config path %q: %w", configPath, err)
	}
	configPath = absPath

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, configPath, fmt.Errorf("cannot read config file %s: %w", configPath, err)
	}
	return cfg, configPath, nil
}

// retrieveProfile resolves a profile by name from the AWS config, recursively
// resolving source_profile and sso_session references.
func retrieveProfile(profileName string, profileType string) (*Profile, error) {
	cfg, configPath, err := readAWSConfig()
	if err != nil {
		return nil, err
	}
	return resolveProfile(cfg, configPath, profileName, profileType, nil)
}

func resolveProfile(cfg *ini.File, configPath, profileName, profileType string, visited map[string]bool) (*Profile, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}
	key := profileType + ":" + profileName
	if visited[key] {
		return nil, fmt.Errorf("circular profile reference detected at %s %q in %s", profileType, profileName, configPath)
	}
	visited[key] = true
	// Try "profile <name>" first, then bare "<name>"
	sectionName := fmt.Sprintf("%s %s", profileType, profileName)
	section, err := cfg.GetSection(sectionName)
	if err != nil {
		section, err = cfg.GetSection(profileName)
		if err != nil {
			return nil, fmt.Errorf("cannot find %s %q in %s", profileType, profileName, configPath)
		}
	}

	raw := make(map[string]string)
	for _, key := range section.Keys() {
		raw[key.Name()] = key.String()
	}

	p := &Profile{
		Name: profileName,
		Raw:  raw,
	}

	// Recursively resolve source_profile
	if sp, ok := raw["source_profile"]; ok {
		p.SourceProfile, err = resolveProfile(cfg, configPath, sp, "profile", visited)
		if err != nil {
			return nil, err
		}
	}

	// Recursively resolve sso_session
	if ss, ok := raw["sso_session"]; ok {
		p.SSOSession, err = resolveProfile(cfg, configPath, ss, "sso-session", visited)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
