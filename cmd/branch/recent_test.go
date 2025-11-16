package branch

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestGetRecentBranchesList は最近のブランチリスト取得をテストします
func TestGetRecentBranchesList(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 複数のブランチを作成
	repo.CreateBranch("feature/test1")
	repo.CreateBranch("feature/test2")

	// ディレクトリを変更してテスト実行
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branches, err := getRecentBranchesList()
	if err != nil {
		t.Errorf("getRecentBranchesList returned error: %v", err)
	}

	if len(branches) == 0 {
		t.Error("getRecentBranchesList should return at least one branch")
	}

	// ブランチ名が取得できていることを確認
	branchNames := make(map[string]bool)
	for _, branch := range branches {
		branchNames[branch.Name] = true
		if branch.Name == "" {
			t.Error("Branch name should not be empty")
		}
		if branch.LastCommitAt == "" {
			t.Errorf("LastCommitAt should not be empty for branch %s", branch.Name)
		}
	}

	// 作成したブランチが含まれていることを確認
	// 現在のブランチ名を取得（main または master）
	currentBranch := repo.CurrentBranch()
	expectedBranches := []string{currentBranch, "feature/test1", "feature/test2"}
	for _, expected := range expectedBranches {
		if !branchNames[expected] {
			t.Errorf("Expected branch %q not found in list", expected)
		}
	}
}

// TestGetRecentBranchesList_EmptyRepo は空のリポジトリでのテスト
func TestGetRecentBranchesList_NoBranches(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// コミットなしでブランチもなし
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branches, err := getRecentBranchesList()
	if err != nil {
		t.Errorf("getRecentBranchesList returned error: %v", err)
	}

	// 空のリポジトリでは空のリストが返される
	if len(branches) != 0 {
		t.Errorf("Expected empty list for repo without commits, got %d branches", len(branches))
	}
}

// TestGetCurrentBranchNow は現在のブランチ取得をテストします
func TestGetCurrentBranchNow(t *testing.T) {
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

	branch, err := getCurrentBranchNow()
	if err != nil {
		t.Errorf("getCurrentBranchNow returned error: %v", err)
	}

	// testutil.NewGitRepoで作成されたデフォルトブランチと一致することを確認
	expectedBranch := repo.CurrentBranch()
	if branch != expectedBranch {
		t.Errorf("getCurrentBranchNow = %q, want %q", branch, expectedBranch)
	}
}

// TestGetCurrentBranchNow_AfterSwitch はブランチ切り替え後のテスト
func TestGetCurrentBranchNow_AfterSwitch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ブランチを作成して切り替え
	repo.CreateAndCheckoutBranch("feature/new")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branch, err := getCurrentBranchNow()
	if err != nil {
		t.Errorf("getCurrentBranchNow returned error: %v", err)
	}

	if branch != "feature/new" {
		t.Errorf("getCurrentBranchNow = %q, want %q", branch, "feature/new")
	}
}

// TestBranchInfo_Fields はBranchInfo構造体のフィールドをテストします
func TestBranchInfo_Fields(t *testing.T) {
	tests := []struct {
		name         string
		branchName   string
		lastCommitAt string
	}{
		{
			name:         "通常のブランチ",
			branchName:   "main",
			lastCommitAt: "2 hours ago",
		},
		{
			name:         "スラッシュ付きブランチ",
			branchName:   "feature/test-feature",
			lastCommitAt: "1 day ago",
		},
		{
			name:         "長い時間表記",
			branchName:   "develop",
			lastCommitAt: "3 weeks ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := BranchInfo{
				Name:         tt.branchName,
				LastCommitAt: tt.lastCommitAt,
			}

			if info.Name != tt.branchName {
				t.Errorf("BranchInfo.Name = %q, want %q", info.Name, tt.branchName)
			}

			if info.LastCommitAt != tt.lastCommitAt {
				t.Errorf("BranchInfo.LastCommitAt = %q, want %q", info.LastCommitAt, tt.lastCommitAt)
			}
		})
	}
}

// TestSwitchToSelectedBranch はブランチ切り替え関数をテストします
func TestSwitchToSelectedBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 切り替え先のブランチを作成
	repo.CreateBranch("feature/target")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// デフォルトブランチにいることを確認
	initialBranch := repo.CurrentBranch()
	current, _ := getCurrentBranchNow()
	if current != initialBranch {
		t.Fatalf("Should start on %s branch, got %s", initialBranch, current)
	}

	// ブランチを切り替え
	err = switchToSelectedBranch("feature/target")
	if err != nil {
		t.Errorf("switchToSelectedBranch returned error: %v", err)
	}

	// 切り替わったことを確認
	newBranch, _ := getCurrentBranchNow()
	if newBranch != "feature/target" {
		t.Errorf("Branch should be switched to feature/target, got %s", newBranch)
	}
}

// TestSwitchToSelectedBranch_NonexistentBranch は存在しないブランチへの切り替えをテスト
func TestSwitchToSelectedBranch_NonexistentBranch(t *testing.T) {
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

	// 存在しないブランチへの切り替えはエラーになるはず
	err = switchToSelectedBranch("nonexistent-branch")
	if err == nil {
		t.Error("switchToSelectedBranch should return error for nonexistent branch")
	}
}

// TestGetRecentBranchesList_ParsesCorrectly はパース処理が正しいことをテストします
func TestGetRecentBranchesList_ParsesCorrectly(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 複数のブランチを作成してコミット
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateAndCheckoutBranch("branch-a")
	repo.CreateFile("a.txt", "content a")
	repo.Commit("Commit on branch-a")

	// デフォルトブランチに戻る
	repo.MustGit("checkout", "-")
	repo.CreateAndCheckoutBranch("branch-b")
	repo.CreateFile("b.txt", "content b")
	repo.Commit("Commit on branch-b")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branches, err := getRecentBranchesList()
	if err != nil {
		t.Fatalf("getRecentBranchesList returned error: %v", err)
	}

	// 少なくとも3つのブランチがあるはず（main, branch-a, branch-b）
	if len(branches) < 3 {
		t.Errorf("Expected at least 3 branches, got %d", len(branches))
	}

	// 各ブランチの情報が正しく設定されているか確認
	for _, branch := range branches {
		if branch.Name == "" {
			t.Error("Branch name should not be empty")
		}
		if branch.LastCommitAt == "" {
			t.Errorf("LastCommitAt should not be empty for branch %s", branch.Name)
		}
	}
}
