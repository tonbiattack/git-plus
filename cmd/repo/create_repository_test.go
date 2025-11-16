package repo

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestCreateRepositoryCmd_CommandSetup はcreate-repositoryコマンドの設定をテストします
func TestCreateRepositoryCmd_CommandSetup(t *testing.T) {
	if createRepositoryCmd.Use != "create-repository <リポジトリ名>" {
		t.Errorf("createRepositoryCmd.Use = %q, want %q", createRepositoryCmd.Use, "create-repository")
	}

	if createRepositoryCmd.Short == "" {
		t.Error("createRepositoryCmd.Short should not be empty")
	}

	if createRepositoryCmd.Long == "" {
		t.Error("createRepositoryCmd.Long should not be empty")
	}

	if createRepositoryCmd.Example == "" {
		t.Error("createRepositoryCmd.Example should not be empty")
	}
}

// TestCreateRepositoryCmd_InRootCmd はcreate-repositoryコマンドがRootCmdに登録されていることを確認します
func TestCreateRepositoryCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "create-repository <リポジトリ名>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("createRepositoryCmd should be registered in RootCmd")
	}
}

// TestCreateRepositoryCmd_HasRunE はRunE関数が設定されていることを確認します
func TestCreateRepositoryCmd_HasRunE(t *testing.T) {
	if createRepositoryCmd.RunE == nil {
		t.Error("createRepositoryCmd.RunE should not be nil")
	}
}
