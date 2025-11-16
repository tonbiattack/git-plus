package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestBatchCloneCmd_CommandSetup はbatch-cloneコマンドの設定をテストします
func TestBatchCloneCmd_CommandSetup(t *testing.T) {
	if batchCloneCmd.Use != "batch-clone <file>" {
		t.Errorf("batchCloneCmd.Use = %q, want %q", batchCloneCmd.Use, "batch-clone <file>")
	}

	if batchCloneCmd.Short == "" {
		t.Error("batchCloneCmd.Short should not be empty")
	}

	if batchCloneCmd.Long == "" {
		t.Error("batchCloneCmd.Long should not be empty")
	}

	if batchCloneCmd.Example == "" {
		t.Error("batchCloneCmd.Example should not be empty")
	}
}

// TestBatchCloneCmd_Flags はフラグが正しく設定されていることを確認します
func TestBatchCloneCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"dir", "d"},
		{"shallow", "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := batchCloneCmd.Flags().Lookup(tt.name)
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

// TestExtractRepoName はリポジトリ名の抽出をテストします
func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo", "user/repo"},
		{"https://github.com/user/repo.git", "user/repo"},
		{"git@github.com:user/repo.git", "git@github.com:user/repo"},
		{"https://github.com/org/project", "org/project"},
		{"https://gitlab.com/group/subgroup/project", "subgroup/project"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := extractRepoName(tt.url)
			if result != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// TestReadRepositoryURLs はファイルからURLを読み込むテスト
func TestReadRepositoryURLs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "repos.txt")

	content := `# コメント行
https://github.com/user/repo1
https://github.com/user/repo2

# 空行の後
git@github.com:user/repo3.git
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := readRepositoryURLs(filePath)
	if err != nil {
		t.Errorf("readRepositoryURLs returned error: %v", err)
	}

	if len(urls) != 3 {
		t.Errorf("Expected 3 URLs, got %d", len(urls))
	}

	expectedURLs := []string{
		"https://github.com/user/repo1",
		"https://github.com/user/repo2",
		"git@github.com:user/repo3.git",
	}

	for i, expected := range expectedURLs {
		if i < len(urls) && urls[i] != expected {
			t.Errorf("URL[%d] = %q, want %q", i, urls[i], expected)
		}
	}
}

// TestReadRepositoryURLs_EmptyFile は空ファイルの読み込みをテスト
func TestReadRepositoryURLs_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := readRepositoryURLs(filePath)
	if err != nil {
		t.Errorf("readRepositoryURLs returned error: %v", err)
	}

	if len(urls) != 0 {
		t.Errorf("Expected 0 URLs, got %d", len(urls))
	}
}

// TestReadRepositoryURLs_OnlyComments はコメントのみのファイルをテスト
func TestReadRepositoryURLs_OnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "comments.txt")

	content := `# コメント1
# コメント2
# コメント3
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := readRepositoryURLs(filePath)
	if err != nil {
		t.Errorf("readRepositoryURLs returned error: %v", err)
	}

	if len(urls) != 0 {
		t.Errorf("Expected 0 URLs, got %d", len(urls))
	}
}

// TestReadRepositoryURLs_InvalidURLs は無効なURLの処理をテスト
func TestReadRepositoryURLs_InvalidURLs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.txt")

	content := `https://github.com/user/valid
invalid-url-without-protocol
ftp://example.com/repo
`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := readRepositoryURLs(filePath)
	if err != nil {
		t.Errorf("readRepositoryURLs returned error: %v", err)
	}

	// https:// のみが有効
	if len(urls) != 1 {
		t.Errorf("Expected 1 valid URL, got %d", len(urls))
	}
}

// TestBatchCloneCmd_InRootCmd はbatch-cloneコマンドがRootCmdに登録されていることを確認します
func TestBatchCloneCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "batch-clone <file>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("batchCloneCmd should be registered in RootCmd")
	}
}
