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
