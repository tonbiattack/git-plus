// ================================================================================
// Git Plus - Cobraルートコマンド定義
// ================================================================================
// このファイルは、Cobraコマンドフレームワークのルートコマンドを定義します。
//
// Cobraフレームワークについて:
// - Cobraは、Go言語で最も人気のあるCLIフレームワークの一つです
// - kubectl、docker、githubなど多くの有名なCLIツールで使用されています
// - サブコマンド、フラグ、引数の解析を簡単に実装できます
//
// このファイルの役割:
// - rootCmd: すべてのサブコマンドの親となるルートコマンド
// - Execute(): main.goから呼び出されるエントリーポイント
// - 各サブコマンドは init() 関数で rootCmd に登録されます
// ================================================================================
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd は、Git Plusのルートコマンドを定義します。
// すべてのサブコマンド（newbranch, amend, squashなど）は、
// 各サブパッケージのinit()関数でこのRootCmdに登録されます。
//
// Cobraのコマンド構造:
// RootCmd (git plus)
//   ├── branch/ (newbranch, back, recent, sync, delete-local-branches)
//   ├── tag/ (reset-tag, tag-diff, new-tag, etc.)
//   ├── commit/ (amend, squash, undo-last-commit, track)
//   ├── stash/ (stash-cleanup, stash-select, pause, resume)
//   ├── pr/ (pr-create-merge, pr-list, pr-merge, pr-checkout)
//   ├── repo/ (create-repository, clone-org, batch-clone, browse, repo-others)
//   ├── issue/ (issue-list, issue-create, issue-edit)
//   ├── release/ (release-notes)
//   └── stats/ (step)
var RootCmd = &cobra.Command{
	Use:   "plus",
	Short: "Git の日常操作を少しだけ楽にするための拡張コマンド集",
	Long: `git plus は Git の日常操作をより簡単にするための拡張コマンド集です。

git のサブコマンドとして動作し、ブランチ管理、コミット操作、
PR管理、統計分析など、様々な便利な機能を提供します。

使用方法: git plus <サブコマンド>`,
}

// Execute は、Cobraのルートコマンドを実行します。
// この関数は、main.goから呼び出されるエントリーポイントです。
//
// 処理フロー:
// 1. main.go で os.Args を解析してサブコマンドを推測
// 2. cmd.Execute() を呼び出し
// 3. Cobraがサブコマンドを解析して対応するコマンドを実行
// 4. エラーが発生した場合は、標準エラー出力に表示して終了
//
// 戻り値:
// エラーが発生した場合は、終了コード1でプロセスを終了します。
func Execute() {
	// Cobraコマンドを実行
	if err := RootCmd.Execute(); err != nil {
		// エラーメッセージを標準エラー出力に表示
		fmt.Fprintln(os.Stderr, err)
		// 終了コード1でプロセスを終了
		os.Exit(1)
	}
}
