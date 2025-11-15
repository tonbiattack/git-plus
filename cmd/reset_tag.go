/*
Package cmd は git の拡張コマンド各種コマンドを定義します。

このファイル (reset_tag.go) は、Git タグをリセットして再作成するコマンドを提供します。
既存のタグをローカルとリモートから削除し、最新のコミットに同名のタグを再作成します。

主な機能:
  - ローカルタグの削除
  - リモートタグの削除
  - 最新コミットへのタグの再作成
  - リモートへのタグのプッシュ

使用例:
  git reset-tag v1.2.3  # v1.2.3 タグをリセット
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// resetTagCmd は指定されたタグをリセットして再作成するコマンドです。
// ローカルとリモートのタグを削除してから、最新のコミットに同じ名前のタグを作成します。
var resetTagCmd = &cobra.Command{
	Use:   "reset-tag <タグ名>",
	Short: "タグをリセットして再作成",
	Long: `指定したタグをローカルとリモートから削除し、
最新コミットに同名のタグを再作成してリモートにプッシュします。`,
	Example: `  git reset-tag v1.2.3    # v1.2.3 タグをリセット`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tagName := args[0]

		// ローカルタグ削除（既に存在しない場合があるためエラーは無視）
		// エラーが発生しても処理を継続する
		if err := runGitCommandIgnoreError("tag", "-d", tagName); err != nil {
			fmt.Printf("ローカルタグの削除で予期せぬエラーが発生しました: %v\n", err)
		}

		// リモートタグ削除（存在しないこともあるので警告として扱う）
		// エラーが発生しても処理を継続する
		if err := runGitCommandIgnoreError("push", "--delete", "origin", tagName); err != nil {
			fmt.Printf("リモートタグの削除に失敗しました: %v\n", err)
		}

		// 最新コミットにタグを再付与
		if err := gitcmd.RunWithIO("tag", tagName); err != nil {
			return fmt.Errorf("タグの再作成に失敗しました: %w", err)
		}

		// リモートにタグをプッシュ
		if err := gitcmd.RunWithIO("push", "origin", tagName); err != nil {
			return fmt.Errorf("タグのプッシュに失敗しました: %w", err)
		}

		fmt.Printf("タグ %s をリセットして再作成しました。\n", tagName)
		return nil
	},
}

// runGitCommandIgnoreError は Git コマンドを実行しますが、エラーを無視します。
// タグの削除など、既に存在しない可能性がある操作に使用します。
//
// パラメータ:
//   args: Git コマンドの引数
//
// 戻り値:
//   常に nil を返す（エラーは無視される）
func runGitCommandIgnoreError(args ...string) error {
	err := gitcmd.RunWithIO(args...)
	if err != nil {
		// エラーを無視
		return nil
	}
	return nil
}

// init はコマンドの初期化を行います。
// resetTagCmd を rootCmd に登録することで、CLI から実行可能にします。
func init() {
	rootCmd.AddCommand(resetTagCmd)
}
