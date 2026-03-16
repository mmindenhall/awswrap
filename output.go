package main

import (
	"fmt"
	"os"
)

// exportCredentials prints credential export statements for the parent shell.
func exportCredentials(creds *RoleCredentials, profile *Profile) error {
	shellName, err := parentShellName()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not detect parent shell (%v), assuming POSIX syntax\n", err)
		shellName = "sh"
	}

	if isPowerShell(shellName) {
		fmt.Printf("$ENV:AWS_ACCESS_KEY_ID=\"%s\"\n", creds.AccessKeyID)
		fmt.Printf("$ENV:AWS_SECRET_ACCESS_KEY=\"%s\"\n", creds.SecretAccessKey)
		fmt.Printf("$ENV:AWS_SESSION_TOKEN=\"%s\"\n", creds.SessionToken)
		if os.Getenv("AWS_DEFAULT_REGION") == "" {
			if region, err := profile.getAttribute("region"); err == nil {
				fmt.Printf("$ENV:AWS_DEFAULT_REGION=\"%s\"\n", region)
			}
		}
	} else {
		fmt.Printf("export AWS_ACCESS_KEY_ID='%s'\n", creds.AccessKeyID)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY='%s'\n", creds.SecretAccessKey)
		fmt.Printf("export AWS_SESSION_TOKEN='%s'\n", creds.SessionToken)
		if os.Getenv("AWS_DEFAULT_REGION") == "" {
			if region, err := profile.getAttribute("region"); err == nil {
				fmt.Printf("export AWS_DEFAULT_REGION='%s'\n", region)
			}
		}
	}
	return nil
}
