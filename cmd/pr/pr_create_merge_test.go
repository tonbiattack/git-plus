package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrCreateMergeCmd_CommandSetup はpr-create-mergeコマンドの設定をテストします
func TestPrCreateMergeCmd_CommandSetup(t *testing.T) {
	if prCreateMergeCmd.Use != "pr-create-merge [ベースブランチ名]" {
		t.Errorf("prCreateMergeCmd.Use = %q, want %q", prCreateMergeCmd.Use, "pr-create-merge")
	}

	if prCreateMergeCmd.Short == "" {
		t.Error("prCreateMergeCmd.Short should not be empty")
	}

	if prCreateMergeCmd.Long == "" {
		t.Error("prCreateMergeCmd.Long should not be empty")
	}

	if prCreateMergeCmd.Example == "" {
		t.Error("prCreateMergeCmd.Example should not be empty")
	}
}

// TestPrCreateMergeCmd_InRootCmd はpr-create-mergeコマンドがRootCmdに登録されていることを確認します
func TestPrCreateMergeCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-create-merge [ベースブランチ名]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prCreateMergeCmd should be registered in RootCmd")
	}
}

// TestPrCreateMergeCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrCreateMergeCmd_HasRunE(t *testing.T) {
	if prCreateMergeCmd.RunE == nil {
		t.Error("prCreateMergeCmd.RunE should not be nil")
	}
}
