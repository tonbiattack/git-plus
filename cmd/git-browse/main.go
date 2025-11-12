package main

import (
	"fmt"
	"os"
	"os/exec"
)

// main はリポジトリをブラウザで開くメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. gh repo view --web を実行してブラウザでリポジトリを開く
//
// 使用する主なコマンド:
//  - gh repo view --web: リポジトリをデフォルトブラウザで開く
//
// 終了コード:
//   - 0: 正常終了
//   - 1: エラー発生
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
	}

	// gh repo view --web を実行
	cmd := exec.Command("gh", "repo", "view", "--web")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: リポジトリをブラウザで開けませんでした: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n注意事項:")
		fmt.Fprintln(os.Stderr, "  - GitHub CLI (gh) がインストールされている必要があります")
		fmt.Fprintln(os.Stderr, "  - gh auth login でログイン済みである必要があります")
		fmt.Fprintln(os.Stderr, "  - リポジトリがGitHubでホストされている必要があります")
		os.Exit(1)
	}
}

// printHelp はコマンドのヘルプメッセージを表示する
func printHelp() {
	help := `git browse - リポジトリをブラウザで開く

使い方:
  git browse

説明:
  現在のリポジトリをデフォルトのウェブブラウザで開きます。
  リポジトリの概要を素早く確認したい場合に便利です。
  CLIでの操作と詳細なウェブビューをスムーズに連携できます。

オプション:
  -h, --help            このヘルプを表示

使用例:
  git browse            # 現在のリポジトリをブラウザで開く

使用する主なコマンド:
  - gh repo view --web: リポジトリをブラウザで開く

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
  gh auth login
`
	fmt.Print(help)
}
