package main

import "testing"

func TestIsPowerShell(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"pwsh", true},
		{"pwsh.exe", true},
		{"powershell", true},
		{"powershell.exe", true},
		{"bash", false},
		{"zsh", false},
		{"fish", false},
		{"cmd.exe", false},
		{"", false},
		{"PWSH", true},
		{"PowerShell.exe", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPowerShell(tt.name); got != tt.expected {
				t.Errorf("isPowerShell(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}
