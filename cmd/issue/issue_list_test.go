package issue

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestIssueListCmd_CommandSetup はissue-listコマンドの設定をテストします
func TestIssueListCmd_CommandSetup(t *testing.T) {
	if issueListCmd.Use != "issue-list" {
		t.Errorf("issueListCmd.Use = %q, want %q", issueListCmd.Use, "issue-list")
	}

	if issueListCmd.Short == "" {
		t.Error("issueListCmd.Short should not be empty")
	}

	if issueListCmd.Long == "" {
		t.Error("issueListCmd.Long should not be empty")
	}

	if issueListCmd.Example == "" {
		t.Error("issueListCmd.Example should not be empty")
	}
}

// TestIssueListCmd_InRootCmd はissue-listコマンドがrootCmdに登録されていることを確認します
func TestIssueListCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "issue-list" {
			found = true
			break
		}
	}

	if !found {
		t.Error("issueListCmd should be registered in rootCmd")
	}
}

// TestIssueListCmd_HasRunE はRunE関数が設定されていることを確認します
func TestIssueListCmd_HasRunE(t *testing.T) {
	if issueListCmd.RunE == nil {
		t.Error("issueListCmd.RunE should not be nil")
	}
}

// TestDisplayIssueList はissue一覧の表示機能をテストします
func TestDisplayIssueList(t *testing.T) {
	tests := []struct {
		name   string
		issues []IssueEntry
	}{
		{
			name: "single issue",
			issues: []IssueEntry{
				{Number: 1, Title: "Test Issue", Body: "Test Body", URL: "https://github.com/test/test/issues/1"},
			},
		},
		{
			name: "multiple issues",
			issues: []IssueEntry{
				{Number: 1, Title: "First Issue", Body: "First Body", URL: "https://github.com/test/test/issues/1"},
				{Number: 2, Title: "Second Issue", Body: "Second Body", URL: "https://github.com/test/test/issues/2"},
			},
		},
		{
			name: "issue with long body",
			issues: []IssueEntry{
				{Number: 1, Title: "Long Body Issue", Body: "This is a very long body that exceeds fifty characters and should be truncated", URL: "https://github.com/test/test/issues/1"},
			},
		},
		{
			name: "issue with empty body",
			issues: []IssueEntry{
				{Number: 1, Title: "No Body Issue", Body: "", URL: "https://github.com/test/test/issues/1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// displayIssueList should not panic
			displayIssueList(tt.issues)
		})
	}
}

// TestDisplayIssueDetail はissue詳細の表示機能をテストします
func TestDisplayIssueDetail(t *testing.T) {
	tests := []struct {
		name  string
		issue *IssueEntry
	}{
		{
			name:  "issue with body",
			issue: &IssueEntry{Number: 1, Title: "Test Issue", Body: "Test Body", URL: "https://github.com/test/test/issues/1"},
		},
		{
			name:  "issue without body",
			issue: &IssueEntry{Number: 2, Title: "No Body", Body: "", URL: "https://github.com/test/test/issues/2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// displayIssueDetail should not panic
			displayIssueDetail(tt.issue)
		})
	}
}

// TestIssueListCmd_HasBulkCloseOption はissue-listコマンドに一括クローズオプションが含まれていることを確認します
func TestIssueListCmd_HasBulkCloseOption(t *testing.T) {
	// Longの説明に一括クローズが含まれていることを確認
	if !containsString(issueListCmd.Long, "一括クローズ") {
		t.Error("issueListCmd.Long should mention '一括クローズ' option")
	}
}

// containsString は文字列に部分文字列が含まれているかを確認するヘルパー関数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}

// TestPromptForBulkCloseComment_CreatesTempFile は一括クローズ用のコメントファイル作成をテストします
func TestPromptForBulkCloseComment_CreatesTempFile(t *testing.T) {
	issueNumbers := []int{1, 2, 3}
	tmpFile, err := createTempBulkCloseCommentFile(issueNumbers)
	if err != nil {
		t.Errorf("createTempBulkCloseCommentFile() error = %v", err)
		return
	}
	if tmpFile == "" {
		t.Error("createTempBulkCloseCommentFile() returned empty file path")
	}
}
