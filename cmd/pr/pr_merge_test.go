package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrMergeCmd_CommandSetup はpr-mergeコマンドの設定をテストします
func TestPrMergeCmd_CommandSetup(t *testing.T) {
	if prMergeCmd.Use != "pr-merge [PR番号] [オプション...]" {
		t.Errorf("prMergeCmd.Use = %q, want %q", prMergeCmd.Use, "pr-merge")
	}

	if prMergeCmd.Short == "" {
		t.Error("prMergeCmd.Short should not be empty")
	}

	if prMergeCmd.Long == "" {
		t.Error("prMergeCmd.Long should not be empty")
	}

	if prMergeCmd.Example == "" {
		t.Error("prMergeCmd.Example should not be empty")
	}
}

// TestPrMergeCmd_InRootCmd はpr-mergeコマンドがRootCmdに登録されていることを確認します
func TestPrMergeCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-merge [PR番号] [オプション...]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prMergeCmd should be registered in RootCmd")
	}
}

// TestPrMergeCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrMergeCmd_HasRunE(t *testing.T) {
	if prMergeCmd.RunE == nil {
		t.Error("prMergeCmd.RunE should not be nil")
	}
}
