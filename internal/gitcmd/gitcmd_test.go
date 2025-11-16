package gitcmd

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestRun_GitVersion(t *testing.T) {
	// git --version は常に成功するはず
	output, err := Run("--version")
	if err != nil {
		t.Fatalf("Run(--version) returned error: %v", err)
	}

	result := string(output)
	if !strings.HasPrefix(result, "git version") {
		t.Errorf("Expected output to start with 'git version', got: %s", result)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	_, err := Run("invalid-command-that-does-not-exist")
	if err == nil {
		t.Error("Expected error for invalid git command, got nil")
	}
}

func TestRun_MultipleArgs(t *testing.T) {
	// git config --get-regexp user を使用
	// 設定がなくてもエラーにはならない
	_, err := Run("config", "--list")
	if err != nil {
		t.Errorf("Run(config --list) returned unexpected error: %v", err)
	}
}

func TestRunQuiet_Success(t *testing.T) {
	// git --version を静かに実行
	err := RunQuiet("--version")
	if err != nil {
		t.Errorf("RunQuiet(--version) returned error: %v", err)
	}
}

func TestRunQuiet_InvalidCommand(t *testing.T) {
	err := RunQuiet("invalid-command-that-does-not-exist")
	if err == nil {
		t.Error("Expected error for invalid git command, got nil")
	}
}

func TestIsExitError_WithExitError(t *testing.T) {
	// 一時ディレクトリを作成（gitリポジトリではない）
	tmpDir := t.TempDir()

	// gitリポジトリではないディレクトリでgit statusを実行して終了エラーを取得
	cmd := exec.Command("git", "status")
	cmd.Dir = tmpDir
	err := cmd.Run()

	if err == nil {
		t.Skip("Expected error but got none")
	}

	// 実際の終了コードを取得
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected *exec.ExitError, got %T", err)
	}
	actualCode := exitErr.ExitCode()

	tests := []struct {
		name         string
		checkCode    int
		expectedBool bool
	}{
		{
			name:         "実際の終了コードをチェック",
			checkCode:    actualCode,
			expectedBool: true,
		},
		{
			name:         "異なるコードをチェック",
			checkCode:    actualCode + 1,
			expectedBool: false,
		},
		{
			name:         "コード0をチェック",
			checkCode:    0,
			expectedBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExitError(err, tt.checkCode)
			if result != tt.expectedBool {
				t.Errorf("IsExitError(err, %d) = %v, want %v", tt.checkCode, result, tt.expectedBool)
			}
		})
	}
}

func TestIsExitError_WithNilError(t *testing.T) {
	result := IsExitError(nil, 1)
	if result {
		t.Error("Expected false for nil error, got true")
	}
}

func TestIsExitError_WithGenericError(t *testing.T) {
	err := errors.New("generic error")
	result := IsExitError(err, 1)
	if result {
		t.Error("Expected false for generic error, got true")
	}
}

func TestIsExitError_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     int
		expected bool
	}{
		{
			name:     "nilエラー",
			err:      nil,
			code:     1,
			expected: false,
		},
		{
			name:     "一般エラー",
			err:      errors.New("general error"),
			code:     1,
			expected: false,
		},
		{
			name:     "異なるエラー型",
			err:      errors.New("not an exit error"),
			code:     128,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExitError(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("IsExitError(%v, %d) = %v, want %v",
					tt.err, tt.code, result, tt.expected)
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkRun_Version(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Run("--version")
	}
}

func BenchmarkRunQuiet_Version(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = RunQuiet("--version")
	}
}
