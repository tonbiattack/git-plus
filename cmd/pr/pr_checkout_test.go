package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrCheckoutCmd_CommandSetup はpr-checkoutコマンドの設定をテストします
func TestPrCheckoutCmd_CommandSetup(t *testing.T) {
	if prCheckoutCmd.Use != "pr-checkout [PR番号]" {
		t.Errorf("prCheckoutCmd.Use = %q, want %q", prCheckoutCmd.Use, "pr-checkout")
	}

	if prCheckoutCmd.Short == "" {
		t.Error("prCheckoutCmd.Short should not be empty")
	}

	if prCheckoutCmd.Long == "" {
		t.Error("prCheckoutCmd.Long should not be empty")
	}

	if prCheckoutCmd.Example == "" {
		t.Error("prCheckoutCmd.Example should not be empty")
	}
}

// TestPrCheckoutCmd_InRootCmd はpr-checkoutコマンドがRootCmdに登録されていることを確認します
func TestPrCheckoutCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-checkout [PR番号]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prCheckoutCmd should be registered in RootCmd")
	}
}

// TestPrCheckoutCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrCheckoutCmd_HasRunE(t *testing.T) {
	if prCheckoutCmd.RunE == nil {
		t.Error("prCheckoutCmd.RunE should not be nil")
	}
}
