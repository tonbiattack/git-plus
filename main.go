// ================================================================================
// Git Plus - メインエントリーポイント
// ================================================================================
// このファイルは、Git Plusアプリケーションのメインエントリーポイントです。
// 実行ファイル名からサブコマンドを自動的に推測する機能を提供します。
//
// 仕組み:
// - 実行ファイル名が "git-xxx" の場合、"xxx" をサブコマンドとして扱います
// - これにより、同一バイナリから複数のコマンドを実行できます
// - Linux/macOS: シンボリックリンクを使用 (git-newbranch → git-plus)
// - Windows: 実行ファイルのコピーを使用 (git-newbranch.exe)
//
// 例:
// - git-newbranch を実行 → "newbranch" サブコマンドが実行される
// - git-amend を実行 → "amend" サブコマンドが実行される
// - git-plus newbranch を直接実行することも可能
// ================================================================================
package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tonbiattack/git-plus/cmd"
)

// main は、Git Plusアプリケーションのエントリーポイントです。
// 実行ファイル名を解析してサブコマンドを推測し、Cobraコマンドフレームワークに処理を渡します。
func main() {
	// 実行ファイル名からサブコマンドを推測
	// os.Args[0] には実行ファイルのフルパスが含まれるため、
	// filepath.Base() を使ってファイル名のみを取得
	// 例: "/home/user/bin/git-newbranch" → "git-newbranch"
	execName := filepath.Base(os.Args[0])

	// "git-" で始まる場合、それをサブコマンドとして扱う
	// これにより、git-xxx という名前の実行ファイルまたはシンボリックリンクが
	// 自動的に対応するサブコマンドを実行します
	if strings.HasPrefix(execName, "git-") {
		// "git-newbranch" → "newbranch"
		// プレフィックス "git-" を削除してサブコマンド名を取得
		subCommand := strings.TrimPrefix(execName, "git-")

		// os.Argsを書き換えて、サブコマンドを挿入
		// これにより、Cobraフレームワークが適切にサブコマンドを認識できます
		//
		// 変換例:
		// 実行前: ["git-newbranch", "feature-xxx"]
		// 実行後: ["git-newbranch", "newbranch", "feature-xxx"]
		//
		// この変換により、Cobraは "newbranch" サブコマンドとして処理します
		newArgs := make([]string, len(os.Args)+1)
		newArgs[0] = os.Args[0]          // 実行ファイル名を保持
		newArgs[1] = subCommand          // サブコマンドを挿入
		copy(newArgs[2:], os.Args[1:])   // 元の引数をコピー
		os.Args = newArgs
	}

	// Cobraコマンドフレームワークに制御を渡して、
	// 対応するサブコマンドを実行します
	cmd.Execute()
}
