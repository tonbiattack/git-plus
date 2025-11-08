package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	args := append([]string{"commit", "--amend"}, os.Args[1:]...)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "git commit --amend の実行に失敗しました: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	help := `git amend - 直前のコミットを修正

使い方:
  git amend                # 直前のコミットを修正（エディタを開く）
  git amend --no-edit      # コミットメッセージを変更せずに修正
  git amend --reset-author # 作成者情報をリセット

説明:
  git commit --amend のショートカットです。
  直前のコミットを修正します。引数はそのまま git commit --amend に渡されます。

オプション:
  -h                       このヘルプを表示
  --no-edit                コミットメッセージを変更しない
  --reset-author           作成者情報をリセット
  その他のオプション       git commit --amend のオプションがそのまま使用できます
`
	fmt.Print(help)
}
