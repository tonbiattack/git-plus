/*
Package cmd は git-plus の各種コマンドを定義します。

このファイル (browse.go) は、リポジトリをブラウザで開くコマンドを提供します。
GitHub CLI (gh) を使用して、現在のリポジトリをウェブブラウザで開きます。

主な機能:
  - gh コマンドを使用したリポジトリのブラウザ表示
  - エラーハンドリングと要件の明確な表示
  - GitHub でホストされているリポジトリの素早い確認

前提条件:
  - GitHub CLI (gh) がインストールされていること
  - gh auth login でログイン済みであること
  - リポジトリが GitHub でホストされていること

使用例:
  git-plus browse  # 現在のリポジトリをブラウザで開く
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// browseCmd はリポジトリをブラウザで開くコマンドです。
// GitHub CLI (gh) の "gh repo view --web" コマンドを実行し、
// デフォルトのウェブブラウザでリポジトリを開きます。
var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "リポジトリをブラウザで開く",
	Long: `現在のリポジトリをデフォルトのウェブブラウザで開きます。
リポジトリの概要を素早く確認したい場合に便利です。
CLIでの操作と詳細なウェブビューをスムーズに連携できます。

注意事項:
  - GitHub CLI (gh) がインストールされている必要があります
  - gh auth login でログイン済みである必要があります
  - リポジトリがGitHubでホストされている必要があります
  - リモートリポジトリの設定が必要です

GitHub CLI のインストール:
  Windows: winget install --id GitHub.cli
  macOS:   brew install gh
  Linux:   sudo apt install gh

認証方法:
  gh auth login`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// gh repo view --web コマンドを実行してブラウザでリポジトリを開く
		ghCmd := exec.Command("gh", "repo", "view", "--web")
		// 標準入出力を接続して、gh コマンドの出力をそのまま表示する
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		ghCmd.Stdin = os.Stdin

		// コマンドを実行
		if err := ghCmd.Run(); err != nil {
			// エラーが発生した場合、ユーザーに要件を再度表示
			fmt.Fprintln(os.Stderr, "\n注意事項:")
			fmt.Fprintln(os.Stderr, "  - GitHub CLI (gh) がインストールされている必要があります")
			fmt.Fprintln(os.Stderr, "  - gh auth login でログイン済みである必要があります")
			fmt.Fprintln(os.Stderr, "  - リポジトリがGitHubでホストされている必要があります")
			return fmt.Errorf("リポジトリをブラウザで開けませんでした: %w", err)
		}
		return nil
	},
}

// init はコマンドの初期化を行います。
// browseCmd を rootCmd に登録することで、CLI から実行可能にします。
func init() {
	rootCmd.AddCommand(browseCmd)
}
