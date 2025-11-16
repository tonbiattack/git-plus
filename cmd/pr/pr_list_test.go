package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrListCmd_CommandSetup はpr-listコマンドの設定をテストします
func TestPrListCmd_CommandSetup(t *testing.T) {
	if prListCmd.Use != "pr-list [オプション...]" {
		t.Errorf("prListCmd.Use = %q, want %q", prListCmd.Use, "pr-list")
	}

	if prListCmd.Short == "" {
		t.Error("prListCmd.Short should not be empty")
	}

	if prListCmd.Long == "" {
		t.Error("prListCmd.Long should not be empty")
	}

	if prListCmd.Example == "" {
		t.Error("prListCmd.Example should not be empty")
	}
}

// TestPrListCmd_InRootCmd はpr-listコマンドがRootCmdに登録されていることを確認します
func TestPrListCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-list [オプション...]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prListCmd should be registered in RootCmd")
	}
}

// TestPrListCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrListCmd_HasRunE(t *testing.T) {
	if prListCmd.RunE == nil {
		t.Error("prListCmd.RunE should not be nil")
	}
}
