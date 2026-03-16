package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// parentShellName returns the name of the parent process (the invoking shell).
func parentShellName() (string, error) {
	ppid := os.Getppid()
	switch runtime.GOOS {
	case "windows":
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", ppid), "/FO", "CSV", "/NH").Output()
		if err != nil {
			return "", fmt.Errorf("cannot determine parent shell: %w", err)
		}
		// tasklist CSV output: "name.exe","pid",...
		line := strings.TrimSpace(string(out))
		if parts := strings.SplitN(line, ",", 2); len(parts) > 0 {
			return strings.Trim(parts[0], "\""), nil
		}
		return "", fmt.Errorf("cannot parse tasklist output")
	default:
		out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=").Output()
		if err != nil {
			return "", fmt.Errorf("cannot determine parent shell: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}
}

// isPowerShell returns true if the given process name is a PowerShell variant.
func isPowerShell(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "pwsh", "pwsh.exe", "powershell", "powershell.exe":
		return true
	}
	return false
}
