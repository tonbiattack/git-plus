package pr

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestGetBranchCurrent_Success は getBranchCurrent が現在のブランチ名を正しく取得することをテストします
func TestGetBranchCurrent_Success(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branch, err := getBranchCurrent()
	if err != nil {
		t.Errorf("getBranchCurrent() returned error: %v", err)
	}

	// デフォルトブランチは master か main
	if branch != "master" && branch != "main" {
		t.Errorf("getBranchCurrent() = %q, want 'master' or 'main'", branch)
	}
}

// TestGetBranchCurrent_FeatureBranch はフィーチャーブランチでの動作をテストします
func TestGetBranchCurrent_FeatureBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// フィーチャーブランチを作成してチェックアウト
	repo.CreateAndCheckoutBranch("feature/test-branch")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branch, err := getBranchCurrent()
	if err != nil {
		t.Errorf("getBranchCurrent() returned error: %v", err)
	}

	if branch != "feature/test-branch" {
		t.Errorf("getBranchCurrent() = %q, want %q", branch, "feature/test-branch")
	}
}

// TestCheckUncommittedChanges_NoChanges は変更がない場合をテストします
func TestCheckUncommittedChanges_NoChanges(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges() returned error: %v", err)
	}

	if hasChanges {
		t.Error("checkUncommittedChanges() = true, want false (no changes)")
	}
}

// TestCheckUncommittedChanges_WithUnstagedChanges は未ステージの変更がある場合をテストします
func TestCheckUncommittedChanges_WithUnstagedChanges(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ファイルを変更（ステージしない）
	repo.CreateFile("README.md", "# Modified Test")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges() returned error: %v", err)
	}

	if !hasChanges {
		t.Error("checkUncommittedChanges() = false, want true (has unstaged changes)")
	}
}

// TestCheckUncommittedChanges_WithStagedChanges はステージ済みの変更がある場合をテストします
func TestCheckUncommittedChanges_WithStagedChanges(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 新しいファイルを作成してステージ
	repo.CreateFile("new_file.txt", "new content")
	repo.MustGit("add", "new_file.txt")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges() returned error: %v", err)
	}

	if !hasChanges {
		t.Error("checkUncommittedChanges() = false, want true (has staged changes)")
	}
}

// TestCheckUncommittedChanges_WithUntrackedFiles は未追跡ファイルがある場合をテストします
func TestCheckUncommittedChanges_WithUntrackedFiles(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 新しい未追跡ファイルを作成
	repo.CreateFile("untracked.txt", "untracked content")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	hasChanges, err := checkUncommittedChanges()
	if err != nil {
		t.Errorf("checkUncommittedChanges() returned error: %v", err)
	}

	if !hasChanges {
		t.Error("checkUncommittedChanges() = false, want true (has untracked files)")
	}
}

// TestCreateStashWithMessage_Success は stash が正常に作成されることをテストします
func TestCreateStashWithMessage_Success(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ファイルを変更
	repo.CreateFile("README.md", "# Modified Test")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	stashRef, err := createStashWithMessage("test stash message")
	if err != nil {
		t.Errorf("createStashWithMessage() returned error: %v", err)
	}

	// stash 参照が SHA-1 ハッシュ形式であることを確認
	if len(stashRef) != 40 {
		t.Errorf("createStashWithMessage() returned %q, want 40-character SHA-1 hash", stashRef)
	}

	// stash が作成されたことを確認
	output, err := repo.Git("stash", "list")
	if err != nil {
		t.Errorf("git stash list failed: %v", err)
	}

	if output == "" {
		t.Error("No stash was created")
	}
}

// TestCreateStashWithMessage_NoChanges は変更がない場合のエラーをテストします
func TestCreateStashWithMessage_NoChanges(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 変更がない状態で stash を作成しようとする
	_, err := createStashWithMessage("empty stash")
	if err == nil {
		t.Error("createStashWithMessage() should return error when no changes to stash")
	}
}
