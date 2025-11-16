package commit

import (
	"os"
	"strings"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestSquashCmdDefinition はsquashコマンドの定義をテストします
func TestSquashCmdDefinition(t *testing.T) {
	if squashCmd == nil {
		t.Fatal("squashCmd should not be nil")
	}

	if squashCmd.Use != "squash [コミット数]" {
		t.Errorf("squashCmd.Use = %q, want %q", squashCmd.Use, "squash [コミット数]")
	}

	if squashCmd.Short == "" {
		t.Error("squashCmd.Short should not be empty")
	}

	if squashCmd.Long == "" {
		t.Error("squashCmd.Long should not be empty")
	}

	if squashCmd.Example == "" {
		t.Error("squashCmd.Example should not be empty")
	}

	if squashCmd.RunE == nil {
		t.Error("squashCmd.RunE should not be nil")
	}
}

// TestCommitInfoStruct はcommitInfo構造体をテストします
func TestCommitInfoStruct(t *testing.T) {
	c := commitInfo{
		hash:    "abc123def456",
		subject: "Fix bug in user authentication",
	}

	if c.hash != "abc123def456" {
		t.Errorf("commitInfo.hash = %q, want %q", c.hash, "abc123def456")
	}

	if c.subject != "Fix bug in user authentication" {
		t.Errorf("commitInfo.subject = %q, want %q", c.subject, "Fix bug in user authentication")
	}
}

// TestGetRecentCommitsList は最近のコミットリスト取得をテストします
func TestGetRecentCommitsList(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 複数のコミットを作成
	repo.CreateFile("file1.txt", "content 1")
	repo.Commit("First commit")

	repo.CreateFile("file2.txt", "content 2")
	repo.Commit("Second commit")

	repo.CreateFile("file3.txt", "content 3")
	repo.Commit("Third commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits, err := getRecentCommitsList(3)
	if err != nil {
		t.Errorf("getRecentCommitsList returned error: %v", err)
	}

	if len(commits) != 3 {
		t.Errorf("Expected 3 commits, got %d", len(commits))
	}

	// コミットの順序を確認（最新が最初）
	if !strings.Contains(commits[0].subject, "Third") {
		t.Errorf("First commit should be 'Third commit', got %q", commits[0].subject)
	}

	if !strings.Contains(commits[1].subject, "Second") {
		t.Errorf("Second commit should be 'Second commit', got %q", commits[1].subject)
	}

	if !strings.Contains(commits[2].subject, "First") {
		t.Errorf("Third commit should be 'First commit', got %q", commits[2].subject)
	}

	// ハッシュが空でないことを確認
	for i, c := range commits {
		if c.hash == "" {
			t.Errorf("Commit %d hash should not be empty", i)
		}
		if len(c.hash) < 40 {
			t.Errorf("Commit %d hash should be at least 40 characters, got %d", i, len(c.hash))
		}
	}
}

// TestGetRecentCommitsList_LessThanRequested は要求より少ないコミットの場合のテスト
func TestGetRecentCommitsList_LessThanRequested(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 2つのコミットを作成
	repo.CreateFile("file1.txt", "content 1")
	repo.Commit("First commit")

	repo.CreateFile("file2.txt", "content 2")
	repo.Commit("Second commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 10個要求するが2つしかない
	commits, err := getRecentCommitsList(10)
	if err != nil {
		t.Errorf("getRecentCommitsList returned error: %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("Expected 2 commits (all available), got %d", len(commits))
	}
}

// TestGetRecentCommitsList_SingleCommit は単一コミットの場合のテスト
func TestGetRecentCommitsList_SingleCommit(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	repo.CreateFile("file1.txt", "content 1")
	repo.Commit("Only commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits, err := getRecentCommitsList(1)
	if err != nil {
		t.Errorf("getRecentCommitsList returned error: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}

	if !strings.Contains(commits[0].subject, "Only commit") {
		t.Errorf("Expected 'Only commit', got %q", commits[0].subject)
	}
}

// TestGetRecentCommitsList_NoCommits はコミットがない場合のテスト
func TestGetRecentCommitsList_NoCommits(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// コミットなし
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err = getRecentCommitsList(5)
	// コミットがない場合はエラーになるはず
	if err == nil {
		t.Error("getRecentCommitsList should return error when no commits exist")
	}
}

// TestGetRecentCommitsList_LongSubject は長い件名のコミットのテスト
func TestGetRecentCommitsList_LongSubject(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	longSubject := "This is a very long commit message that describes a complex change in the codebase"
	repo.CreateFile("file1.txt", "content 1")
	repo.Commit(longSubject)

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits, err := getRecentCommitsList(1)
	if err != nil {
		t.Errorf("getRecentCommitsList returned error: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}

	if commits[0].subject != longSubject {
		t.Errorf("Subject mismatch: got %q, want %q", commits[0].subject, longSubject)
	}
}

// TestGetRecentCommitsList_SpecialCharacters は特殊文字を含むコミットのテスト
func TestGetRecentCommitsList_SpecialCharacters(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 日本語のコミットメッセージ
	japaneseSubject := "日本語のコミットメッセージ"
	repo.CreateFile("file1.txt", "content 1")
	repo.Commit(japaneseSubject)

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits, err := getRecentCommitsList(1)
	if err != nil {
		t.Errorf("getRecentCommitsList returned error: %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(commits))
	}

	if commits[0].subject != japaneseSubject {
		t.Errorf("Subject mismatch: got %q, want %q", commits[0].subject, japaneseSubject)
	}
}

// TestSquashCommandRegistration はsquashコマンドがRootCmdに登録されているかテストします
func TestSquashCommandRegistration(t *testing.T) {
	foundCmd, _, err := cmd.RootCmd.Find([]string{"squash"})
	if err != nil {
		t.Errorf("squash command not found: %v", err)
	}
	if foundCmd == nil {
		t.Error("squash command is nil")
	}
	if foundCmd.Name() != "squash" {
		t.Errorf("Command name = %q, want %q", foundCmd.Name(), "squash")
	}
}

// TestCommitHashLength はコミットハッシュの長さをテストします
func TestCommitHashLength(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	repo.CreateFile("file1.txt", "content 1")
	repo.Commit("Test commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits, err := getRecentCommitsList(1)
	if err != nil {
		t.Fatalf("getRecentCommitsList returned error: %v", err)
	}

	// SHA-1ハッシュは40文字
	if len(commits[0].hash) != 40 {
		t.Errorf("Expected hash length 40, got %d", len(commits[0].hash))
	}

	// ハッシュは16進数文字のみで構成される
	for _, c := range commits[0].hash {
		isDigit := c >= '0' && c <= '9'
		isHexLetter := c >= 'a' && c <= 'f'
		if !isDigit && !isHexLetter {
			t.Errorf("Invalid character in hash: %c", c)
		}
	}
}
