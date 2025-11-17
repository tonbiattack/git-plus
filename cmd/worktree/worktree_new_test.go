package worktree

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestWorktreeNewCmd_CommandSetup はworktree-newコマンドの設定をテストします
func TestWorktreeNewCmd_CommandSetup(t *testing.T) {
	if worktreeNewCmd.Use != "worktree-new <branch-name>" {
		t.Errorf("worktreeNewCmd.Use = %q, want %q", worktreeNewCmd.Use, "worktree-new <branch-name>")
	}

	if worktreeNewCmd.Short == "" {
		t.Error("worktreeNewCmd.Short should not be empty")
	}

	if worktreeNewCmd.Long == "" {
		t.Error("worktreeNewCmd.Long should not be empty")
	}

	if worktreeNewCmd.Example == "" {
		t.Error("worktreeNewCmd.Example should not be empty")
	}
}

// TestWorktreeNewCmd_InRootCmd はworktree-newコマンドがrootCmdに登録されていることを確認します
func TestWorktreeNewCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "worktree-new <branch-name>" {
			found = true
			break
		}
	}

	if !found {
		t.Error("worktreeNewCmd should be registered in rootCmd")
	}
}

// TestWorktreeNewCmd_HasRunE はRunE関数が設定されていることを確認します
func TestWorktreeNewCmd_HasRunE(t *testing.T) {
	if worktreeNewCmd.RunE == nil {
		t.Error("worktreeNewCmd.RunE should not be nil")
	}
}

// TestWorktreeNewCmd_NoCodeFlag は--no-codeフラグが正しく設定されていることを確認します
func TestWorktreeNewCmd_NoCodeFlag(t *testing.T) {
	flag := worktreeNewCmd.Flags().Lookup("no-code")
	if flag == nil {
		t.Error("worktreeNewCmd should have --no-code flag")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("--no-code flag default value = %q, want %q", flag.DefValue, "false")
	}

	if flag.Usage == "" {
		t.Error("--no-code flag should have usage description")
	}
}

// TestWorktreeNewCmd_BaseFlag は--baseフラグが正しく設定されていることを確認します
func TestWorktreeNewCmd_BaseFlag(t *testing.T) {
	flag := worktreeNewCmd.Flags().Lookup("base")
	if flag == nil {
		t.Error("worktreeNewCmd should have --base flag")
		return
	}

	if flag.DefValue != "" {
		t.Errorf("--base flag default value = %q, want %q", flag.DefValue, "")
	}

	if flag.Usage == "" {
		t.Error("--base flag should have usage description")
	}
}

// TestWorktreeNewCmd_RequiresArgs はworktree-newコマンドが引数を必要とすることを確認します
func TestWorktreeNewCmd_RequiresArgs(t *testing.T) {
	if worktreeNewCmd.Args == nil {
		t.Error("worktreeNewCmd should have Args validation")
	}
}

// TestCheckBranchExists_InvalidBranch はブランチ存在確認の動作をテストします
func TestCheckBranchExists_InvalidBranch(t *testing.T) {
	// 非常に長いブランチ名の場合でもpanicしないことを確認
	_, err := checkBranchExists("very-long-branch-name-that-does-not-exist-" + string(make([]byte, 100)))
	// エラーが発生する可能性があるが、panicしないことが重要
	_ = err
}
