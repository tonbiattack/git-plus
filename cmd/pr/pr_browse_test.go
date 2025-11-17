package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrBrowseCmd_CommandSetup はpr-browseコマンドの設定をテストします
func TestPrBrowseCmd_CommandSetup(t *testing.T) {
	if prBrowseCmd.Use != "pr-browse [PR番号]" {
		t.Errorf("prBrowseCmd.Use = %q, want %q", prBrowseCmd.Use, "pr-browse [PR番号]")
	}

	if prBrowseCmd.Short == "" {
		t.Error("prBrowseCmd.Short should not be empty")
	}

	if prBrowseCmd.Long == "" {
		t.Error("prBrowseCmd.Long should not be empty")
	}

	if prBrowseCmd.Example == "" {
		t.Error("prBrowseCmd.Example should not be empty")
	}
}

// TestPrBrowseCmd_InRootCmd はpr-browseコマンドがRootCmdに登録されていることを確認します
func TestPrBrowseCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-browse [PR番号]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prBrowseCmd should be registered in RootCmd")
	}
}

// TestPrBrowseCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrBrowseCmd_HasRunE(t *testing.T) {
	if prBrowseCmd.RunE == nil {
		t.Error("prBrowseCmd.RunE should not be nil")
	}
}
