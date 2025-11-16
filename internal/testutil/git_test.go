package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGitRepo(t *testing.T) {
	repo := NewGitRepo(t)

	// ディレクトリが存在することを確認
	if _, err := os.Stat(repo.Dir); err != nil {
		t.Errorf("Repository directory should exist: %v", err)
	}

	// .git ディレクトリが存在することを確認
	gitDir := filepath.Join(repo.Dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		t.Errorf(".git directory should exist: %v", err)
	}
}

func TestGitRepo_Git(t *testing.T) {
	repo := NewGitRepo(t)

	output, err := repo.Git("--version")
	if err != nil {
		t.Errorf("Git --version should succeed: %v", err)
	}

	if !strings.HasPrefix(output, "git version") {
		t.Errorf("Expected git version output, got: %s", output)
	}
}

func TestGitRepo_CreateFile(t *testing.T) {
	repo := NewGitRepo(t)

	// 単純なファイル作成
	repo.CreateFile("test.txt", "Hello, World!")

	content := repo.ReadFile("test.txt")
	if content != "Hello, World!" {
		t.Errorf("File content mismatch: got %q, want %q", content, "Hello, World!")
	}
}

func TestGitRepo_CreateFileWithSubdirectory(t *testing.T) {
	repo := NewGitRepo(t)

	// サブディレクトリを含むファイル作成
	repo.CreateFile("src/main.go", "package main")

	if !repo.FileExists("src/main.go") {
		t.Error("File src/main.go should exist")
	}

	content := repo.ReadFile("src/main.go")
	if content != "package main" {
		t.Errorf("File content mismatch: got %q", content)
	}
}

func TestGitRepo_Commit(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// コミットが作成されたことを確認
	count := repo.CommitCount()
	if count != 1 {
		t.Errorf("Expected 1 commit, got %d", count)
	}
}

func TestGitRepo_CreateBranch(t *testing.T) {
	repo := NewGitRepo(t)

	// 初期コミットが必要
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateBranch("feature")

	// ブランチが作成されたことを確認
	output, _ := repo.Git("branch")
	if !strings.Contains(output, "feature") {
		t.Error("Branch 'feature' should be created")
	}
}

func TestGitRepo_CheckoutBranch(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateBranch("feature")
	repo.CheckoutBranch("feature")

	current := repo.CurrentBranch()
	if current != "feature" {
		t.Errorf("Current branch should be 'feature', got %q", current)
	}
}

func TestGitRepo_CreateAndCheckoutBranch(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateAndCheckoutBranch("develop")

	current := repo.CurrentBranch()
	if current != "develop" {
		t.Errorf("Current branch should be 'develop', got %q", current)
	}
}

func TestGitRepo_CurrentBranch(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// デフォルトブランチを確認
	current := repo.CurrentBranch()
	if current == "" {
		t.Error("Current branch should not be empty")
	}
}

func TestGitRepo_CreateTag(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateTag("v1.0.0", "Version 1.0.0")

	output, _ := repo.Git("tag")
	if !strings.Contains(output, "v1.0.0") {
		t.Error("Tag 'v1.0.0' should be created")
	}
}

func TestGitRepo_CreateLightweightTag(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	repo.CreateLightweightTag("v0.1.0")

	output, _ := repo.Git("tag")
	if !strings.Contains(output, "v0.1.0") {
		t.Error("Tag 'v0.1.0' should be created")
	}
}

func TestGitRepo_HasUncommittedChanges(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 変更がない状態
	if repo.HasUncommittedChanges() {
		t.Error("Should have no uncommitted changes")
	}

	// ファイルを変更
	repo.CreateFile("new.txt", "New file")

	if !repo.HasUncommittedChanges() {
		t.Error("Should have uncommitted changes")
	}
}

func TestGitRepo_CommitCount(t *testing.T) {
	repo := NewGitRepo(t)

	repo.CreateFile("README.md", "# Test")
	repo.Commit("First commit")

	if repo.CommitCount() != 1 {
		t.Errorf("Expected 1 commit, got %d", repo.CommitCount())
	}

	repo.CreateFile("file2.txt", "content")
	repo.Commit("Second commit")

	if repo.CommitCount() != 2 {
		t.Errorf("Expected 2 commits, got %d", repo.CommitCount())
	}

	repo.CreateFile("file3.txt", "content")
	repo.Commit("Third commit")

	if repo.CommitCount() != 3 {
		t.Errorf("Expected 3 commits, got %d", repo.CommitCount())
	}
}

func TestGitRepo_Path(t *testing.T) {
	repo := NewGitRepo(t)

	path := repo.Path("src/main.go")
	expected := filepath.Join(repo.Dir, "src/main.go")

	if path != expected {
		t.Errorf("Path mismatch: got %q, want %q", path, expected)
	}
}

func TestGitRepo_FileExists(t *testing.T) {
	repo := NewGitRepo(t)

	// 存在しないファイル
	if repo.FileExists("nonexistent.txt") {
		t.Error("File should not exist")
	}

	// ファイルを作成
	repo.CreateFile("test.txt", "content")

	if !repo.FileExists("test.txt") {
		t.Error("File should exist")
	}
}

func TestGitRepo_StashOperations(t *testing.T) {
	repo := NewGitRepo(t)

	// 初期コミット
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 変更を加える
	repo.CreateFile("new.txt", "content")
	repo.MustGit("add", ".")

	// スタッシュ
	repo.StashPush("Test stash")

	// 変更がなくなっていることを確認
	if repo.HasUncommittedChanges() {
		t.Error("Should have no uncommitted changes after stash")
	}

	// スタッシュを適用
	repo.StashPop()

	// 変更が戻っていることを確認
	if !repo.HasUncommittedChanges() {
		t.Error("Should have uncommitted changes after stash pop")
	}
}

func TestGitRepo_MultipleCommits(t *testing.T) {
	repo := NewGitRepo(t)

	// 複数のコミットを作成
	for i := 1; i <= 5; i++ {
		repo.CreateFile("file.txt", strings.Repeat("x", i))
		repo.Commit("Commit " + string(rune('0'+i)))
	}

	if repo.CommitCount() != 5 {
		t.Errorf("Expected 5 commits, got %d", repo.CommitCount())
	}
}
