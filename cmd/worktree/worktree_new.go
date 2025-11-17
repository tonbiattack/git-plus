// ================================================================================
// worktree_new.go
// ================================================================================
// このファイルは git の拡張コマンド worktree-new を実装しています。
//
// 【概要】
// worktree-new コマンドは、新しいブランチを git worktree として別ディレクトリに作成し、
// 必要に応じてVSCodeを開く機能を提供します。
//
// 【主な機能】
// - 新しいブランチを worktree として別ディレクトリに作成
// - ディレクトリ名は自動的に生成（feature/xxx → ../repo-feature-xxx）
// - --no-code フラグでVSCodeを開かないことも可能
// - --base フラグでベースブランチを指定可能
//
// 【使用例】
//   git worktree-new feature/xxx           # ../repo-feature-xxx に作成してVSCodeを開く
//   git worktree-new feature/xxx --no-code # VSCodeを開かない
//   git worktree-new feature/xxx --base develop  # developブランチから作成
// ================================================================================

package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var (
	noCode     bool
	baseBranch string
)

var worktreeNewCmd = &cobra.Command{
	Use:   "worktree-new <branch-name>",
	Short: "新しいブランチを worktree として別ディレクトリに作成",
	Long: `新しいブランチを git worktree として別ディレクトリに作成し、VSCodeを開きます。

ディレクトリ名は自動的に生成されます:
  feature/xxx → ../repo-feature-xxx
  bugfix/abc  → ../repo-bugfix-abc

これにより、複数のタスクを並行して作業することが容易になります。`,
	Example: `  git worktree-new feature/new-login
  git worktree-new bugfix/issue-123 --no-code
  git worktree-new feature/api-v2 --base develop`,
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		branchName := args[0]

		// リポジトリのルートディレクトリを取得
		repoRoot, err := getRepoRoot()
		if err != nil {
			return fmt.Errorf("リポジトリのルートディレクトリを取得できませんでした: %w", err)
		}

		// リポジトリ名を取得
		repoName := filepath.Base(repoRoot)

		// worktree用のディレクトリ名を生成
		// feature/xxx → repo-feature-xxx
		sanitizedBranch := strings.ReplaceAll(branchName, "/", "-")
		worktreeDirName := fmt.Sprintf("%s-%s", repoName, sanitizedBranch)
		worktreePath := filepath.Join(filepath.Dir(repoRoot), worktreeDirName)

		// ディレクトリが既に存在するかチェック
		if _, err := os.Stat(worktreePath); err == nil {
			return fmt.Errorf("ディレクトリが既に存在します: %s", worktreePath)
		}

		// ブランチが既に存在するかチェック
		branchExists, err := checkBranchExists(branchName)
		if err != nil {
			return fmt.Errorf("ブランチの確認に失敗しました: %w", err)
		}

		fmt.Printf("Worktree を作成しています...\n")
		fmt.Printf("  ブランチ: %s\n", branchName)
		fmt.Printf("  パス: %s\n", worktreePath)

		// worktree を作成
		if branchExists {
			// 既存のブランチを使用
			fmt.Printf("  既存のブランチを使用します\n")
			if err := gitcmd.RunWithIO("worktree", "add", worktreePath, branchName); err != nil {
				return fmt.Errorf("worktree の作成に失敗しました: %w", err)
			}
		} else {
			// 新しいブランチを作成
			if baseBranch != "" {
				fmt.Printf("  ベースブランチ: %s\n", baseBranch)
				if err := gitcmd.RunWithIO("worktree", "add", "-b", branchName, worktreePath, baseBranch); err != nil {
					return fmt.Errorf("worktree の作成に失敗しました: %w", err)
				}
			} else {
				if err := gitcmd.RunWithIO("worktree", "add", "-b", branchName, worktreePath); err != nil {
					return fmt.Errorf("worktree の作成に失敗しました: %w", err)
				}
			}
		}

		fmt.Printf("\n✓ Worktree を作成しました: %s\n", worktreePath)

		// VSCodeを開く（--no-code フラグがない場合）
		if !noCode {
			fmt.Printf("VSCode を開いています...\n")
			if err := openVSCode(worktreePath); err != nil {
				fmt.Printf("警告: VSCode を開けませんでした: %v\n", err)
				fmt.Printf("手動で開いてください: code %s\n", worktreePath)
			} else {
				fmt.Printf("✓ VSCode を開きました\n")
			}
		}

		return nil
	},
}

// getRepoRoot は現在のリポジトリのルートディレクトリを取得します
func getRepoRoot() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkBranchExists は指定されたブランチが存在するかチェックします
func checkBranchExists(branchName string) (bool, error) {
	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
	if err == nil {
		return true, nil
	}
	if gitcmd.IsExitError(err, 1) {
		return false, nil
	}
	return false, err
}

// openVSCode は指定されたディレクトリでVSCodeを開きます
func openVSCode(path string) error {
	cmd := exec.Command("code", path)
	return cmd.Start()
}

func init() {
	worktreeNewCmd.Flags().BoolVar(&noCode, "no-code", false, "VSCode を開かない")
	worktreeNewCmd.Flags().StringVar(&baseBranch, "base", "", "ベースブランチを指定（デフォルトは現在のブランチ）")
	cmd.RootCmd.AddCommand(worktreeNewCmd)
}
