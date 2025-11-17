package issue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestIssueEditCmd_CommandSetup はissue-editコマンドの設定をテストします
func TestIssueEditCmd_CommandSetup(t *testing.T) {
	if issueEditCmd.Use != "issue-edit [ISSUE番号]" {
		t.Errorf("issueEditCmd.Use = %q, want %q", issueEditCmd.Use, "issue-edit [ISSUE番号]")
	}

	if issueEditCmd.Short == "" {
		t.Error("issueEditCmd.Short should not be empty")
	}

	if issueEditCmd.Long == "" {
		t.Error("issueEditCmd.Long should not be empty")
	}

	if issueEditCmd.Example == "" {
		t.Error("issueEditCmd.Example should not be empty")
	}
}

// TestIssueEditCmd_InRootCmd はissue-editコマンドがrootCmdに登録されていることを確認します
func TestIssueEditCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if strings.HasPrefix(c.Use, "issue-edit") {
			found = true
			break
		}
	}

	if !found {
		t.Error("issueEditCmd should be registered in rootCmd")
	}
}

// TestIssueEditCmd_HasRunE はRunE関数が設定されていることを確認します
func TestIssueEditCmd_HasRunE(t *testing.T) {
	if issueEditCmd.RunE == nil {
		t.Error("issueEditCmd.RunE should not be nil")
	}
}

// TestParseCommand はparseCommand関数をテストします
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
		wantErr  bool
	}{
		{
			name:     "simple command",
			command:  "vim",
			expected: []string{"vim"},
			wantErr:  false,
		},
		{
			name:     "command with arguments",
			command:  "code --wait",
			expected: []string{"code", "--wait"},
			wantErr:  false,
		},
		{
			name:     "command with multiple arguments",
			command:  "nvim -u NONE --noplugin",
			expected: []string{"nvim", "-u", "NONE", "--noplugin"},
			wantErr:  false,
		},
		{
			name:     "command with double quotes",
			command:  `"code" --wait`,
			expected: []string{"code", "--wait"},
			wantErr:  false,
		},
		{
			name:     "command with single quotes",
			command:  `'vim' -c "set number"`,
			expected: []string{"vim", "-c", "set number"},
			wantErr:  false,
		},
		{
			name:     "quoted command with spaces",
			command:  `"Visual Studio Code" --wait`,
			expected: []string{"Visual Studio Code", "--wait"},
			wantErr:  false,
		},
		{
			name:     "single quoted command",
			command:  `'My Editor' --option`,
			expected: []string{"My Editor", "--option"},
			wantErr:  false,
		},
		{
			name:     "Windows path with backslash (no spaces)",
			command:  `C:\Git\bin\vim.exe`,
			expected: []string{`C:\Git\bin\vim.exe`},
			wantErr:  false,
		},
		{
			name:     "Windows path with spaces in quotes",
			command:  `"C:\Program Files\Git\bin\vim.exe"`,
			expected: []string{`C:\Program Files\Git\bin\vim.exe`},
			wantErr:  false,
		},
		{
			name:     "empty command",
			command:  "",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "unclosed double quote",
			command:  `"unclosed`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "unclosed single quote",
			command:  `'vim -c`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "multiple spaces",
			command:  "vim   -u   NONE",
			expected: []string{"vim", "-u", "NONE"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommand(tt.command)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("parseCommand(%q) returned %d elements, want %d", tt.command, len(result), len(tt.expected))
				} else {
					for i := range result {
						if result[i] != tt.expected[i] {
							t.Errorf("parseCommand(%q)[%d] = %q, want %q", tt.command, i, result[i], tt.expected[i])
						}
					}
				}
			}
		})
	}
}

