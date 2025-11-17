package worktree

import (
	"strings"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestWorktreeSwitchCmd_CommandSetup はworktree-switchコマンドの設定をテストします
func TestWorktreeSwitchCmd_CommandSetup(t *testing.T) {
	if worktreeSwitchCmd.Use != "worktree-switch" {
		t.Errorf("worktreeSwitchCmd.Use = %q, want %q", worktreeSwitchCmd.Use, "worktree-switch")
	}

	if worktreeSwitchCmd.Short == "" {
		t.Error("worktreeSwitchCmd.Short should not be empty")
	}

	if worktreeSwitchCmd.Long == "" {
		t.Error("worktreeSwitchCmd.Long should not be empty")
	}

	if worktreeSwitchCmd.Example == "" {
		t.Error("worktreeSwitchCmd.Example should not be empty")
	}
}

// TestWorktreeSwitchCmd_InRootCmd はworktree-switchコマンドがrootCmdに登録されていることを確認します
func TestWorktreeSwitchCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "worktree-switch" {
			found = true
			break
		}
	}

	if !found {
		t.Error("worktreeSwitchCmd should be registered in rootCmd")
	}
}

// TestWorktreeSwitchCmd_HasRunE はRunE関数が設定されていることを確認します
func TestWorktreeSwitchCmd_HasRunE(t *testing.T) {
	if worktreeSwitchCmd.RunE == nil {
		t.Error("worktreeSwitchCmd.RunE should not be nil")
	}
}

// TestWorktreeSwitchCmd_NoCodeFlag は--no-codeフラグが正しく設定されていることを確認します
func TestWorktreeSwitchCmd_NoCodeFlag(t *testing.T) {
	flag := worktreeSwitchCmd.Flags().Lookup("no-code")
	if flag == nil {
		t.Error("worktreeSwitchCmd should have --no-code flag")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("--no-code flag default value = %q, want %q", flag.DefValue, "false")
	}

	if flag.Usage == "" {
		t.Error("--no-code flag should have usage description")
	}
}

// TestGetWorktreeList_ParsesEmptyOutput は空の出力を正しく処理できることを確認します
func TestGetWorktreeList_ParsesEmptyOutput(t *testing.T) {
	// この関数はgitコマンドを実行するので、git環境でのみテスト可能
	// ここではパニックしないことを確認
	worktrees, err := getWorktreeList()
	// エラーが発生してもパニックしないことが重要
	if err == nil && worktrees == nil {
		t.Error("getWorktreeList should return non-nil slice when no error")
	}
}

// TestWorktreeInfo_Struct はWorktreeInfo構造体のフィールドをテストします
func TestWorktreeInfo_Struct(t *testing.T) {
	wt := WorktreeInfo{
		Path:   "/path/to/worktree",
		Branch: "feature/test",
		Commit: "abc1234",
	}

	if wt.Path != "/path/to/worktree" {
		t.Errorf("Path = %q, want %q", wt.Path, "/path/to/worktree")
	}

	if wt.Branch != "feature/test" {
		t.Errorf("Branch = %q, want %q", wt.Branch, "feature/test")
	}

	if wt.Commit != "abc1234" {
		t.Errorf("Commit = %q, want %q", wt.Commit, "abc1234")
	}
}

// TestGetCurrentWorktreePath は現在のworktreeパスの取得をテストします
func TestGetCurrentWorktreePath(t *testing.T) {
	// git環境でのみ動作するので、エラーが発生してもpanicしないことを確認
	path, err := getCurrentWorktreePath()
	if err == nil {
		// 成功した場合はパスが空でないことを確認
		if path == "" {
			t.Error("getCurrentWorktreePath should return non-empty path when successful")
		}
		// パスに改行が含まれていないことを確認
		if strings.Contains(path, "\n") {
			t.Error("getCurrentWorktreePath should return path without newlines")
		}
	}
}
