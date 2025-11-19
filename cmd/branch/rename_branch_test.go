package branch

import (
	"os"
	"strings"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// resetRenameBranchFlags はテスト毎に rename-branch コマンドのフラグ状態を初期化します。
func resetRenameBranchFlags() {
	renamePush = false
	renameDeleteRemote = false
	renameRemoteName = "origin"
}

// TestRenameBranchCmdDefinition はコマンド定義の基本情報を確認します。
func TestRenameBranchCmdDefinition(t *testing.T) {
	if renameBranchCmd.Use != "rename-branch <新しいブランチ名>" {
		t.Errorf("renameBranchCmd.Use = %q, want %q", renameBranchCmd.Use, "rename-branch <新しいブランチ名>")
	}

	if renameBranchCmd.Short == "" {
		t.Error("renameBranchCmd.Short should not be empty")
	}

	if renameBranchCmd.Long == "" {
		t.Error("renameBranchCmd.Long should not be empty")
	}

	if renameBranchCmd.Example == "" {
		t.Error("renameBranchCmd.Example should not be empty")
	}

	if renameBranchCmd.RunE == nil {
		t.Error("renameBranchCmd.RunE should not be nil")
	}
}

// TestRenameBranchCmdFlags はフラグ定義を確認します。
func TestRenameBranchCmdFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"push flag", "push", "false"},
		{"delete-remote flag", "delete-remote", "false"},
		{"remote flag", "remote", "origin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := renameBranchCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("Flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

// TestRenameBranchCmdRegistered は rootCmd に登録されていることを確認します。
func TestRenameBranchCmdRegistered(t *testing.T) {
	found, _, err := cmd.RootCmd.Find([]string{"rename-branch"})
	if err != nil {
		t.Fatalf("rename-branch command not found: %v", err)
	}
	if found == nil {
		t.Fatal("rename-branch command should not be nil")
	}
}

// TestRunRenameBranchCommand_RenamesCurrentBranch はローカルブランチがリネームされることを確認します。
func TestRunRenameBranchCommand_RenamesCurrentBranch(t *testing.T) {
	repo := testutil.NewGitRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial")

	oldBranch := repo.CurrentBranch()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	resetRenameBranchFlags()

	if err := runRenameBranchCommand("feature/renamed"); err != nil {
		t.Fatalf("runRenameBranchCommand returned error: %v", err)
	}

	newBranch := repo.CurrentBranch()
	if newBranch != "feature/renamed" {
		t.Errorf("Branch should be renamed to feature/renamed, got %s", newBranch)
	}

	exists, err := checkBranchExists(oldBranch)
	if err != nil {
		t.Fatalf("checkBranchExists returned error: %v", err)
	}
	if exists {
		t.Errorf("Old branch %s should be removed after rename", oldBranch)
	}
}

// TestRunRenameBranchCommand_TargetExists は新しいブランチ名が既に存在する場合のエラーを確認します。
func TestRunRenameBranchCommand_TargetExists(t *testing.T) {
	repo := testutil.NewGitRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial")
	repo.CreateBranch("feature/exist")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	resetRenameBranchFlags()

	if err := runRenameBranchCommand("feature/exist"); err == nil {
		t.Fatal("runRenameBranchCommand should return error when target branch already exists")
	}
}

// TestRunRenameBranchCommand_DeleteRemoteRequiresPush は --delete-remote に --push が必須であることを確認します。
func TestRunRenameBranchCommand_DeleteRemoteRequiresPush(t *testing.T) {
	repo := testutil.NewGitRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	resetRenameBranchFlags()
	renameDeleteRemote = true

	err = runRenameBranchCommand("feature/renamed")
	if err == nil {
		t.Fatal("runRenameBranchCommand should return error when --delete-remote is used without --push")
	}

	if !strings.Contains(err.Error(), "--delete-remote") {
		t.Errorf("error message = %q, want mention of --delete-remote", err.Error())
	}
}
