package repo

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestBrowseCmd_CommandSetup はbrowseコマンドの設定をテストします
func TestBrowseCmd_CommandSetup(t *testing.T) {
	if browseCmd.Use != "browse" {
		t.Errorf("browseCmd.Use = %q, want %q", browseCmd.Use, "browse")
	}

	if browseCmd.Short == "" {
		t.Error("browseCmd.Short should not be empty")
	}

	if browseCmd.Long == "" {
		t.Error("browseCmd.Long should not be empty")
	}
}

// TestBrowseCmd_InRootCmd はbrowseコマンドがRootCmdに登録されていることを確認します
func TestBrowseCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "browse" {
			found = true
			break
		}
	}

	if !found {
		t.Error("browseCmd should be registered in RootCmd")
	}
}

// TestBrowseCmd_HasRunE はRunE関数が設定されていることを確認します
func TestBrowseCmd_HasRunE(t *testing.T) {
	if browseCmd.RunE == nil {
		t.Error("browseCmd.RunE should not be nil")
	}
}

// TestBrowseCmd_LongDescription は詳細説明が要件を含んでいることを確認します
func TestBrowseCmd_LongDescription(t *testing.T) {
	longDesc := browseCmd.Long

	// GitHub CLI の要件が記載されているか確認
	requiredInfo := []string{
		"GitHub CLI",
		"gh",
		"auth",
		"login",
	}

	for _, info := range requiredInfo {
		found := false
		for i := 0; i <= len(longDesc)-len(info); i++ {
			if longDesc[i:i+len(info)] == info {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("browseCmd.Long should contain %q", info)
		}
	}
}
