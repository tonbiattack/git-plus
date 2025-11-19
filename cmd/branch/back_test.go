package branch

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestBackCmd_CommandSetup はbackコマンドの設定をテストします
func TestBackCmd_CommandSetup(t *testing.T) {
	if backCmd.Use != "back" {
		t.Errorf("backCmd.Use = %q, want %q", backCmd.Use, "back")
	}

	if backCmd.Short == "" {
		t.Error("backCmd.Short should not be empty")
	}

	if backCmd.Long == "" {
		t.Error("backCmd.Long should not be empty")
	}

	if backCmd.Example == "" {
		t.Error("backCmd.Example should not be empty")
	}
}

// TestBackCmd_InRootCmd はbackコマンドがRootCmdに登録されていることを確認します
func TestBackCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "back" {
			found = true
			break
		}
	}

	if !found {
		t.Error("backCmd should be registered in RootCmd")
	}
}

// TestBackCmd_HasRunE はRunE関数が設定されていることを確認します
func TestBackCmd_HasRunE(t *testing.T) {
	if backCmd.RunE == nil {
		t.Error("backCmd.RunE should not be nil")
	}
}

// TestBackCmd_SwitchBranch はブランチ切り替え機能をテストします
func TestBackCmd_SwitchBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 現在のブランチ名を取得（環境によって異なる可能性がある）
	defaultBranch := repo.CurrentBranch()

	// 新しいブランチを作成して切り替え
	repo.CreateAndCheckoutBranch("feature/test")

	// デフォルトブランチに戻る
	repo.CheckoutBranch(defaultBranch)

	// 現在の作業ディレクトリを取得します。
	// - `os.Getwd()` は現在のカレントワーキングディレクトリの絶対パスを返します。
	// - 戻り値: (string, error)
	//   - string: カレントディレクトリのパス
	//   - error: 取得に失敗した場合のエラー（例: 権限がない、パス不在など）
	// このテストでは、テスト実行中にカレントディレクトリを一時的にリポジトリのディレクトリに変更するため、
	// 終了時に `defer` で元のディレクトリへ戻す目的で保存しています。
	// 現在の作業ディレクトリを取得します。
	// - `os.Getwd()` は現在のカレントワーキングディレクトリの絶対パスを返します。
	// - 戻り値: (string, error)
	//   - string: カレントディレクトリのパス
	//   - error: 取得に失敗した場合のエラー（例: 権限がない、パス不在など）
	// このテストでは、テスト実行中にカレントディレクトリを一時的にリポジトリのディレクトリに変更するため、
	// 終了時に `defer` で元のディレクトリへ戻す目的で保存しています。
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// backコマンドを実行（feature/testブランチに戻るはず）
	err = backCmd.RunE(backCmd, []string{})
	if err != nil {
		t.Errorf("backCmd.RunE returned error: %v", err)
	}

	// feature/testブランチに戻っていることを確認
	currentBranch := repo.CurrentBranch()
	if currentBranch != "feature/test" {
		t.Errorf("Current branch = %q, want %q", currentBranch, "feature/test")
	}
}

// TestBackCmd_MultipleSwitches は複数回の切り替えをテストします
func TestBackCmd_MultipleSwitches(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 複数のブランチを作成
	repo.CreateBranch("branch-a")
	repo.CreateBranch("branch-b")

	// branch-aに切り替え
	repo.CheckoutBranch("branch-a")
	// branch-bに切り替え
	repo.CheckoutBranch("branch-b")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// backコマンドを実行（branch-aに戻る）
	err = backCmd.RunE(backCmd, []string{})
	if err != nil {
		t.Errorf("backCmd.RunE returned error: %v", err)
	}

	currentBranch := repo.CurrentBranch()
	if currentBranch != "branch-a" {
		t.Errorf("Current branch = %q, want %q", currentBranch, "branch-a")
	}
}

// TestBackCmd_NoHistory は履歴がない場合のテスト
func TestBackCmd_NoHistory(t *testing.T) {
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

	// 履歴がない場合はエラーになるはず
	err = backCmd.RunE(backCmd, []string{})
	if err == nil {
		t.Error("backCmd.RunE should return error when there's no checkout history")
	}
}
