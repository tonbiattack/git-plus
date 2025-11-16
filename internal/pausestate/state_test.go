package pausestate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestHome はテスト用のホームディレクトリを設定する
func setupTestHome(t *testing.T) (cleanup func()) {
	t.Helper()

	// 元のHOMEを保存
	originalHome := os.Getenv("HOME")

	// 一時ディレクトリを作成
	tmpDir := t.TempDir()

	// HOMEを一時ディレクトリに設定
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// クリーンアップ関数を返す
	return func() {
		_ = os.Setenv("HOME", originalHome)
	}
}

func TestSaveAndLoad(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// テスト用の状態を作成
	now := time.Now()
	state := &PauseState{
		FromBranch:   "feature/test-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "WIP on feature/test-branch",
		Timestamp:    now,
	}

	// 状態を保存
	err := Save(state)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// 状態を読み込み
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil state")
	}

	// フィールドを検証
	if loaded.FromBranch != state.FromBranch {
		t.Errorf("FromBranch mismatch: got %q, want %q", loaded.FromBranch, state.FromBranch)
	}

	if loaded.ToBranch != state.ToBranch {
		t.Errorf("ToBranch mismatch: got %q, want %q", loaded.ToBranch, state.ToBranch)
	}

	if loaded.StashRef != state.StashRef {
		t.Errorf("StashRef mismatch: got %q, want %q", loaded.StashRef, state.StashRef)
	}

	if loaded.StashMessage != state.StashMessage {
		t.Errorf("StashMessage mismatch: got %q, want %q", loaded.StashMessage, state.StashMessage)
	}

	// タイムスタンプはJSON経由で精度が落ちる可能性があるため、秒単位で比較
	if loaded.Timestamp.Unix() != state.Timestamp.Unix() {
		t.Errorf("Timestamp mismatch: got %v, want %v", loaded.Timestamp, state.Timestamp)
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// ファイルが存在しない場合はnilを返す
	state, err := Load()
	if err != nil {
		t.Errorf("Load() returned error for nonexistent file: %v", err)
	}

	if state != nil {
		t.Errorf("Load() returned non-nil state for nonexistent file: %v", state)
	}
}

func TestExists_FileExists(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// 状態を保存
	state := &PauseState{
		FromBranch:   "test-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "test",
		Timestamp:    time.Now(),
	}

	if err := Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// ファイルが存在することを確認
	exists, err := Exists()
	if err != nil {
		t.Errorf("Exists() returned error: %v", err)
	}

	if !exists {
		t.Error("Exists() returned false, want true")
	}
}

func TestExists_FileNotExists(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// ファイルが存在しないことを確認
	exists, err := Exists()
	if err != nil {
		t.Errorf("Exists() returned error: %v", err)
	}

	if exists {
		t.Error("Exists() returned true, want false")
	}
}

func TestDelete_FileExists(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// 状態を保存
	state := &PauseState{
		FromBranch:   "test-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "test",
		Timestamp:    time.Now(),
	}

	if err := Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// ファイルが存在することを確認
	exists, _ := Exists()
	if !exists {
		t.Fatal("File should exist before deletion")
	}

	// ファイルを削除
	if err := Delete(); err != nil {
		t.Errorf("Delete() returned error: %v", err)
	}

	// ファイルが削除されたことを確認
	exists, _ = Exists()
	if exists {
		t.Error("File should not exist after deletion")
	}
}

func TestDelete_FileNotExists(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// 存在しないファイルを削除してもエラーにならない
	err := Delete()
	if err != nil {
		t.Errorf("Delete() returned error for nonexistent file: %v", err)
	}
}

func TestDelete_Idempotent(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// 状態を保存
	state := &PauseState{
		FromBranch:   "test-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "test",
		Timestamp:    time.Now(),
	}

	if err := Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// 複数回削除してもエラーにならない（冪等性）
	for i := 0; i < 3; i++ {
		if err := Delete(); err != nil {
			t.Errorf("Delete() iteration %d returned error: %v", i, err)
		}
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	homeDir := os.Getenv("HOME")
	gitPlusDir := filepath.Join(homeDir, ".git-plus")

	// ディレクトリが存在しないことを確認
	if _, err := os.Stat(gitPlusDir); !os.IsNotExist(err) {
		t.Skip("Directory already exists")
	}

	// 状態を保存（ディレクトリが自動作成される）
	state := &PauseState{
		FromBranch:   "test-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "test",
		Timestamp:    time.Now(),
	}

	if err := Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// ディレクトリが作成されたことを確認
	if _, err := os.Stat(gitPlusDir); err != nil {
		t.Errorf("Expected directory to be created: %v", err)
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	// 最初の状態を保存
	state1 := &PauseState{
		FromBranch:   "first-branch",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "first",
		Timestamp:    time.Now(),
	}

	if err := Save(state1); err != nil {
		t.Fatalf("First Save() failed: %v", err)
	}

	// 2番目の状態を保存（上書き）
	state2 := &PauseState{
		FromBranch:   "second-branch",
		ToBranch:     "develop",
		StashRef:     "stash@{1}",
		StashMessage: "second",
		Timestamp:    time.Now(),
	}

	if err := Save(state2); err != nil {
		t.Fatalf("Second Save() failed: %v", err)
	}

	// 読み込んで2番目の状態が保存されていることを確認
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.FromBranch != "second-branch" {
		t.Errorf("FromBranch should be overwritten: got %q, want %q", loaded.FromBranch, "second-branch")
	}
}

func TestPauseState_JSONFormat(t *testing.T) {
	cleanup := setupTestHome(t)
	defer cleanup()

	state := &PauseState{
		FromBranch:   "feature/my-feature",
		ToBranch:     "main",
		StashRef:     "stash@{0}",
		StashMessage: "WIP: 日本語メッセージテスト",
		Timestamp:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	if err := Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// JSONファイルの内容を直接確認
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".git-plus", "pause-state.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	content := string(data)

	// JSONに必要なキーが含まれていることを確認
	requiredKeys := []string{
		"from_branch",
		"to_branch",
		"stash_ref",
		"stash_message",
		"timestamp",
	}

	for _, key := range requiredKeys {
		if !contains(content, key) {
			t.Errorf("JSON should contain key %q", key)
		}
	}

	// 日本語が正しくエンコードされていることを確認
	if !contains(content, "日本語メッセージテスト") {
		t.Error("JSON should contain Japanese message correctly")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
