// ================================================================================
// Package ui - ユーザーインタラクションユーティリティ
// ================================================================================
// このパッケージは、ユーザーとの対話的な操作を提供するユーティリティ関数を提供します。
//
// 提供する機能:
// - Confirm(): ユーザーに確認プロンプトを表示してyes/noの回答を取得
//
// 使用目的:
// すべてのサブコマンドで共通して使用するユーザーインタラクション機能を一元化し、
// 一貫したUIとコードの重複を避けます。
//
// 設計思想:
// - シンプルなインターフェース: 簡潔な関数シグネチャ
// - 安全性: デフォルト値の設定により、安全な動作を保証
// - 柔軟性: defaultYesパラメータで肯定/否定どちらをデフォルトにするか選択可能
// ================================================================================
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm はユーザーに確認プロンプトを表示し、yes/noの回答を得る
//
// パラメータ:
//   - prompt: 確認メッセージ（例: "タグを作成しますか？"）
//   - defaultYes: trueの場合、Enterキーでyesと判定する (Y/n)
//                falseの場合、Enterキーでnoと判定する (y/N)
//
// 戻り値:
//   - bool: trueの場合yes、falseの場合no
//
// 使用例:
//
//	// 通常の確認（Enterでyes）
//	if !ui.Confirm("タグを作成しますか？", true) {
//	    fmt.Println("キャンセルしました。")
//	    return
//	}
//
//	// 破壊的操作の確認（Enterでno、安全優先）
//	if !ui.Confirm("すべてのブランチを削除しますか？", false) {
//	    fmt.Println("キャンセルしました。")
//	    return
//	}
//
// 判定ロジック:
//   - 空入力（Enter）: defaultYesの値を返す
//   - "y", "yes" (大文字小文字不問): true
//   - "n", "no" (大文字小文字不問): false
//   - その他: false
func Confirm(prompt string, defaultYes bool) bool {
	// プロンプト表記を決定
	yn := "(Y/n)"
	if !defaultYes {
		yn = "(y/N)"
	}
	fmt.Printf("%s %s: ", prompt, yn)

	// 入力を読み取る
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// 入力エラーの場合はデフォルト値を返す
		return defaultYes
	}

	// 入力を正規化（小文字化、前後空白削除）
	input = strings.TrimSpace(strings.ToLower(input))

	// 空入力の場合はデフォルト値を返す
	if input == "" {
		return defaultYes
	}

	// yes/noの判定
	return input == "y" || input == "yes"
}

// NormalizeNumberInput は数字入力を正規化し、全角数字を半角数字に変換する
//
// パラメータ:
//   - input: ユーザー入力文字列
//
// 戻り値:
//   - string: 正規化された入力（前後の空白を削除し、全角数字を半角数字に変換）
//
// 変換ルール:
//   - １ → 1
//   - ２ → 2
//   - ３ → 3
//   - ４ → 4
//   - ５ → 5
//   - ６ → 6
//   - ７ → 7
//   - ８ → 8
//   - ９ → 9
//   - ０ → 0
//
// 使用例:
//
//	input := ui.NormalizeNumberInput(rawInput)
//	selection, err := strconv.Atoi(input)
func NormalizeNumberInput(input string) string {
	// 前後の空白を削除
	normalized := strings.TrimSpace(input)

	// 全角数字を半角数字に変換
	replacer := strings.NewReplacer(
		"１", "1",
		"２", "2",
		"３", "3",
		"４", "4",
		"５", "5",
		"６", "6",
		"７", "7",
		"８", "8",
		"９", "9",
		"０", "0",
	)

	return replacer.Replace(normalized)
}
