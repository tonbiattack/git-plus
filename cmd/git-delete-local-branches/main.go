package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// main はマージ済みローカルブランチを削除するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. マージ済みブランチの一覧を取得（保護対象ブランチと現在のブランチを除外）
//  3. 削除対象ブランチを表示して確認を求める
//  4. ユーザー確認後、各ブランチを git branch -d で削除
//  5. 削除結果を表示
//
// 保護されるブランチ:
//  - main, master, develop（プロジェクトの主要ブランチ）
//  - 現在チェックアウト中のブランチ（誤削除を防止）
//
// 終了コード:
//  - 0: 正常終了（削除成功またはキャンセル）
//  - 1: エラー発生（ブランチ取得失敗、削除失敗など）
func main() {
	// -h オプションのチェック
	// コマンドライン引数に -h が含まれている場合はヘルプを表示して終了
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// マージ済みブランチの一覧を取得
	// 保護対象ブランチ（main/master/develop）と現在のブランチは自動的に除外される
	branches, err := mergedBranches()
	if err != nil {
		fmt.Println("マージ済みブランチの取得に失敗しました:", err)
		os.Exit(1)
	}

	// 削除対象のブランチが存在しない場合は終了
	if len(branches) == 0 {
		fmt.Println("削除対象のブランチはありません。")
		return
	}

	// 削除対象のブランチ一覧を表示
	fmt.Println("以下のブランチを削除します:")
	for _, b := range branches {
		fmt.Println(b)
	}

	// ユーザーに削除の確認を求める（破壊的操作なのでEnterでno）
	if !ui.Confirm("本当に削除しますか？", false) {
		fmt.Println("キャンセルしました。")
		return
	}

	// 各ブランチを順番に削除
	// git branch -d を使用することで、未マージのブランチは削除されない
	var deleteErrors bool
	for _, branch := range branches {
		if err := gitcmd.RunWithIO("branch", "-d", branch); err != nil {
			deleteErrors = true
			fmt.Fprintf(os.Stderr, "ブランチ %s の削除に失敗しました: %v\n", branch, err)
		}
	}

	// 削除処理中にエラーが発生した場合は終了コード1で終了
	if deleteErrors {
		fmt.Println("一部のブランチの削除に失敗しました。")
		os.Exit(1)
	}

	fmt.Println("削除しました。")
}

// mergedBranches はマージ済みブランチの一覧を取得する
//
// git branch --merged コマンドを実行し、マージ済みブランチを取得する。
// 以下のブランチは自動的に除外される:
//  - 現在チェックアウト中のブランチ（行頭に * が付いているブランチ）
//  - 保護対象ブランチ（main, master, develop）
//
// 戻り値:
//  - []string: 削除可能なマージ済みブランチのスライス
//  - error: git コマンドの実行エラー、またはスキャンエラー
//
// 実装の詳細:
//  git branch --merged の出力形式:
//    * current-branch   ← 現在のブランチ（*付き）
//      feature-1        ← 通常のブランチ
//      feature-2
//      main             ← 保護対象
//
// この関数は上記の出力から、* 付きと保護対象を除外したブランチ名のみを返す
func mergedBranches() ([]string, error) {
	// git branch --merged でマージ済みブランチの一覧を取得
	output, err := gitcmd.Run("branch", "--merged")
	if err != nil {
		return nil, err
	}

	// 出力を行ごとに処理
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var branches []string
	for scanner.Scan() {
		// 各行の前後の空白を削除
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 現在のブランチ（*が付いている）はスキップ
		// 例: "* enhanced_delete_local_branches" → スキップ
		// これにより、誤って現在作業中のブランチを削除することを防ぐ
		if strings.HasPrefix(line, "*") {
			continue
		}

		// 念のため再度空チェック（通常は不要だが安全のため）
		if line == "" {
			continue
		}

		// 保護対象ブランチ（main/master/develop）はスキップ
		if shouldSkipBranch(line) {
			continue
		}

		// 削除対象ブランチとしてリストに追加
		branches = append(branches, line)
	}

	// スキャン中のエラーチェック
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}

// shouldSkipBranch は指定されたブランチが保護対象かどうかを判定する
//
// プロジェクトの主要ブランチを誤って削除することを防ぐため、
// 以下のブランチ名は保護対象として true を返す:
//  - main: Git のデフォルトメインブランチ
//  - master: 従来の Git のデフォルトブランチ
//  - develop: Git Flow などで使用される開発ブランチ
//
// パラメータ:
//  - branch: チェックするブランチ名
//
// 戻り値:
//  - true: 保護対象ブランチ（削除すべきでない）
//  - false: 通常のブランチ（削除可能）
func shouldSkipBranch(branch string) bool {
	switch branch {
	case "main", "master", "develop":
		return true
	default:
		return false
	}
}

// printHelp はコマンドのヘルプメッセージを表示する
//
// 使い方、説明、オプション、注意事項を含む詳細なヘルプを標準出力に表示する。
// -h オプションが指定された場合や、使い方が分からない場合にユーザーに情報を提供する。
func printHelp() {
	help := `git delete-local-branches - マージ済みのローカルブランチを削除

使い方:
  git delete-local-branches

説明:
  git branch --merged に含まれるマージ済みブランチのうち、
  main / master / develop / 現在のブランチ 以外をまとめて削除します。
  削除前に確認プロンプトが表示されます。

オプション:
  -h                    このヘルプを表示

注意:
  - 削除対象のブランチは削除前に一覧表示されます
  - y または yes を入力すると削除が実行されます
  - 未マージのブランチは削除されません
  - 現在のブランチは自動的に除外されます

実装詳細:
  git branch --merged | egrep -v "(^\*|main|master|develop)" | xargs git branch -d
  と同等の動作をGoで実装しています。
`
	fmt.Print(help)
}
