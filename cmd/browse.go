package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

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
		ghCmd := exec.Command("gh", "repo", "view", "--web")
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		ghCmd.Stdin = os.Stdin

		if err := ghCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "\n注意事項:")
			fmt.Fprintln(os.Stderr, "  - GitHub CLI (gh) がインストールされている必要があります")
			fmt.Fprintln(os.Stderr, "  - gh auth login でログイン済みである必要があります")
			fmt.Fprintln(os.Stderr, "  - リポジトリがGitHubでホストされている必要があります")
			return fmt.Errorf("リポジトリをブラウザで開けませんでした: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
