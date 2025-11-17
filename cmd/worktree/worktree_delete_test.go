package worktree

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestWorktreeDeleteCmd_CommandSetup はworktree-deleteコマンドの設定をテストします
func TestWorktreeDeleteCmd_CommandSetup(t *testing.T) {
	if worktreeDeleteCmd.Use != "worktree-delete" {
		t.Errorf("worktreeDeleteCmd.Use = %q, want %q", worktreeDeleteCmd.Use, "worktree-delete")
	}

	if worktreeDeleteCmd.Short == "" {
		t.Error("worktreeDeleteCmd.Short should not be empty")
	}

	if worktreeDeleteCmd.Long == "" {
		t.Error("worktreeDeleteCmd.Long should not be empty")
	}

	if worktreeDeleteCmd.Example == "" {
		t.Error("worktreeDeleteCmd.Example should not be empty")
	}
}

// TestWorktreeDeleteCmd_InRootCmd はworktree-deleteコマンドがrootCmdに登録されていることを確認します
func TestWorktreeDeleteCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "worktree-delete" {
			found = true
			break
		}
	}

	if !found {
		t.Error("worktreeDeleteCmd should be registered in rootCmd")
	}
}

// TestWorktreeDeleteCmd_HasRunE はRunE関数が設定されていることを確認します
func TestWorktreeDeleteCmd_HasRunE(t *testing.T) {
	if worktreeDeleteCmd.RunE == nil {
		t.Error("worktreeDeleteCmd.RunE should not be nil")
	}
}

// TestWorktreeDeleteCmd_ForceFlag は--forceフラグが正しく設定されていることを確認します
func TestWorktreeDeleteCmd_ForceFlag(t *testing.T) {
	flag := worktreeDeleteCmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("worktreeDeleteCmd should have --force flag")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("--force flag default value = %q, want %q", flag.DefValue, "false")
	}

	if flag.Usage == "" {
		t.Error("--force flag should have usage description")
	}
}
