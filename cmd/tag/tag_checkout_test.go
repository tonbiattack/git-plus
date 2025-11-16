package tag

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestTagCheckoutCmd_CommandSetup はtag-checkoutコマンドの設定をテストします
func TestTagCheckoutCmd_CommandSetup(t *testing.T) {
	if tagCheckoutCmd.Use != "tag-checkout" {
		t.Errorf("tagCheckoutCmd.Use = %q, want %q", tagCheckoutCmd.Use, "tag-checkout")
	}

	if tagCheckoutCmd.Short == "" {
		t.Error("tagCheckoutCmd.Short should not be empty")
	}

	if tagCheckoutCmd.Long == "" {
		t.Error("tagCheckoutCmd.Long should not be empty")
	}

	if tagCheckoutCmd.Example == "" {
		t.Error("tagCheckoutCmd.Example should not be empty")
	}
}

// TestTagCheckoutCmd_Flags はフラグが正しく設定されていることを確認します
func TestTagCheckoutCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"limit", "n"},
		{"yes", "y"},
		{"latest", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tagCheckoutCmd.Flags().Lookup(tt.name)
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

// TestGetTagsSortedByVersion はタグの取得をテストします
func TestGetTagsSortedByVersion(t *testing.T) {
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

	tags, err := getTagsSortedByVersion()
	if err != nil {
		t.Errorf("getTagsSortedByVersion returned error: %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	// セマンティックバージョン順で最新が先頭にくるはず
	if len(tags) > 0 && tags[0] != "v2.0.0" {
		t.Errorf("First tag should be v2.0.0, got %s", tags[0])
	}
}

// TestGetTagsSortedByVersion_NoTags はタグがない場合をテストします
func TestGetTagsSortedByVersion_NoTags(t *testing.T) {
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

	tags, err := getTagsSortedByVersion()
	if err != nil {
		t.Errorf("getTagsSortedByVersion returned error: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(tags))
	}
}

// TestCheckoutTag はタグへのチェックアウトをテストします
func TestCheckoutTag(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")
	repo.CreateLightweightTag("v1.0.0")

	repo.CreateFile("file1.txt", "content1")
	repo.Commit("Add file1")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// タグにチェックアウト
	err = checkoutTag("v1.0.0")
	if err != nil {
		t.Errorf("checkoutTag returned error: %v", err)
	}

	// file1.txt が存在しないことを確認（タグ時点では存在しない）
	if repo.FileExists("file1.txt") {
		t.Error("file1.txt should not exist after checkout to v1.0.0")
	}
}

// TestCheckoutTag_NonexistentTag は存在しないタグへのチェックアウトをテストします
func TestCheckoutTag_NonexistentTag(t *testing.T) {
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

	// 存在しないタグへのチェックアウトはエラーになるはず
	err = checkoutTag("nonexistent-tag")
	if err == nil {
		t.Error("checkoutTag should return error for nonexistent tag")
	}
}

// TestTagCheckoutCmd_InRootCmd はtag-checkoutコマンドがRootCmdに登録されていることを確認します
func TestTagCheckoutCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "tag-checkout" {
			found = true
			break
		}
	}

	if !found {
		t.Error("tagCheckoutCmd should be registered in RootCmd")
	}
}
