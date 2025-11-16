package release

import (
	"fmt"
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestReleaseNotesCmd_CommandSetup はrelease-notesコマンドの設定をテストします
func TestReleaseNotesCmd_CommandSetup(t *testing.T) {
	if releaseNotesCmd.Use != "release-notes" {
		t.Errorf("releaseNotesCmd.Use = %q, want %q", releaseNotesCmd.Use, "release-notes")
	}

	if releaseNotesCmd.Short == "" {
		t.Error("releaseNotesCmd.Short should not be empty")
	}

	if releaseNotesCmd.Long == "" {
		t.Error("releaseNotesCmd.Long should not be empty")
	}

	if releaseNotesCmd.Example == "" {
		t.Error("releaseNotesCmd.Example should not be empty")
	}
}

// TestReleaseNotesCmd_Flags はフラグが正しく設定されていることを確認します
func TestReleaseNotesCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"tag", "t"},
		{"draft", "d"},
		{"prerelease", "p"},
		{"latest", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := releaseNotesCmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.name)
				return
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// TestReleaseNotesCmd_InRootCmd はrelease-notesコマンドがrootCmdに登録されていることを確認します
func TestReleaseNotesCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "release-notes" {
			found = true
			break
		}
	}

	if !found {
		t.Error("releaseNotesCmd should be registered in rootCmd")
	}
}

// TestGetRecentTags は最近のタグ取得をテストします
func TestGetRecentTags(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// タグを作成
	repo.CreateLightweightTag("v1.0.0")

	repo.CreateFile("file1.txt", "content1")
	repo.Commit("Add file1")
	repo.CreateLightweightTag("v1.1.0")

	repo.CreateFile("file2.txt", "content2")
	repo.Commit("Add file2")
	repo.CreateLightweightTag("v2.0.0")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tags, err := getRecentTags(10)
	if err != nil {
		t.Errorf("getRecentTags returned error: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}
}

// TestGetRecentTags_Limit はタグ数の制限をテストします
func TestGetRecentTags_Limit(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 5つのタグを作成
	for i := 1; i <= 5; i++ {
		repo.CreateFile("file.txt", fmt.Sprintf("content %d", i))
		repo.Commit(fmt.Sprintf("Commit %d", i))
		repo.CreateLightweightTag(fmt.Sprintf("v%d.0.0", i))
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 3つまでに制限
	tags, err := getRecentTags(3)
	if err != nil {
		t.Errorf("getRecentTags returned error: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags (limited), got %d", len(tags))
	}
}

// TestGetRecentTags_NoTags はタグがない場合をテストします
func TestGetRecentTags_NoTags(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成（タグなし）
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

	tags, err := getRecentTags(10)
	if err != nil {
		t.Errorf("getRecentTags returned error: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}
