package repo

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestCloneOrgCmd_CommandSetup はclone-orgコマンドの設定をテストします
func TestCloneOrgCmd_CommandSetup(t *testing.T) {
	if cloneOrgCmd.Use != "clone-org <organization>" {
		t.Errorf("cloneOrgCmd.Use = %q, want %q", cloneOrgCmd.Use, "clone-org <organization>")
	}

	if cloneOrgCmd.Short == "" {
		t.Error("cloneOrgCmd.Short should not be empty")
	}

	if cloneOrgCmd.Long == "" {
		t.Error("cloneOrgCmd.Long should not be empty")
	}

	if cloneOrgCmd.Example == "" {
		t.Error("cloneOrgCmd.Example should not be empty")
	}
}

// TestCloneOrgCmd_Args は引数の検証をテストします
func TestCloneOrgCmd_Args(t *testing.T) {
	if cloneOrgCmd.Args == nil {
		t.Error("cloneOrgCmd.Args should not be nil")
	}
}

// TestCloneOrgCmd_InRootCmd はclone-orgコマンドがRootCmdに登録されていることを確認します
func TestCloneOrgCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "clone-org <organization>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("cloneOrgCmd should be registered in RootCmd")
	}
}

// TestCloneOrgCmd_HasRunE はRunE関数が設定されていることを確認します
func TestCloneOrgCmd_HasRunE(t *testing.T) {
	if cloneOrgCmd.RunE == nil {
		t.Error("cloneOrgCmd.RunE should not be nil")
	}
}