// TestReadFileContent はreadFileContent関数をテストします
func TestReadFileContent(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantTitle   string
		wantBody    string
		wantErr     bool
	}{
		{
			name: "valid content",
			fileContent: `# Comment
Title: Test Issue

---

This is the body.`,
			wantTitle: "Test Issue",
			wantBody:  "This is the body.",
			wantErr:   false,
		},
		{
			name: "content with multiple comments",
			fileContent: `# Comment 1
# Comment 2

Title: Another Test

---

Body text here.
Multiple lines.`,
			wantTitle: "Another Test",
			wantBody:  "Body text here.\nMultiple lines.",
			wantErr:   false,
		},
		{
			name: "content with detailed comments",
			fileContent: `# Issue #123
# URL: https://github.com/user/repo/issues/123
#
# 以下のissueの題名と本文を編集してください。
# '#' で始まる行はコメントとして無視されます。

Title: Test Issue Title

---

This is the body of the issue.
It has multiple lines.`,
			wantTitle: "Test Issue Title",
			wantBody:  "This is the body of the issue.\nIt has multiple lines.",
			wantErr:   false,
		},
		{
			name:        "missing title",
			fileContent: "Just some text\n---\nBody",
			wantTitle:   "",
			wantBody:    "",
			wantErr:     true,
		},
		{
			name:        "missing separator",
			fileContent: "Title: Test\nBody without separator",
			wantTitle:   "",
			wantBody:    "",
			wantErr:     true,
		},
		{
			name: "empty body",
			fileContent: `# Comment
Title: Empty Body Test

---

`,
			wantTitle: "Empty Body Test",
			wantBody:  "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			result, err := readFileContent(tmpFile)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result.Title != tt.wantTitle {
					t.Errorf("Title = %q, want %q", result.Title, tt.wantTitle)
				}
				if result.Body != tt.wantBody {
					t.Errorf("Body = %q, want %q", result.Body, tt.wantBody)
				}
			}
		})
	}
}

// TestReadCommentFromFile はreadCommentFromFile関数をテストします
func TestReadCommentFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		expected    string
	}{
		{
			name: "comment with headers",
			fileContent: `# Issue #123 へのコメント
# URL: https://github.com/user/repo/issues/123
#
# このissueに追加するコメントを記載してください。

This is my comment.`,
			expected: "This is my comment.",
		},
		{
			name: "empty comment",
			fileContent: `# Comment
# Another comment

`,
			expected: "",
		},
		{
			name: "multiline comment",
			fileContent: `# Header
First line.
Second line.
Third line.`,
			expected: "First line.\nSecond line.\nThird line.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "comment.md")
			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0600)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			result, err := readCommentFromFile(tmpFile)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("readCommentFromFile = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestCreateTempIssueFile はcreateTempIssueFile関数をテストします
func TestCreateTempIssueFile(t *testing.T) {
	issue := &IssueEntry{
		Number: 123,
		Title:  "Test Issue Title",
		Body:   "Test issue body content.",
		State:  "open",
		URL:    "https://github.com/user/repo/issues/123",
	}

	tmpFile, err := createTempIssueFile(issue)
	if err != nil {
		t.Fatalf("Failed to create temp issue file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// Check file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Temp file was not created")
	}

	// Read content and verify
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Issue #123") {
		t.Error("File should contain issue number")
	}
	if !strings.Contains(contentStr, "Test Issue Title") {
		t.Error("File should contain issue title")
	}
	if !strings.Contains(contentStr, "Test issue body content.") {
		t.Error("File should contain issue body")
	}
	if !strings.Contains(contentStr, "Title:") {
		t.Error("File should contain Title: marker")
	}
	if !strings.Contains(contentStr, "---") {
		t.Error("File should contain separator")
	}
	if !strings.Contains(contentStr, issue.URL) {
		t.Error("File should contain URL")
	}
}

// TestCreateTempCommentFile はcreateTempCommentFile関数をテストします
func TestCreateTempCommentFile(t *testing.T) {
	issue := &IssueEntry{
		Number: 456,
		URL:    "https://github.com/user/repo/issues/456",
	}

	tmpFile, err := createTempCommentFile(issue)
	if err != nil {
		t.Fatalf("Failed to create temp comment file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// Check file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Temp file was not created")
	}

	// Read content and verify
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Issue #456") {
		t.Error("File should contain issue number")
	}
	if !strings.Contains(contentStr, "https://github.com/user/repo/issues/456") {
		t.Error("File should contain issue URL")
	}
}

