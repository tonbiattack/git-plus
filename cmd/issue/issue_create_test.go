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
