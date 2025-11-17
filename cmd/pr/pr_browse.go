// ================================================================================
// pr_browse.go
// ================================================================================
// このファイルは git の拡張コマンド pr-browse コマンドを実装しています。
// pr パッケージはPR関連のコマンドを提供します。
//
// 【概要】
// pr-browse コマンドは、GitHub CLI の `gh pr view --web` をラップして、
// Git コマンドとしてPRをブラウザで開けるようにします。
//
// 【主な機能】
// - PRをデフォルトブラウザで開く
// - PR番号を指定して特定のPRを開く
// - 引数なしの場合は現在のブランチのPRを開く
//
// 【使用例】
//   git pr-browse          # 現在のブランチのPRをブラウザで開く
//   git pr-browse 123      # PR #123 をブラウザで開く
//
// 【内部仕様】
// - GitHub CLI (gh) の gh pr view --web コマンドをラップ
// - PR番号が指定された場合はそのPRを開く
// - 指定がない場合は現在のブランチに関連するPRを開く
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package pr

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
)

// prBrowseCmd は pr-browse コマンドの定義です。
// gh pr view --web をラップして Git コマンドとしてPRをブラウザで開けるようにします。
var prBrowseCmd = &cobra.Command{
	Use:   "pr-browse [PR番号]",
	Short: "プルリクエストをブラウザで開く（gh pr view --web のラッパー）",
	Long: `GitHub CLI の gh pr view --web をラップして、Git コマンドとしてPRをブラウザで開けるようにします。

PR番号を指定すると、そのPRをブラウザで開きます。
引数なしで実行すると、現在のブランチに関連するPRを開きます。

注意事項:
  - GitHub CLI (gh) がインストールされている必要があります
  - gh auth login でログイン済みである必要があります
  - リポジトリがGitHubでホストされている必要があります

GitHub CLI のインストール:
  Windows: winget install --id GitHub.cli
  macOS:   brew install gh
  Linux:   sudo apt install gh

認証方法:
  gh auth login`,
	Example: `  git pr-browse          # 現在のブランチのPRをブラウザで開く
  git pr-browse 123      # PR #123 をブラウザで開く
  git pr-browse 456      # PR #456 をブラウザで開く`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLI() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// gh pr view --web コマンドを構築
		ghArgs := []string{"pr", "view", "--web"}

		// PR番号が指定されている場合は追加
		if len(args) > 0 {
			ghArgs = append(ghArgs, args[0])
		}

		// gh コマンドを実行
		ghCmd := exec.Command("gh", ghArgs...)
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		ghCmd.Stdin = os.Stdin

		if err := ghCmd.Run(); err != nil {
			// エラーが発生した場合、ユーザーに要件を再度表示
			fmt.Fprintln(os.Stderr, "\n注意事項:")
			fmt.Fprintln(os.Stderr, "  - GitHub CLI (gh) がインストールされている必要があります")
			fmt.Fprintln(os.Stderr, "  - gh auth login でログイン済みである必要があります")
			fmt.Fprintln(os.Stderr, "  - リポジトリがGitHubでホストされている必要があります")
			if len(args) > 0 {
				fmt.Fprintf(os.Stderr, "  - PR #%s が存在することを確認してください\n", args[0])
			} else {
				fmt.Fprintln(os.Stderr, "  - 現在のブランチにPRが存在することを確認してください")
			}
			return fmt.Errorf("PRをブラウザで開けませんでした: %w", err)
		}

		return nil
	},
}

// init は pr-browse コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	cmd.RootCmd.AddCommand(prBrowseCmd)
}
