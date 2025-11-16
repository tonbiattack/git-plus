package branch

import (
	"io"
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestCheckBranchExists_BranchExists は存在するブランチのチェックをテストします
func TestCheckBranchExists_BranchExists(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 新しいブランチを作成
	repo.CreateBranch("feature/test")

	// ディレクトリを変更してテスト実行
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	exists, err := checkBranchExists("feature/test")
	if err != nil {
		t.Errorf("checkBranchExists returned error: %v", err)
	}

	if !exists {
		t.Error("checkBranchExists should return true for existing branch")
	}
}

// TestCheckBranchExists_BranchNotExists は存在しないブランチのチェックをテストします
func TestCheckBranchExists_BranchNotExists(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ディレクトリを変更してテスト実行
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	exists, err := checkBranchExists("nonexistent-branch")
	if err != nil {
		t.Errorf("checkBranchExists returned error: %v", err)
	}

	if exists {
		t.Error("checkBranchExists should return false for nonexistent branch")
	}
}

// TestCheckBranchExists_MainBranch はmainブランチのチェックをテストします
func TestCheckBranchExists_MainBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ディレクトリを変更してテスト実行
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 現在のブランチ名を取得
	currentBranch := repo.CurrentBranch()

	// 現在のブランチは初期化時に作成されている
	exists, err := checkBranchExists(currentBranch)
	if err != nil {
		t.Errorf("checkBranchExists returned error: %v", err)
	}

	if !exists {
		t.Errorf("checkBranchExists should return true for current branch (%s)", currentBranch)
	}
}

// TestAskUserAction_Recreate はrecreateアクションをテストします
func TestAskUserAction_Recreate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"r入力", "r\n", "recreate"},
		{"recreate入力", "recreate\n", "recreate"},
		{"大文字R", "R\n", "recreate"},
		{"大文字RECREATE", "RECREATE\n", "recreate"},
		{"前後空白付き", "  r  \n", "recreate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準入力をモック
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() {
				os.Stdin = oldStdin
				_ = r.Close()
			}()

			// 標準出力を抑制
			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
			}()

			go func() {
				_, _ = io.WriteString(w, tt.input)
				_ = w.Close()
			}()

			action, err := askUserAction("test-branch")
			if err != nil {
				t.Errorf("askUserAction returned error: %v", err)
			}

			if action != tt.expected {
				t.Errorf("askUserAction = %q, want %q", action, tt.expected)
			}
		})
	}
}

// TestAskUserAction_Switch はswitchアクションをテストします
func TestAskUserAction_Switch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"s入力", "s\n", "switch"},
		{"switch入力", "switch\n", "switch"},
		{"大文字S", "S\n", "switch"},
		{"大文字SWITCH", "SWITCH\n", "switch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() {
				os.Stdin = oldStdin
				_ = r.Close()
			}()

			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
			}()

			go func() {
				_, _ = io.WriteString(w, tt.input)
				_ = w.Close()
			}()

			action, err := askUserAction("test-branch")
			if err != nil {
				t.Errorf("askUserAction returned error: %v", err)
			}

			if action != tt.expected {
				t.Errorf("askUserAction = %q, want %q", action, tt.expected)
			}
		})
	}
}

// TestAskUserAction_Cancel はcancelアクションをテストします
func TestAskUserAction_Cancel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"c入力", "c\n", "cancel"},
		{"cancel入力", "cancel\n", "cancel"},
		{"大文字C", "C\n", "cancel"},
		{"空入力", "\n", "cancel"},
		{"無効な入力", "invalid\n", "cancel"},
		{"数字入力", "123\n", "cancel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() {
				os.Stdin = oldStdin
				_ = r.Close()
			}()

			oldStdout := os.Stdout
			os.Stdout, _ = os.Open(os.DevNull)
			defer func() {
				os.Stdout = oldStdout
			}()

			go func() {
				_, _ = io.WriteString(w, tt.input)
				_ = w.Close()
			}()

			action, err := askUserAction("test-branch")
			if err != nil {
				t.Errorf("askUserAction returned error: %v", err)
			}

			if action != tt.expected {
				t.Errorf("askUserAction = %q, want %q", action, tt.expected)
			}
		})
	}
}

// TestAskUserAction_EOF はEOFの場合にcancelを返すことをテストします
func TestAskUserAction_EOF(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		_ = r.Close()
	}()

	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() {
		os.Stdout = oldStdout
	}()

	// 何も書かずにパイプを閉じてEOFをシミュレート
	_ = w.Close()

	action, err := askUserAction("test-branch")
	if err != nil {
		t.Errorf("askUserAction returned error: %v", err)
	}

	if action != "cancel" {
		t.Errorf("askUserAction = %q, want %q on EOF", action, "cancel")
	}
}

// TestIsBranchNotFound_WithNilError はnilエラーのテスト
func TestIsBranchNotFound_WithNilError(t *testing.T) {
	if isBranchNotFound(nil) {
		t.Error("isBranchNotFound(nil) should return false")
	}
}
