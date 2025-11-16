package stash

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestResumeCmd_CommandSetup はresumeコマンドの設定をテストします
func TestResumeCmd_CommandSetup(t *testing.T) {
	if resumeCmd.Use != "resume" {
		t.Errorf("resumeCmd.Use = %q, want %q", resumeCmd.Use, "resume")
	}

	if resumeCmd.Short == "" {
		t.Error("resumeCmd.Short should not be empty")
	}

	if resumeCmd.Long == "" {
		t.Error("resumeCmd.Long should not be empty")
	}

	if resumeCmd.Example == "" {
		t.Error("resumeCmd.Example should not be empty")
	}
}

// TestResumeCmd_InRootCmd はresumeコマンドがrootCmdに登録されていることを確認します
func TestResumeCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "resume" {
			found = true
			break
		}
	}

	if !found {
		t.Error("resumeCmd should be registered in rootCmd")
	}
}

// TestResumeCmd_HasRunE はRunE関数が設定されていることを確認します
func TestResumeCmd_HasRunE(t *testing.T) {
	if resumeCmd.RunE == nil {
		t.Error("resumeCmd.RunE should not be nil")
	}
}

// TestGetCurrentBranchName は現在のブランチ名の取得をテストします
func TestGetCurrentBranchName(t *testing.T) {
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

	branch, err := getCurrentBranchName()
	if err != nil {
		t.Errorf("getCurrentBranchName returned error: %v", err)
	}

	expectedBranch := repo.CurrentBranch()
	if branch != expectedBranch {
		t.Errorf("getCurrentBranchName() = %q, want %q", branch, expectedBranch)
	}
}

// TestSwitchBranchTo はブランチ切り替えをテストします
func TestSwitchBranchTo(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 別のブランチを作成
	repo.CreateBranch("feature/resume")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = switchBranchTo("feature/resume")
	if err != nil {
		t.Errorf("switchBranchTo returned error: %v", err)
	}

	currentBranch := repo.CurrentBranch()
	if currentBranch != "feature/resume" {
		t.Errorf("Current branch = %q, want %q", currentBranch, "feature/resume")
	}
}

// TestSwitchBranchTo_NonexistentBranch は存在しないブランチへの切り替えをテストします
func TestSwitchBranchTo_NonexistentBranch(t *testing.T) {
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

	err = switchBranchTo("nonexistent-branch")
	if err == nil {
		t.Error("switchBranchTo should return error for nonexistent branch")
	}
}
