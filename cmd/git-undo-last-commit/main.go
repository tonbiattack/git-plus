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

	cmd := exec.Command("git", "reset", "--soft", "HEAD^")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("コミットの取り消しに失敗しました:", err)
		os.Exit(1)
	}

	fmt.Println("最後のコミットを取り消しました（変更は残っています）")
}

func printHelp() {
	help := `git undo-last-commit - 直前のコミットを取り消し

使い方:
  git undo-last-commit

説明:
  git reset --soft HEAD^ を実行し、直前のコミットだけを取り消します。
  作業ツリーとステージング内容はそのまま残ります。

オプション:
  -h                    このヘルプを表示

注意:
  - コミットは取り消されますが、変更内容は保持されます
  - ステージングエリアに変更が残ります
  - コミットメッセージを修正したい時などに便利です
`
	fmt.Print(help)
}
