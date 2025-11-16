package stash

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestStashSelectCmd_CommandSetup はstash-selectコマンドの設定をテストします
func TestStashSelectCmd_CommandSetup(t *testing.T) {
	if stashSelectCmd.Use != "stash-select" {
		t.Errorf("stashSelectCmd.Use = %q, want %q", stashSelectCmd.Use, "stash-select")
	}

	if stashSelectCmd.Short == "" {
		t.Error("stashSelectCmd.Short should not be empty")
	}

	if stashSelectCmd.Long == "" {
		t.Error("stashSelectCmd.Long should not be empty")
	}

	if stashSelectCmd.Example == "" {
		t.Error("stashSelectCmd.Example should not be empty")
	}
}

// TestStashSelectCmd_InRootCmd はstash-selectコマンドがrootCmdに登録されていることを確認します
func TestStashSelectCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "stash-select" {
			found = true
			break
		}
	}

	if !found {
		t.Error("stashSelectCmd should be registered in rootCmd")
	}
}

// TestStashEntry_Fields はStashEntry構造体をテストします
func TestStashEntry_Fields(t *testing.T) {
	entry := StashEntry{
		Index:   0,
		Ref:     "stash@{0}",
		Message: "WIP on main: abc123 commit message",
		Files:   []string{"file1.txt", "file2.txt"},
		Branch:  "main",
	}

	if entry.Index != 0 {
		t.Errorf("StashEntry.Index = %d, want %d", entry.Index, 0)
	}

	if entry.Ref != "stash@{0}" {
		t.Errorf("StashEntry.Ref = %q, want %q", entry.Ref, "stash@{0}")
	}

	if entry.Message != "WIP on main: abc123 commit message" {
		t.Errorf("StashEntry.Message = %q, want %q", entry.Message, "WIP on main: abc123 commit message")
	}

	if len(entry.Files) != 2 {
		t.Errorf("StashEntry.Files length = %d, want %d", len(entry.Files), 2)
	}

	if entry.Branch != "main" {
		t.Errorf("StashEntry.Branch = %q, want %q", entry.Branch, "main")
	}
}

// TestExtractBranch はブランチ名の抽出をテストします
func TestExtractBranch(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"WIP on main: abc123 commit", "main"},
		{"WIP on feature/test: xyz789 add feature", "feature/test"},
		{"WIP on develop: 123456 fix bug", "develop"},
		{"On master: abc123 commit", "master"},
		{"On feature/login: xyz789 add login", "feature/login"},
		{"random message", "(unknown)"},
		{"", "(unknown)"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := extractBranch(tt.message)
			if result != tt.expected {
				t.Errorf("extractBranch(%q) = %q, want %q", tt.message, result, tt.expected)
			}
		})
	}
}
