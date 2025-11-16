package stash

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestPauseCmd_CommandSetup はpauseコマンドの設定をテストします
func TestPauseCmd_CommandSetup(t *testing.T) {
	if pauseCmd.Use != "pause <branch>" {
		t.Errorf("pauseCmd.Use = %q, want %q", pauseCmd.Use, "pause <branch>")
	}

	if pauseCmd.Short == "" {
		t.Error("pauseCmd.Short should not be empty")
	}

	if pauseCmd.Long == "" {
		t.Error("pauseCmd.Long should not be empty")
	}

	if pauseCmd.Example == "" {
		t.Error("pauseCmd.Example should not be empty")
	}
}

// TestPauseCmd_Args は引数の検証をテストします
func TestPauseCmd_Args(t *testing.T) {
	// ExactArgs(1)が設定されているはず
	if pauseCmd.Args == nil {
		t.Error("pauseCmd.Args should not be nil")
	}
}

// TestPauseCmd_InRootCmd はpauseコマンドがrootCmdに登録されていることを確認します
func TestPauseCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pause <branch>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("pauseCmd should be registered in rootCmd")
	}
}

// TestGetBranchCurrent は現在のブランチ名の取得をテストします
func TestGetBranchCurrent(t *testing.T) {
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

	branch, err := getBranchCurrent()
	if err != nil {
		t.Errorf("getBranchCurrent returned error: %v", err)
	}

	expectedBranch := repo.CurrentBranch()
	if branch != expectedBranch {
		t.Errorf("getBranchCurrent() = %q, want %q", branch, expectedBranch)
	}
}

// TestCheckUncommittedChanges_NoChanges は変更なしの状態をテストします
func TestCheckUncommittedChanges_NoChanges(t *testing.T) {
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

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges returned error: %v", err)
	}

	if hasChanges {
		t.Error("checkUncommittedChanges should return false when there are no changes")
	}
}

// TestCheckUncommittedChanges_WithChanges は変更ありの状態をテストします
func TestCheckUncommittedChanges_WithChanges(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 未コミットの変更を作成
	repo.CreateFile("new_file.txt", "uncommitted content")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges returned error: %v", err)
	}

	if !hasChanges {
		t.Error("checkUncommittedChanges should return true when there are changes")
	}
}

// TestCheckoutBranch はブランチのチェックアウトをテストします
func TestCheckoutBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 別のブランチを作成
	repo.CreateBranch("feature/test")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = checkoutBranch("feature/test")
	if err != nil {
		t.Errorf("checkoutBranch returned error: %v", err)
	}

	currentBranch := repo.CurrentBranch()
	if currentBranch != "feature/test" {
		t.Errorf("Current branch = %q, want %q", currentBranch, "feature/test")
	}
}

// TestCheckoutBranch_NonexistentBranch は存在しないブランチへのチェックアウトをテストします
func TestCheckoutBranch_NonexistentBranch(t *testing.T) {
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

	err = checkoutBranch("nonexistent-branch")
	if err == nil {
		t.Error("checkoutBranch should return error for nonexistent branch")
	}
}