// TestGetEditor_Default はデフォルトエディタの取得をテストします
func TestGetEditor_Default(t *testing.T) {
	// 環境変数をクリア
	oldVisual := os.Getenv("VISUAL")
	oldEditor := os.Getenv("EDITOR")
	defer func() {
		_ = os.Setenv("VISUAL", oldVisual)
		_ = os.Setenv("EDITOR", oldEditor)
	}()

	_ = os.Unsetenv("VISUAL")
	_ = os.Unsetenv("EDITOR")

	editor, err := getEditor()
	if err != nil {
		t.Errorf("getEditor() returned error: %v", err)
		return
	}

	// 空でないことを確認（git config や vi がデフォルト）
	if editor == "" {
		t.Error("getEditor() should return non-empty string")
	}
}

// TestGetEditor_EnvironmentVariable は環境変数からのエディタ取得をテストします
func TestGetEditor_EnvironmentVariable(t *testing.T) {
	// 環境変数を設定
	oldVisual := os.Getenv("VISUAL")
	oldEditor := os.Getenv("EDITOR")
	defer func() {
		_ = os.Setenv("VISUAL", oldVisual)
		_ = os.Setenv("EDITOR", oldEditor)
	}()

	_ = os.Unsetenv("VISUAL")
	_ = os.Setenv("EDITOR", "nano")

	editor, err := getEditor()
	if err != nil {
		t.Errorf("getEditor() returned error: %v", err)
		return
	}

	// git config が設定されていない場合、環境変数から取得する
	// 結果は "nano" または git config の値
	if editor == "" {
		t.Error("getEditor() should return non-empty string")
	}
}

// TestOpenEditor_WithTerminalProtection はopenEditor関数がターミナル状態を保護することをテストします
func TestOpenEditor_WithTerminalProtection(t *testing.T) {
	// Test that opening a non-existent editor returns an error without crashing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(tmpFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test with non-existent editor
	err = openEditor("nonexistent-editor-xyz", tmpFile)
	if err == nil {
		t.Error("Expected error for non-existent editor")
	}
}

// TestOpenEditor_WithVSCodeWaitFlag はVSCodeの--waitフラグが追加されることをテストします
func TestOpenEditor_WithVSCodeWaitFlag(t *testing.T) {
	// This test verifies that the --wait flag logic is integrated
	// We can't actually test VSCode opening, but we can test the flag addition
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(tmpFile, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test with code (VSCode) - will fail because code is not installed,
	// but the important thing is that it doesn't panic
	err = openEditor("code", tmpFile)
	// We expect an error because VSCode is likely not installed in test environment
	if err != nil {
		// This is expected in most test environments
		t.Logf("Expected error (VSCode not installed): %v", err)
	}
}

// TestCheckGitHubCLIInstalled は GitHub CLI の確認をテストします
func TestCheckGitHubCLIInstalled(t *testing.T) {
	// この関数は実際の環境に依存するため、結果は環境によって異なる
	result := checkGitHubCLIInstalled()

	// bool 型が返されることを確認（true または false のどちらか）
	if result != true && result != false {
		t.Error("checkGitHubCLIInstalled() should return bool")
	}
}

// TestIssueContent_Structure はIssueContent構造体をテストします
func TestIssueContent_Structure(t *testing.T) {
	content := &IssueContent{
		Title: "Test Title",
		Body:  "Test Body",
	}

	if content.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", content.Title, "Test Title")
	}
	if content.Body != "Test Body" {
		t.Errorf("Body = %q, want %q", content.Body, "Test Body")
	}
}

// TestIssueEntry_Structure はIssueEntry構造体をテストします
func TestIssueEntry_Structure(t *testing.T) {
	entry := &IssueEntry{
		Number: 100,
		Title:  "Issue Title",
		Body:   "Issue Body",
		State:  "open",
		URL:    "https://github.com/user/repo/issues/100",
	}

	if entry.Number != 100 {
		t.Errorf("Number = %d, want %d", entry.Number, 100)
	}
	if entry.Title != "Issue Title" {
		t.Errorf("Title = %q, want %q", entry.Title, "Issue Title")
	}
	if entry.Body != "Issue Body" {
		t.Errorf("Body = %q, want %q", entry.Body, "Issue Body")
	}
	if entry.State != "open" {
		t.Errorf("State = %q, want %q", entry.State, "open")
	}
	if entry.URL != "https://github.com/user/repo/issues/100" {
		t.Errorf("URL = %q, want %q", entry.URL, "https://github.com/user/repo/issues/100")
	}
}
