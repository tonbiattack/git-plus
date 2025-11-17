package issue

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestIssueCreateCmd_CommandSetup はissue-createコマンドの設定をテストします
func TestIssueCreateCmd_CommandSetup(t *testing.T) {
	if issueCreateCmd.Use != "issue-create" {
		t.Errorf("issueCreateCmd.Use = %q, want %q", issueCreateCmd.Use, "issue-create")
	}

	if issueCreateCmd.Short == "" {
		t.Error("issueCreateCmd.Short should not be empty")
	}

	if issueCreateCmd.Long == "" {
		t.Error("issueCreateCmd.Long should not be empty")
	}

	if issueCreateCmd.Example == "" {
		t.Error("issueCreateCmd.Example should not be empty")
	}
}

// TestIssueCreateCmd_InRootCmd はissue-createコマンドがrootCmdに登録されていることを確認します
func TestIssueCreateCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "issue-create" {
			found = true
			break
		}
	}

	if !found {
		t.Error("issueCreateCmd should be registered in rootCmd")
	}
}

// TestIssueCreateCmd_HasRunE はRunE関数が設定されていることを確認します
func TestIssueCreateCmd_HasRunE(t *testing.T) {
	if issueCreateCmd.RunE == nil {
		t.Error("issueCreateCmd.RunE should not be nil")
	}
}

// TestExtractIssueNumber はextractIssueNumber関数をテストします
func TestExtractIssueNumber(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "standard GitHub issue URL",
			url:      "https://github.com/user/repo/issues/123",
			expected: "123",
		},
		{
			name:     "issue number with multiple digits",
			url:      "https://github.com/org/project/issues/45678",
			expected: "45678",
		},
		{
			name:     "issue number 1",
			url:      "https://github.com/test/test/issues/1",
			expected: "1",
		},
		{
			name:     "invalid URL without issue number",
			url:      "https://github.com/user/repo/pulls/123",
			expected: "",
		},
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "URL with trailing slash",
			url:      "https://github.com/user/repo/issues/123/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIssueNumber(tt.url)
			if result != tt.expected {
				t.Errorf("extractIssueNumber(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}
