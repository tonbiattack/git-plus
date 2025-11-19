package branch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestNormalizeAbortOperation はユーザー入力の正規化をテストします
func TestNormalizeAbortOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"merge", "merge"},
		{"Rebase", "rebase"},
		{"CHERRY-PICK", "cherry-pick"},
		{"cherry_pick", "cherry-pick"},
		{"cherry", "cherry-pick"},
		{"revert", "revert"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			actual, err := normalizeAbortOperation(tt.input)
			if err != nil {
				t.Fatalf("normalizeAbortOperation(%q) returned error: %v", tt.input, err)
			}
			if actual != tt.expected {
				t.Fatalf("normalizeAbortOperation(%q) = %q, want %q", tt.input, actual, tt.expected)
			}
		})
	}

	if _, err := normalizeAbortOperation("unknown"); err == nil {
		t.Fatalf("normalizeAbortOperation should fail for unsupported operation")
	}
}

// TestDetectAbortOperation_Merge はMERGE_HEAD検出をテストします
func TestDetectAbortOperation_Merge(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "MERGE_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "merge" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "merge")
		}
	})
}

// TestDetectAbortOperation_Rebase は rebase ディレクトリ検出をテストします
func TestDetectAbortOperation_Rebase(t *testing.T) {
	withRepo(t, func(gitDir string) {
		createDir(t, filepath.Join(gitDir, "rebase-merge"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "rebase" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "rebase")
		}
	})
}

// TestDetectAbortOperation_Revert は REVERT_HEAD を検出することをテストします
func TestDetectAbortOperation_Revert(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "REVERT_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "revert" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "revert")
		}
	})
}

// TestDetectAbortOperation_CherryPick は CHERRY_PICK_HEAD を検出することをテストします
func TestDetectAbortOperation_CherryPick(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "CHERRY_PICK_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "cherry-pick" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "cherry-pick")
		}
	})
}

// TestDetectAbortOperation_NoOp は操作が検出されない場合にエラーとなることをテストします
func TestDetectAbortOperation_NoOp(t *testing.T) {
	withRepo(t, func(_ string) {
		if _, err := detectAbortOperation(); err == nil {
			t.Fatalf("detectAbortOperation should fail when no operation is detected")
		}
	})
}

// withRepo は一時Gitリポジトリでコールバックを実行します
func withRepo(t *testing.T, fn func(gitDir string)) {
	t.Helper()

	repo := testutil.NewGitRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	fn(filepath.Join(repo.Dir, ".git"))
}

// writeIndicator は指定されたパスに空ファイルを作成します
func writeIndicator(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("dummy"), 0o600); err != nil {
		t.Fatalf("failed to write indicator file: %v", err)
	}
}

// createDir は指定されたディレクトリを作成します
func createDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
}
