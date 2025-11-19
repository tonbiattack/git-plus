// ================================================================================
// abort.go
// ================================================================================
// このファイルは git の拡張コマンド abort を実装しています。
//
// 【概要】
// 進行中の Git 操作（rebase / merge / cherry-pick / revert）を安全に中止します。
// 引数を指定しない場合は現在の状態を判定し、該当する操作を自動で選択します。
//
// 【使用例】
//
//	git abort             # 状態から自動検出して中止
//	git abort merge       # マージを強制的に中止
//	git abort rebase      # リベースを強制的に中止
//
// ================================================================================
package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// abortCmd は進行中のGit操作を中止するコマンドです
var abortCmd = &cobra.Command{
	Use:   "abort [merge|rebase|cherry-pick|revert]",
	Short: "進行中のGit操作を安全に中止",
	Long: `進行中の rebase / merge / cherry-pick / revert を安全に中止します。

引数を指定しない場合は現在の状態を判定して自動的に操作を選択します。`,
	Example: `  git abort           # 自動検出して中止
  git abort merge   # マージを中止
  git abort rebase  # リベースを中止`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAbortCommand,
}

// runAbortCommand は abort コマンドのメイン処理です
func runAbortCommand(_ *cobra.Command, args []string) error {
	var (
		operation string
		err       error
	)

	if len(args) > 0 {
		operation, err = normalizeAbortOperation(args[0])
		if err != nil {
			return err
		}
	} else {
		operation, err = detectAbortOperation()
		if err != nil {
			return err
		}
	}

	label := abortOperationLabel(operation)
	fmt.Printf("%sを中止します...\n", label)

	if err := abortOperation(operation); err != nil {
		return fmt.Errorf("%sの中止に失敗しました: %w", label, err)
	}

	fmt.Println("中止が完了しました。")
	return nil
}

// normalizeAbortOperation はユーザー入力をサポートする操作名に変換します
func normalizeAbortOperation(op string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(op))
	normalized = strings.ReplaceAll(normalized, "_", "-")

	switch normalized {
	case "merge":
		return "merge", nil
	case "rebase":
		return "rebase", nil
	case "cherry", "cherry-pick", "cherrypick":
		return "cherry-pick", nil
	case "revert":
		return "revert", nil
	default:
		return "", fmt.Errorf("サポートされていない操作です: %s", op)
	}
}

// detectAbortOperation は現在のGitディレクトリから進行中の操作を判定します
func detectAbortOperation() (string, error) {
	gitDir, err := getGitDir()
	if err != nil {
		return "", err
	}

	rebaseDirs := []string{"rebase-apply", "rebase-merge"}
	for _, dir := range rebaseDirs {
		if pathExists(filepath.Join(gitDir, dir)) {
			return "rebase", nil
		}
	}

	if pathExists(filepath.Join(gitDir, "CHERRY_PICK_HEAD")) {
		return "cherry-pick", nil
	}

	if pathExists(filepath.Join(gitDir, "REVERT_HEAD")) {
		return "revert", nil
	}

	if pathExists(filepath.Join(gitDir, "MERGE_HEAD")) {
		return "merge", nil
	}

	return "", fmt.Errorf("中止できる操作が検出されませんでした。引数で操作を指定してください")
}

// abortOperation は指定された操作を実際に中止します
func abortOperation(operation string) error {
	switch operation {
	case "merge":
		return gitcmd.RunWithIO("merge", "--abort")
	case "rebase":
		return gitcmd.RunWithIO("rebase", "--abort")
	case "cherry-pick":
		return gitcmd.RunWithIO("cherry-pick", "--abort")
	case "revert":
		return gitcmd.RunWithIO("revert", "--abort")
	default:
		return fmt.Errorf("未対応の操作です: %s", operation)
	}
}

// abortOperationLabel は日本語の表示名を返します
func abortOperationLabel(operation string) string {
	switch operation {
	case "merge":
		return "マージ"
	case "rebase":
		return "リベース"
	case "cherry-pick":
		return "チェリーピック"
	case "revert":
		return "リバート"
	default:
		return operation
	}
}

// getGitDir は現在のリポジトリの .git ディレクトリへの絶対パスを返します
func getGitDir() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("Gitディレクトリの取得に失敗しました: %w", err)
	}

	dir := strings.TrimSpace(string(output))
	if filepath.IsAbs(dir) {
		return dir, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
	}

	return filepath.Join(cwd, dir), nil
}

// pathExists はファイルまたはディレクトリの存在を確認します
func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	cmd.RootCmd.AddCommand(abortCmd)
}
