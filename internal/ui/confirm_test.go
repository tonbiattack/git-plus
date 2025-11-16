package ui

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// mockStdin は標準入力をモック化するヘルパー関数
func mockStdin(t *testing.T, input string) (cleanup func()) {
	t.Helper()

	// 元のStdinを保存
	originalStdin := os.Stdin

	// パイプを作成
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// 標準入力をパイプの読み取り側に置き換え
	os.Stdin = r

	// 入力をパイプに書き込む
	go func() {
		defer func() { _ = w.Close() }()
		_, _ = io.WriteString(w, input)
	}()

	// クリーンアップ関数を返す
	return func() {
		os.Stdin = originalStdin
		_ = r.Close()
	}
}

// captureStdout は標準出力をキャプチャするヘルパー関数
func captureStdout(t *testing.T) (getOutput func() string, cleanup func()) {
	t.Helper()

	// 元のStdoutを保存
	originalStdout := os.Stdout

	// パイプを作成
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// 標準出力をパイプの書き込み側に置き換え
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan struct{})

	// 別ゴルーチンで出力を読み取る
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	getOutput = func() string {
		_ = w.Close()
		<-done
		return buf.String()
	}

	cleanup = func() {
		os.Stdout = originalStdout
		_ = r.Close()
	}

	return getOutput, cleanup
}

func TestConfirm_YesInputs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{
			name:       "小文字y",
			input:      "y\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "大文字Y",
			input:      "Y\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "小文字yes",
			input:      "yes\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "大文字YES",
			input:      "YES\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "混合大文字Yes",
			input:      "Yes\n",
			defaultYes: true,
			expected:   true,
		},
		{
			name:       "前後に空白",
			input:      "  y  \n",
			defaultYes: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := mockStdin(t, tt.input)
			defer cleanup()

			// 標準出力もキャプチャ（プロンプト表示を抑制）
			_, cleanupStdout := captureStdout(t)
			defer cleanupStdout()

			result := Confirm("Test prompt", tt.defaultYes)
			if result != tt.expected {
				t.Errorf("Confirm() with input %q = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfirm_NoInputs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{
			name:       "小文字n",
			input:      "n\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "大文字N",
			input:      "N\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "小文字no",
			input:      "no\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "大文字NO",
			input:      "NO\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "混合大文字No",
			input:      "No\n",
			defaultYes: true,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := mockStdin(t, tt.input)
			defer cleanup()

			_, cleanupStdout := captureStdout(t)
			defer cleanupStdout()

			result := Confirm("Test prompt", tt.defaultYes)
			if result != tt.expected {
				t.Errorf("Confirm() with input %q = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfirm_EmptyInput_DefaultYes(t *testing.T) {
	cleanup := mockStdin(t, "\n")
	defer cleanup()

	_, cleanupStdout := captureStdout(t)
	defer cleanupStdout()

	result := Confirm("Test prompt", true)
	if !result {
		t.Error("Confirm() with empty input and defaultYes=true should return true")
	}
}

func TestConfirm_EmptyInput_DefaultNo(t *testing.T) {
	cleanup := mockStdin(t, "\n")
	defer cleanup()

	_, cleanupStdout := captureStdout(t)
	defer cleanupStdout()

	result := Confirm("Test prompt", false)
	if result {
		t.Error("Confirm() with empty input and defaultYes=false should return false")
	}
}

func TestConfirm_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultYes bool
		expected   bool
	}{
		{
			name:       "無効な入力",
			input:      "invalid\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "数字入力",
			input:      "123\n",
			defaultYes: true,
			expected:   false,
		},
		{
			name:       "日本語入力",
			input:      "はい\n",
			defaultYes: true,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := mockStdin(t, tt.input)
			defer cleanup()

			_, cleanupStdout := captureStdout(t)
			defer cleanupStdout()

			result := Confirm("Test prompt", tt.defaultYes)
			if result != tt.expected {
				t.Errorf("Confirm() with input %q = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNumberInput_BasicNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "1"},
		{"2", "2"},
		{"3", "3"},
		{"10", "10"},
		{"100", "100"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeNumberInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNumberInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNumberInput_FullWidthNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"全角1", "１", "1"},
		{"全角2", "２", "2"},
		{"全角3", "３", "3"},
		{"全角4", "４", "4"},
		{"全角5", "５", "5"},
		{"全角6", "６", "6"},
		{"全角7", "７", "7"},
		{"全角8", "８", "8"},
		{"全角9", "９", "9"},
		{"全角0", "０", "0"},
		{"全角10", "１０", "10"},
		{"全角123", "１２３", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeNumberInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNumberInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNumberInput_MixedWidthNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"混合1と全角2", "1２", "12"},
		{"全角1と2", "１2", "12"},
		{"混合123", "１2３", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeNumberInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNumberInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNumberInput_TrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"前方空白", "  1", "1"},
		{"後方空白", "1  ", "1"},
		{"両側空白", "  1  ", "1"},
		{"タブ", "\t1\t", "1"},
		{"改行", "\n1\n", "1"},
		{"混合空白", " \t1\n ", "1"},
		{"全角数字と空白", "  １２３  ", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeNumberInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNumberInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeNumberInput_EmptyString(t *testing.T) {
	result := NormalizeNumberInput("")
	if result != "" {
		t.Errorf("NormalizeNumberInput(\"\") = %q, want \"\"", result)
	}
}

func TestNormalizeNumberInput_OnlyWhitespace(t *testing.T) {
	result := NormalizeNumberInput("   ")
	if result != "" {
		t.Errorf("NormalizeNumberInput(\"   \") = %q, want \"\"", result)
	}
}

func TestNormalizeNumberInput_NonNumericInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"アルファベット", "abc", "abc"},
		{"日本語", "あいう", "あいう"},
		{"記号", "!@#", "!@#"},
		{"数字と文字混合", "1a2b", "1a2b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeNumberInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeNumberInput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkNormalizeNumberInput_HalfWidth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeNumberInput("12345")
	}
}

func BenchmarkNormalizeNumberInput_FullWidth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeNumberInput("１２３４５")
	}
}

func BenchmarkNormalizeNumberInput_Mixed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeNumberInput("  １2３4５  ")
	}
}
