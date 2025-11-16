package issue

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestIssueEditCmd_CommandSetup はissue-editコマンドの設定をテストします
func TestIssueEditCmd_CommandSetup(t *testing.T) {
	if issueEditCmd.Use != "issue-edit" {
		t.Errorf("issueEditCmd.Use = %q, want %q", issueEditCmd.Use, "issue-edit")
	}

	if issueEditCmd.Short == "" {
		t.Error("issueEditCmd.Short should not be empty")
	}

	if issueEditCmd.Long == "" {
		t.Error("issueEditCmd.Long should not be empty")
	}

	if issueEditCmd.Example == "" {
		t.Error("issueEditCmd.Example should not be empty")
	}
}

// TestIssueEditCmd_InRootCmd はissue-editコマンドがrootCmdに登録されていることを確認します
func TestIssueEditCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "issue-edit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("issueEditCmd should be registered in rootCmd")
	}
}

// TestIssueEditCmd_HasRunE はRunE関数が設定されていることを確認します
func TestIssueEditCmd_HasRunE(t *testing.T) {
	if issueEditCmd.RunE == nil {
		t.Error("issueEditCmd.RunE should not be nil")
	}
}
