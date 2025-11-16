package repo

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestRepoOthersCmd_CommandSetup はrepo-othersコマンドの設定をテストします
func TestRepoOthersCmd_CommandSetup(t *testing.T) {
	if repoOthersCmd.Use != "repo-others" {
		t.Errorf("repoOthersCmd.Use = %q, want %q", repoOthersCmd.Use, "repo-others")
	}

	if repoOthersCmd.Short == "" {
		t.Error("repoOthersCmd.Short should not be empty")
	}

	if repoOthersCmd.Long == "" {
		t.Error("repoOthersCmd.Long should not be empty")
	}

	if repoOthersCmd.Example == "" {
		t.Error("repoOthersCmd.Example should not be empty")
	}
}

// TestRepoOthersCmd_InRootCmd はrepo-othersコマンドがRootCmdに登録されていることを確認します
func TestRepoOthersCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "repo-others" {
			found = true
			break
		}
	}

	if !found {
		t.Error("repoOthersCmd should be registered in RootCmd")
	}
}

// TestRepoOthersCmd_HasRunE はRunE関数が設定されていることを確認します
func TestRepoOthersCmd_HasRunE(t *testing.T) {
	if repoOthersCmd.RunE == nil {
		t.Error("repoOthersCmd.RunE should not be nil")
	}
}
