package branch

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestDeleteLocalBranchesCmdDefinition はdelete-local-branchesコマンドの定義をテストします
func TestDeleteLocalBranchesCmdDefinition(t *testing.T) {
	if deleteLocalBranchesCmd == nil {
		t.Fatal("deleteLocalBranchesCmd should not be nil")
	}

	if deleteLocalBranchesCmd.Use != "delete-local-branches" {
		t.Errorf("deleteLocalBranchesCmd.Use = %q, want %q", deleteLocalBranchesCmd.Use, "delete-local-branches")
	}

	if deleteLocalBranchesCmd.Short == "" {
		t.Error("deleteLocalBranchesCmd.Short should not be empty")
	}

	if deleteLocalBranchesCmd.Long == "" {
		t.Error("deleteLocalBranchesCmd.Long should not be empty")
	}

	if deleteLocalBranchesCmd.RunE == nil {
		t.Error("deleteLocalBranchesCmd.RunE should not be nil")
	}
}

// TestShouldSkipProtectedBranch はshouldSkipProtectedBranch関数をテストします
func TestShouldSkipProtectedBranch(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		{"main branch", "main", true},
		{"master branch", "master", true},
		{"develop branch", "develop", true},
		{"feature branch", "feature/test", false},
		{"bugfix branch", "bugfix/fix", false},
		{"release branch", "release/v1.0", false},
		{"hotfix branch", "hotfix/urgent", false},
		{"custom branch", "my-branch", false},
		{"dev branch (not develop)", "dev", false},
		{"mainline branch", "mainline", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipProtectedBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("shouldSkipProtectedBranch(%q) = %v, want %v", tt.branch, result, tt.expected)
			}
		})
	}
}

// TestGetMergedBranches はマージ済みブランチの取得をテストします
func TestGetMergedBranches(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミット
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// マージ済みブランチを作成
	repo.CreateAndCheckoutBranch("feature/merged")
	repo.CreateFile("feature.txt", "feature")
	repo.Commit("Feature commit")

	// mainにマージ（現在のデフォルトブランチを取得）
	currentBranch := repo.CurrentBranch()
	if currentBranch != "feature/merged" {
		t.Fatalf("Expected to be on feature/merged, got %s", currentBranch)
	}

	// 最初のブランチに戻る
	repo.MustGit("checkout", "-")
	repo.MustGit("merge", "feature/merged", "--no-ff", "-m", "Merge feature")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branches, err := getMergedBranches()
	if err != nil {
		t.Errorf("getMergedBranches returned error: %v", err)
	}

	// feature/merged がリストに含まれているはず
	found := false
	for _, b := range branches {
		if b == "feature/merged" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'feature/merged' in merged branches list, got %v", branches)
	}
}

// TestGetMergedBranches_ExcludesCurrentBranch は現在のブランチが除外されることをテストします
func TestGetMergedBranches_ExcludesCurrentBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

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

	branches, err := getMergedBranches()
	if err != nil {
		t.Errorf("getMergedBranches returned error: %v", err)
	}

	// 現在のブランチ（main）が含まれていないことを確認
	for _, b := range branches {
		if b == "main" {
			t.Error("Current branch (main) should be excluded from merged branches")
		}
	}
}

// TestGetMergedBranches_ExcludesProtectedBranches は保護ブランチが除外されることをテストします
func TestGetMergedBranches_ExcludesProtectedBranches(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 保護ブランチを作成（全てマージ済み状態で）
	repo.CreateBranch("develop")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branches, err := getMergedBranches()
	if err != nil {
		t.Errorf("getMergedBranches returned error: %v", err)
	}

	// 保護ブランチが含まれていないことを確認
	for _, b := range branches {
		if b == "main" || b == "master" || b == "develop" {
			t.Errorf("Protected branch %q should be excluded from merged branches", b)
		}
	}
}

// TestGetMergedBranches_NoMergedBranches はマージ済みブランチがない場合のテスト
func TestGetMergedBranches_NoMergedBranches(t *testing.T) {
	repo := testutil.NewGitRepo(t)

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

	branches, err := getMergedBranches()
	if err != nil {
		t.Errorf("getMergedBranches returned error: %v", err)
	}

	// マージ済みブランチがないはず
	if len(branches) != 0 {
		t.Errorf("Expected no merged branches, got %v", branches)
	}
}

// TestDeleteLocalBranchesCommandRegistration はdelete-local-branchesコマンドがrootCmdに登録されているかテストします
func TestDeleteLocalBranchesCommandRegistration(t *testing.T) {
	foundCmd, _, err := cmd.RootCmd.Find([]string{"delete-local-branches"})
	if err != nil {
		t.Errorf("delete-local-branches command not found: %v", err)
	}
	if foundCmd == nil {
		t.Error("delete-local-branches command is nil")
	}
	if foundCmd.Name() != "delete-local-branches" {
		t.Errorf("Command name = %q, want %q", foundCmd.Name(), "delete-local-branches")
	}
}

// TestShouldSkipProtectedBranch_CaseSensitive は大文字小文字の区別をテストします
func TestShouldSkipProtectedBranch_CaseSensitive(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected bool
	}{
		{"uppercase MAIN", "MAIN", false},
		{"uppercase MASTER", "MASTER", false},
		{"uppercase DEVELOP", "DEVELOP", false},
		{"mixed case Main", "Main", false},
		{"mixed case Master", "Master", false},
		{"mixed case Develop", "Develop", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipProtectedBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("shouldSkipProtectedBranch(%q) = %v, want %v", tt.branch, result, tt.expected)
			}
		})
	}
}
