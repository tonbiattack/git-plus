package main

import (
	"fmt"
	"os"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// main は直前のコミットを修正するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. git commit --amend コマンドを実行
//  3. 標準入出力をパススルーしてエディタを開く
//  4. コミット修正の結果を返す
//
// 使用するgitコマンド:
//  - git commit --amend: 直前のコミットを修正
//
// 実装の詳細:
//  - すべてのコマンドライン引数を git commit --amend に渡す
//  - 標準入出力をそのまま接続することで、エディタが正常に動作する
//  - git コマンドの終了コードをそのまま返す
//
// 終了コード:
//  - 0: 正常終了（コミット修正成功）
//  - 1: エラー発生（git commit --amend の実行失敗）
//  - その他: git commit --amend の終了コード
func main() {
	// -h オプションのチェック
	// コマンドライン引数に -h が含まれている場合はヘルプを表示して終了
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// git commit --amend に渡す引数を構築
	// os.Args[1:] に含まれるすべての引数を --amend の後ろに追加
	args := append([]string{"commit", "--amend"}, os.Args[1:]...)

	// git commit --amend を実行
	if err := gitcmd.RunWithIO(args...); err != nil {
		// git コマンドが失敗した場合、終了コードを保持して終了
		if gitcmd.IsExitError(err, 1) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "git commit --amend の実行に失敗しました: %v\n", err)
		os.Exit(1)
	}
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - git commit --amend のショートカットとしての使い方を説明
//  - よく使われるオプションを例示
//  - すべての git commit --amend のオプションが使用可能であることを明記
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
