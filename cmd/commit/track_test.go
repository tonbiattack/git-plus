package commit

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestTrackCmd_CommandSetup はtrackコマンドの設定をテストします
func TestTrackCmd_CommandSetup(t *testing.T) {
	if trackCmd.Use != "track [リモート名] [ブランチ名]" {
		t.Errorf("trackCmd.Use = %q, want %q", trackCmd.Use, "track [リモート名] [ブランチ名]")
	}

	if trackCmd.Short == "" {
		t.Error("trackCmd.Short should not be empty")
	}

	if trackCmd.Long == "" {
		t.Error("trackCmd.Long should not be empty")
	}

	if trackCmd.Example == "" {
		t.Error("trackCmd.Example should not be empty")
	}
}

// TestTrackCmd_InRootCmd はtrackコマンドがRootCmdに登録されていることを確認します
func TestTrackCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "track [リモート名] [ブランチ名]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("trackCmd should be registered in RootCmd")
	}
}

// TestTrackCmd_HasRunE はRunE関数が設定されていることを確認します
func TestTrackCmd_HasRunE(t *testing.T) {
	if trackCmd.RunE == nil {
		t.Error("trackCmd.RunE should not be nil")
	}
}

// TestFetchCurrentBranch は現在のブランチ名の取得をテストします
func TestFetchCurrentBranch(t *testing.T) {
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

	branch, err := fetchCurrentBranch()
	if err != nil {
		t.Errorf("fetchCurrentBranch returned error: %v", err)
	}

	expectedBranch := repo.CurrentBranch()
	if branch != expectedBranch {
		t.Errorf("fetchCurrentBranch() = %q, want %q", branch, expectedBranch)
	}
}

// TestFetchCurrentBranch_AfterSwitch はブランチ切り替え後のテスト
func TestFetchCurrentBranch_AfterSwitch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// ブランチを作成して切り替え
	repo.CreateAndCheckoutBranch("feature/test")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	branch, err := fetchCurrentBranch()
	if err != nil {
		t.Errorf("fetchCurrentBranch returned error: %v", err)
	}

	if branch != "feature/test" {
		t.Errorf("fetchCurrentBranch() = %q, want %q", branch, "feature/test")
	}
}

// TestCheckRemoteRefExists はリモート参照の存在確認をテストします
func TestCheckRemoteRefExists(t *testing.T) {
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

	// リモートが設定されていないので、存在しないはず
	exists, err := checkRemoteRefExists("origin/main")
	if err != nil {
		t.Errorf("checkRemoteRefExists returned error: %v", err)
	}

	if exists {
		t.Error("checkRemoteRefExists should return false for nonexistent remote ref")
	}
}
