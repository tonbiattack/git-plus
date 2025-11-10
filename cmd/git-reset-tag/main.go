package main

import (
	"fmt"
	"os"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// main は指定されたタグをリセットして最新コミットに再作成するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. タグ名のコマンドライン引数チェック
//  3. ローカルのタグを削除（エラーは無視）
//  4. リモート（origin）のタグを削除（エラーは無視）
//  5. 最新コミット（HEAD）に同じ名前のタグを再作成
//  6. リモート（origin）に新しいタグをプッシュ
//
// 使用するgitコマンド:
//  - git tag -d <タグ名>: ローカルタグを削除
//  - git push --delete origin <タグ名>: リモートタグを削除
//  - git tag <タグ名>: 現在のコミットにタグを作成
//  - git push origin <タグ名>: タグをリモートにプッシュ
//
// 実装の詳細:
//  - 削除処理はエラーを無視（タグが存在しない場合があるため）
//  - 再作成とプッシュは厳密にエラーチェック
//  - origin リモートを使用（固定）
//
// 注意事項:
//  - タグの削除と再作成により、他の開発者の環境でコンフリクトが発生する可能性がある
//  - リリースタグなどの重要なタグをリセットする際は十分注意が必要
//
// 終了コード:
//  - 0: 正常終了（タグのリセット成功）
//  - 1: エラー発生（引数不足、タグの再作成失敗、プッシュ失敗）
func main() {
	// -h オプションのチェック
	// コマンドライン引数に -h が含まれている場合はヘルプを表示して終了
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// タグ名の引数チェック
	// 最低1つの引数（タグ名）が必要
	if len(os.Args) < 2 {
		fmt.Println("タグ名を指定してください。")
		os.Exit(1)
	}
	tagName := os.Args[1]

	// ローカルタグ削除（既に存在しない場合があるためエラーは無視）
	// git tag -d を実行してローカルリポジトリからタグを削除
	if err := runGitCommand(true, "tag", "-d", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "ローカルタグの削除で予期せぬエラーが発生しました: %v\n", err)
	}

	// リモートタグ削除（存在しないこともあるので警告として扱う）
	// git push --delete origin を実行してリモートリポジトリからタグを削除
	if err := runGitCommand(true, "push", "--delete", "origin", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "リモートタグの削除に失敗しました: %v\n", err)
	}

	// 最新コミットにタグを再付与
	// git tag を実行して現在の HEAD に新しいタグを作成
	// この処理が失敗した場合は致命的エラーとして扱う
	if err := runGitCommand(false, "tag", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "タグの再作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// リモートにタグをプッシュ
	// git push origin を実行して新しいタグをリモートリポジトリに送信
	// この処理が失敗した場合は致命的エラーとして扱う
	if err := runGitCommand(false, "push", "origin", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "タグのプッシュに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("タグ %s をリセットして再作成しました。\n", tagName)
}

// runGitCommand はgitコマンドを実行するヘルパー関数
//
// パラメータ:
//  - ignoreError: true の場合、エラーを無視して nil を返す（削除処理などで使用）
//  - args: git コマンドに渡す引数（可変長引数）
//
// 戻り値:
//  - error: コマンド実行エラー（ignoreError が true の場合は常に nil）
//
// 使用するgitコマンド:
//  - 可変長引数で指定されたgitコマンド
//
// 実装の詳細:
//  - 標準出力と標準エラー出力をそのまま表示
//  - ignoreError が true の場合、エラーが発生しても nil を返す
//  - ignoreError が false の場合、エラーをそのまま返す
func runGitCommand(ignoreError bool, args ...string) error {
	err := gitcmd.RunWithIO(args...)
	if err != nil {
		// エラーを無視するフラグが立っている場合は nil を返す
		if ignoreError {
			return nil
		}
		return err
	}

	return nil
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - タグのリセット機能の使い方を説明
//  - 処理の流れと注意事項を明記
//  - 使用例を提示
func printHelp() {
	help := `git reset-tag - タグをリセットして再作成

使い方:
  git reset-tag <タグ名>

説明:
  指定したタグをローカルとリモートから削除し、
  最新コミットに同名のタグを再作成してリモートにプッシュします。

オプション:
  -h                    このヘルプを表示

例:
  git reset-tag v1.2.3    # v1.2.3 タグをリセット

注意:
  - ローカルとリモート（origin）のタグが削除されます
  - 最新コミットに新しいタグが作成されます
`
	fmt.Print(help)
}
