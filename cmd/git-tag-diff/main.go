package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// main は2つのタグ間のコミット差分を取得してファイルに出力するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. コマンドライン引数から古いタグと新しいタグを取得
//  3. 出力ファイル名を自動生成
//  4. 両方のタグの存在を確認
//  5. git log で2つのタグ間の差分を取得
//  6. 結果をファイルに保存
//  7. サマリー情報を表示
//
// 使用するgitコマンド:
//  - git rev-parse --verify <タグ名>: タグの存在確認
//  - git log <古いタグ>..<新しいタグ> --no-merges: タグ間のコミット差分を取得
//    --pretty=format:- %s (%an, %ad): コミットメッセージ、作成者名、日付を整形
//    --date=iso: ISO 8601形式の日付
//
// 実装の詳細:
//  - マージコミットは自動的に除外（--no-merges）
//  - 出力ファイル名: "tag_diff_<古いタグ>_to_<新しいタグ>.txt"
//  - 差分がない場合は終了コード0で正常終了
//
// 終了コード:
//  - 0: 正常終了（差分取得成功、差分なし）
//  - 1: エラー発生（引数不足、タグ不存在、git log失敗、ファイル書き込み失敗）
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	if len(os.Args) < 3 {
		fmt.Println("使用方法: git tag-diff <古いタグ> <新しいタグ>")
		fmt.Println("例: git tag-diff V4.2.00.00 V4.3.00.00")
		os.Exit(1)
	}

	oldTag := os.Args[1]
	newTag := os.Args[2]

	// 出力ファイル名の自動生成
	outputFile := fmt.Sprintf("tag_diff_%s_to_%s.txt", oldTag, newTag)

	// タグの存在確認
	if err := validateTag(oldTag); err != nil {
		fmt.Printf("エラー: タグ '%s' が存在しません\n", oldTag)
		os.Exit(1)
	}
	if err := validateTag(newTag); err != nil {
		fmt.Printf("エラー: タグ '%s' が存在しません\n", newTag)
		os.Exit(1)
	}

	// git logコマンドの実行
	tagRange := fmt.Sprintf("%s..%s", oldTag, newTag)
	cmd := exec.Command("git", "log", tagRange, "--no-merges", "--pretty=format:- %s (%an, %ad)", "--date=iso")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("git logの実行に失敗しました: %s\n", string(exitErr.Stderr))
		} else {
			fmt.Printf("git logの実行に失敗しました: %v\n", err)
		}
		os.Exit(1)
	}

	// 出力が空の場合
	if len(output) == 0 {
		fmt.Printf("タグ %s と %s の間に差分はありません。\n", oldTag, newTag)
		os.Exit(0)
	}

	// ファイルに書き込み
	absPath, err := filepath.Abs(outputFile)
	if err != nil {
		fmt.Printf("ファイルパスの取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		fmt.Printf("ファイルへの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// サマリー表示
	commits := strings.Split(string(output), "\n")
	fmt.Printf("✓ タグ %s と %s の差分を %s に出力しました。\n", oldTag, newTag, absPath)
	fmt.Printf("  コミット数: %d\n", len(commits))
}

// validateTag はタグが存在するかを確認する
//
// パラメータ:
//  - tag: 確認するタグ名。例: "v1.0.0"
//
// 戻り値:
//  - error: タグが存在しない場合または確認に失敗した場合のエラー
//
// 使用するgitコマンド:
//  - git rev-parse --verify <タグ名>: タグの存在を確認
//
// 実装の詳細:
//  - タグが存在する場合はエラーなし（nil）を返す
//  - タグが存在しない場合はエラーを返す
func validateTag(tag string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", tag)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - タグ間の差分取得機能の使い方を説明
//  - 出力形式とファイル名の規則を明記
//  - マージコミットが除外されることを明示
func printHelp() {
	help := `git tag-diff - 2つのタグ間の差分を取得

使い方:
  git tag-diff <古いタグ> <新しいタグ>

説明:
  2つのタグ間のコミット差分を取得し、ファイルに出力します。
  Mergeコミットは自動的に除外されます。
  出力形式: - コミットメッセージ (作成者名, 日付)
  
オプション:
  -h                    このヘルプを表示

例:
  git tag-diff V4.2.00.00 V4.3.00.00

出力:
  tag_diff_<古いタグ>_to_<新しいタグ>.txt に保存されます。
`
	fmt.Print(help)
}
