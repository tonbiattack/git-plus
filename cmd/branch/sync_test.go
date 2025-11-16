package branch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestSyncCmdDefinition はsyncコマンドの定義をテストします
func TestSyncCmdDefinition(t *testing.T) {
	if syncCmd == nil {
		t.Fatal("syncCmd should not be nil")
	}

	if syncCmd.Use != "sync [ブランチ名]" {
		t.Errorf("syncCmd.Use = %q, want %q", syncCmd.Use, "sync [ブランチ名]")
	}

	if syncCmd.Short == "" {
		t.Error("syncCmd.Short should not be empty")
	}

	if syncCmd.Long == "" {
		t.Error("syncCmd.Long should not be empty")
	}

	if syncCmd.Example == "" {
		t.Error("syncCmd.Example should not be empty")
	}

	if syncCmd.RunE == nil {
		t.Error("syncCmd.RunE should not be nil")
	}
}

// TestSyncCmdFlags はsyncコマンドのフラグをテストします
func TestSyncCmdFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"continue flag", "continue", "c"},
		{"abort flag", "abort", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := syncCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %q not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// TestCheckRebaseInProgress_NotInProgress はリベース中でない場合のテスト
func TestCheckRebaseInProgress_NotInProgress(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	result := checkRebaseInProgress()
	if result {
		t.Error("checkRebaseInProgress should return false when not rebasing")
	}
}

// TestCheckRebaseInProgress_WithRebaseMerge はrebase-mergeディレクトリがある場合のテスト
func TestCheckRebaseInProgress_WithRebaseMerge(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// .git/rebase-merge ディレクトリを作成
	rebaseMergeDir := filepath.Join(repo.Dir, ".git", "rebase-merge")
	if err := os.MkdirAll(rebaseMergeDir, 0755); err != nil {
		t.Fatalf("Failed to create rebase-merge directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(rebaseMergeDir) }()

	result := checkRebaseInProgress()
	if !result {
		t.Error("checkRebaseInProgress should return true when rebase-merge exists")
	}
}

// TestCheckRebaseInProgress_WithRebaseApply はrebase-applyディレクトリがある場合のテスト
func TestCheckRebaseInProgress_WithRebaseApply(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// .git/rebase-apply ディレクトリを作成
	rebaseApplyDir := filepath.Join(repo.Dir, ".git", "rebase-apply")
	if err := os.MkdirAll(rebaseApplyDir, 0755); err != nil {
		t.Fatalf("Failed to create rebase-apply directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(rebaseApplyDir) }()

	result := checkRebaseInProgress()
	if !result {
		t.Error("checkRebaseInProgress should return true when rebase-apply exists")
	}
}

// TestDetectDefaultRemoteBranch_NoRemote はリモートがない場合のテスト
func TestDetectDefaultRemoteBranch_NoRemote(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// リモートがない場合はエラーになるはず
	_, err = detectDefaultRemoteBranch()
	if err == nil {
		t.Error("detectDefaultRemoteBranch should return error when no remote exists")
	}
}

// TestSyncCmdFlagValues はフラグのデフォルト値をテストします
func TestSyncCmdFlagValues(t *testing.T) {
	// フラグのデフォルト値を確認
	continueFlag := syncCmd.Flags().Lookup("continue")
	if continueFlag == nil {
		t.Fatal("continue flag not found")
	}
	if continueFlag.DefValue != "false" {
		t.Errorf("continue flag default value = %q, want %q", continueFlag.DefValue, "false")
	}

	abortFlag := syncCmd.Flags().Lookup("abort")
	if abortFlag == nil {
		t.Fatal("abort flag not found")
	}
	if abortFlag.DefValue != "false" {
		t.Errorf("abort flag default value = %q, want %q", abortFlag.DefValue, "false")
	}
}

// TestSyncCommandRegistration はsyncコマンドがrootCmdに登録されているかテストします
func TestSyncCommandRegistration(t *testing.T) {
	foundCmd, _, err := cmd.RootCmd.Find([]string{"sync"})
	if err != nil {
		t.Errorf("sync command not found: %v", err)
	}
	if foundCmd == nil {
		t.Error("sync command is nil")
	}
	if foundCmd.Name() != "sync" {
		t.Errorf("Command name = %q, want %q", foundCmd.Name(), "sync")
	}
}
