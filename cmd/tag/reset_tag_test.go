package tag

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestResetTagCmd_CommandSetup はreset-tagコマンドの設定をテストします
func TestResetTagCmd_CommandSetup(t *testing.T) {
	if resetTagCmd.Use != "reset-tag <タグ名>" {
		t.Errorf("resetTagCmd.Use = %q, want %q", resetTagCmd.Use, "reset-tag <タグ名>")
	}

	if resetTagCmd.Short == "" {
		t.Error("resetTagCmd.Short should not be empty")
	}

	if resetTagCmd.Long == "" {
		t.Error("resetTagCmd.Long should not be empty")
	}

	if resetTagCmd.Example == "" {
		t.Error("resetTagCmd.Example should not be empty")
	}
}

// TestResetTagCmd_Args は引数の検証をテストします
func TestResetTagCmd_Args(t *testing.T) {
	// ExactArgs(1)が設定されているはず
	if resetTagCmd.Args == nil {
		t.Error("resetTagCmd.Args should not be nil")
	}
}

// TestResetTagCmd_InRootCmd はreset-tagコマンドがRootCmdに登録されていることを確認します
func TestResetTagCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "reset-tag <タグ名>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("resetTagCmd should be registered in RootCmd")
	}
}

// TestResetTagCmd_HasRunE はRunE関数が設定されていることを確認します
func TestResetTagCmd_HasRunE(t *testing.T) {
	if resetTagCmd.RunE == nil {
		t.Error("resetTagCmd.RunE should not be nil")
	}
}

// TestRunGitCommandIgnoreError はエラーを無視する関数をテストします
func TestRunGitCommandIgnoreError(t *testing.T) {
	// この関数は常にnilを返すはず
	err := runGitCommandIgnoreError("invalid-command-that-does-not-exist")
	if err != nil {
		t.Errorf("runGitCommandIgnoreError should return nil, got %v", err)
	}
}
