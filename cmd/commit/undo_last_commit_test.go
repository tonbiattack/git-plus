package commit

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestUndoLastCommitCmd_CommandSetup はundo-last-commitコマンドの設定をテストします
func TestUndoLastCommitCmd_CommandSetup(t *testing.T) {
	if undoLastCommitCmd.Use != "undo-last-commit" {
		t.Errorf("undoLastCommitCmd.Use = %q, want %q", undoLastCommitCmd.Use, "undo-last-commit")
	}

	if undoLastCommitCmd.Short == "" {
		t.Error("undoLastCommitCmd.Short should not be empty")
	}

	if undoLastCommitCmd.Long == "" {
		t.Error("undoLastCommitCmd.Long should not be empty")
	}
}

// TestUndoLastCommitCmd_InRootCmd はundo-last-commitコマンドがRootCmdに登録されていることを確認します
func TestUndoLastCommitCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "undo-last-commit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("undoLastCommitCmd should be registered in RootCmd")
	}
}

// TestUndoLastCommitCmd_HasRunE はRunE関数が設定されていることを確認します
func TestUndoLastCommitCmd_HasRunE(t *testing.T) {
	if undoLastCommitCmd.RunE == nil {
		t.Error("undoLastCommitCmd.RunE should not be nil")
	}
}

// TestUndoLastCommitCmd_UndoCommit はコミット取り消しをテストします
func TestUndoLastCommitCmd_UndoCommit(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// もう1つコミットを作成
	repo.CreateFile("file1.txt", "content")
	repo.Commit("Add file1")

	initialCommitCount := repo.CommitCount()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// undo-last-commitを実行
	err = undoLastCommitCmd.RunE(undoLastCommitCmd, []string{})
	if err != nil {
		t.Errorf("undoLastCommitCmd.RunE returned error: %v", err)
	}

	// コミット数が減っていることを確認
	newCommitCount := repo.CommitCount()
	if newCommitCount != initialCommitCount-1 {
		t.Errorf("Commit count = %d, want %d", newCommitCount, initialCommitCount-1)
	}

	// ファイルは存在するが、未コミットの状態になっているはず
	if !repo.FileExists("file1.txt") {
		t.Error("file1.txt should still exist after undo")
	}

	// 変更がステージングされているはず
	if !repo.HasUncommittedChanges() {
		t.Error("Should have uncommitted changes after undo")
	}
}

// TestUndoLastCommitCmd_NoCommitToUndo は取り消すコミットがない場合をテストします
func TestUndoLastCommitCmd_NoCommitToUndo(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットのみ
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

	// 初期コミットを取り消そうとするとエラーになるはず
	err = undoLastCommitCmd.RunE(undoLastCommitCmd, []string{})
	if err == nil {
		t.Error("undoLastCommitCmd.RunE should return error when there's no commit to undo")
	}
}
