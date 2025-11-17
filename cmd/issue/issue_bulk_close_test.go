package issue

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestIssueBulkCloseCmd_CommandSetup はissue-bulk-closeコマンドの設定をテストします
func TestIssueBulkCloseCmd_CommandSetup(t *testing.T) {
	if issueBulkCloseCmd.Use != "issue-bulk-close <ISSUE番号...>" {
		t.Errorf("issueBulkCloseCmd.Use = %q, want %q", issueBulkCloseCmd.Use, "issue-bulk-close <ISSUE番号...>")
	}

	if issueBulkCloseCmd.Short == "" {
		t.Error("issueBulkCloseCmd.Short should not be empty")
	}

	if issueBulkCloseCmd.Long == "" {
		t.Error("issueBulkCloseCmd.Long should not be empty")
	}

	if issueBulkCloseCmd.Example == "" {
		t.Error("issueBulkCloseCmd.Example should not be empty")
	}
}

// TestIssueBulkCloseCmd_InRootCmd はissue-bulk-closeコマンドがrootCmdに登録されていることを確認します
func TestIssueBulkCloseCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "issue-bulk-close <ISSUE番号...>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("issueBulkCloseCmd should be registered in rootCmd")
	}
}

// TestIssueBulkCloseCmd_HasRunE はRunE関数が設定されていることを確認します
func TestIssueBulkCloseCmd_HasRunE(t *testing.T) {
	if issueBulkCloseCmd.RunE == nil {
		t.Error("issueBulkCloseCmd.RunE should not be nil")
	}
}

// TestIssueBulkCloseCmd_HasMessageFlag はmessageフラグが設定されていることを確認します
func TestIssueBulkCloseCmd_HasMessageFlag(t *testing.T) {
	flag := issueBulkCloseCmd.Flags().Lookup("message")
	if flag == nil {
		t.Fatal("issueBulkCloseCmd should have 'message' flag")
	}

	if flag.Shorthand != "m" {
		t.Errorf("message flag shorthand = %q, want %q", flag.Shorthand, "m")
	}

	if flag.DefValue != "" {
		t.Errorf("message flag default value = %q, want %q", flag.DefValue, "")
	}
}

// TestIssueBulkCloseCmd_RequiresArgs はissue-bulk-closeが引数を必要とすることを確認します
func TestIssueBulkCloseCmd_RequiresArgs(t *testing.T) {
	if issueBulkCloseCmd.Args == nil {
		t.Error("issueBulkCloseCmd.Args should not be nil (should require minimum 1 argument)")
	}
}

// TestCreateTempBulkCloseCommentFile は一括クローズ用の一時ファイル作成をテストします
func TestCreateTempBulkCloseCommentFile(t *testing.T) {
	tests := []struct {
		name         string
		issueNumbers []int
	}{
		{
			name:         "single issue",
			issueNumbers: []int{1},
		},
		{
			name:         "multiple issues",
			issueNumbers: []int{1, 2, 3},
		},
		{
			name:         "many issues",
			issueNumbers: []int{1, 5, 10, 20, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// createTempBulkCloseCommentFile should not panic and should create a file
			tmpFile, err := createTempBulkCloseCommentFile(tt.issueNumbers)
			if err != nil {
				t.Errorf("createTempBulkCloseCommentFile() error = %v", err)
				return
			}
			if tmpFile == "" {
				t.Error("createTempBulkCloseCommentFile() returned empty file path")
			}
			// Cleanup
			// Note: In actual implementation, the file should be removed after use
		})
	}
}
