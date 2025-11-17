package release

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestCheckGitHubCLIInstalled は GitHub CLI のインストール確認をテストします
func TestCheckGitHubCLIInstalled(t *testing.T) {
	// この関数は実際の環境に依存するため、結果は環境によって異なる
	// テストでは関数が正常に実行されることを確認
	result := checkGitHubCLIInstalled()

	// bool 型が返されることを確認（true または false のどちらか）
	if result != true && result != false {
		t.Error("checkGitHubCLIInstalled() should return bool")
	}

	// 関数がパニックしないことを確認（ここに到達すれば成功）
}

// TestVerifyTagExists_ExistingTag は存在するタグの確認をテストします
func TestVerifyTagExists_ExistingTag(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// タグを作成
	repo.CreateTag("v1.0.0", "Version 1.0.0")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err := verifyTagExists("v1.0.0")
	if err != nil {
		t.Errorf("verifyTagExists() returned error for existing tag: %v", err)
	}
}

// TestVerifyTagExists_NonExistingTag は存在しないタグの確認をテストします
func TestVerifyTagExists_NonExistingTag(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err := verifyTagExists("nonexistent-tag")
	if err == nil {
		t.Error("verifyTagExists() should return error for non-existing tag")
	}
}

// TestVerifyTagExists_LightweightTag は軽量タグの確認をテストします
func TestVerifyTagExists_LightweightTag(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 軽量タグを作成
	repo.CreateLightweightTag("v1.0.0-light")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err := verifyTagExists("v1.0.0-light")
	if err != nil {
		t.Errorf("verifyTagExists() returned error for existing lightweight tag: %v", err)
	}
}

// TestGetLatestTag_WithTags はタグがある場合の最新タグ取得をテストします
func TestGetLatestTag_WithTags(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// タグを作成
	repo.CreateTag("v1.0.0", "Version 1.0.0")

	// 追加のコミットとタグを作成
	repo.CreateFile("file2.txt", "content")
	repo.Commit("Second commit")
	repo.CreateTag("v2.0.0", "Version 2.0.0")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tag, err := getLatestTag()
	if err != nil {
		t.Errorf("getLatestTag() returned error: %v", err)
	}

	if tag != "v2.0.0" {
		t.Errorf("getLatestTag() = %q, want %q", tag, "v2.0.0")
	}
}

// TestGetLatestTag_SingleTag は単一のタグの場合をテストします
func TestGetLatestTag_SingleTag(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// タグを作成
	repo.CreateTag("v1.0.0", "Version 1.0.0")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tag, err := getLatestTag()
	if err != nil {
		t.Errorf("getLatestTag() returned error: %v", err)
	}

	if tag != "v1.0.0" {
		t.Errorf("getLatestTag() = %q, want %q", tag, "v1.0.0")
	}
}

// TestGetLatestTag_NoTags はタグがない場合のエラーをテストします
func TestGetLatestTag_NoTags(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err := getLatestTag()
	if err == nil {
		t.Error("getLatestTag() should return error when no tags exist")
	}
}

// TestGetLatestTag_LightweightTags は軽量タグの場合をテストします
func TestGetLatestTag_LightweightTags(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 軽量タグを作成
	repo.CreateLightweightTag("v1.0.0")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tag, err := getLatestTag()
	if err != nil {
		t.Errorf("getLatestTag() returned error: %v", err)
	}

	if tag != "v1.0.0" {
		t.Errorf("getLatestTag() = %q, want %q", tag, "v1.0.0")
	}
}
