/*
Package cmd は git の拡張コマンド各種コマンドを定義します。

このファイル (amend.go) は、git commit --amend のショートカットコマンドを提供します。
直前のコミットを修正したい場合に便利なコマンドです。

主な機能:
  - git commit --amend の簡易実行
  - すべてのフラグを git commit --amend にそのまま渡す
  - エラーハンドリングとユーザーフレンドリーなメッセージ表示

使用例:
  git amend                # エディタを開いて直前のコミットを修正
  git amend --no-edit      # コミットメッセージを変更せずに修正
  git amend --reset-author # 作成者情報をリセット
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// amendCmd は直前のコミットを修正するコマンドです。
// git commit --amend のショートカットとして機能し、
// 渡されたすべての引数をそのまま git commit --amend に転送します。
var amendCmd = &cobra.Command{
	Use:   "amend",
	Short: "直前のコミットを修正",
	Long: `git commit --amend のショートカットです。
直前のコミットを修正します。引数はそのまま git commit --amend に渡されます。`,
	Example: `  git amend                # 直前のコミットを修正（エディタを開く）
  git amend --no-edit      # コミットメッセージを変更せずに修正
  git amend --reset-author # 作成者情報をリセット`,
	DisableFlagParsing: true, // git commit --amend のフラグをそのまま渡すため
	RunE: func(cmd *cobra.Command, args []string) error {
		// git commit --amend に渡す引数を構築
		// 例: ["commit", "--amend", "--no-edit"] のような配列を作成
		gitArgs := append([]string{"commit", "--amend"}, args...)

		// git commit --amend コマンドを実行
		// 標準入出力を接続することで、エディタの起動やユーザー入力を可能にする
		if err := gitcmd.RunWithIO(gitArgs...); err != nil {
			// 終了コード 1 の場合は Git が正常にエラーを処理しているので、
			// そのまま終了コード 1 で終了する
			if gitcmd.IsExitError(err, 1) {
				os.Exit(1)
			}
			// その他のエラーの場合は詳細なエラーメッセージを返す
			return fmt.Errorf("git commit --amend の実行に失敗しました: %w", err)
		}
		return nil
	},
}

// init はコマンドの初期化を行います。
// amendCmd を rootCmd に登録することで、CLI から実行可能にします。
func init() {
	rootCmd.AddCommand(amendCmd)
}
