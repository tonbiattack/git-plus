package main

import (
	"fmt"
	"os"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// main は直前のコミットを取り消すメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. git reset --soft HEAD^ を実行
//  3. 結果を表示
//
// 使用するgitコマンド:
//  - git reset --soft HEAD^: 直前のコミットを取り消し（変更は保持）
//
// 実装の詳細:
//  - --soft オプションにより、ワークツリーとステージングエリアは変更されない
//  - コミットだけが取り消され、変更内容はステージングエリアに残る
//  - コミットメッセージを修正したい場合などに有用
//
// 注意事項:
//  - リモートにプッシュ済みのコミットを取り消す場合は注意が必要
//  - 複数のコミットを取り消したい場合は git reset --soft HEAD~N を使用
//
// 終了コード:
//  - 0: 正常終了（コミット取り消し成功）
//  - 1: エラー発生（git reset の実行失敗）
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	if err := gitcmd.RunWithIO("reset", "--soft", "HEAD^"); err != nil {
		fmt.Println("コミットの取り消しに失敗しました:", err)
		os.Exit(1)
	}

	fmt.Println("最後のコミットを取り消しました（変更は残っています）")
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - コミット取り消し機能の使い方を説明
//  - git reset --soft の動作を明記
//  - 変更内容が保持されることを強調
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
