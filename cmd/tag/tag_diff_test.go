package tag

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestTagDiffCmd_CommandSetup はtag-diffコマンドの設定をテストします
func TestTagDiffCmd_CommandSetup(t *testing.T) {
	if tagDiffCmd.Use != "tag-diff <古いタグ> <新しいタグ>" {
		t.Errorf("tagDiffCmd.Use = %q, want %q", tagDiffCmd.Use, "tag-diff <古いタグ> <新しいタグ>")
	}

	if tagDiffCmd.Short == "" {
		t.Error("tagDiffCmd.Short should not be empty")
	}

	if tagDiffCmd.Long == "" {
		t.Error("tagDiffCmd.Long should not be empty")
	}

	if tagDiffCmd.Example == "" {
		t.Error("tagDiffCmd.Example should not be empty")
	}
}

// TestTagDiffCmd_Args は引数の検証をテストします
func TestTagDiffCmd_Args(t *testing.T) {
	// ExactArgs(2)が設定されているはず
	if tagDiffCmd.Args == nil {
		t.Error("tagDiffCmd.Args should not be nil")
	}
}

// TestVerifyTagExists はタグ存在確認をテストします
func TestVerifyTagExists(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")
	repo.CreateLightweightTag("v1.0.0")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 存在するタグ
	err = verifyTagExists("v1.0.0")
	if err != nil {
		t.Errorf("verifyTagExists should return nil for existing tag, got %v", err)
	}

	// 存在しないタグ
	err = verifyTagExists("nonexistent-tag")
	if err == nil {
		t.Error("verifyTagExists should return error for nonexistent tag")
	}
}

// TestTagDiffCmd_InRootCmd はtag-diffコマンドがRootCmdに登録されていることを確認します
func TestTagDiffCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "tag-diff <古いタグ> <新しいタグ>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("tagDiffCmd should be registered in RootCmd")
	}
}
