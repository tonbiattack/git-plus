/*
Package commit は git の拡張コマンドのうち、コミット関連のコマンドを定義します。

このファイル (undo_last_commit.go) は、直前のコミットを取り消すコマンドを提供します。
git reset --soft HEAD^ のショートカットとして機能し、
コミットのみを取り消して変更内容は保持します。

主な機能:
  - 直前のコミットの取り消し（git reset --soft HEAD^）
  - 変更内容とステージング状態の保持
  - コミットメッセージの修正に便利

使用例:
  git undo-last-commit  # 直前のコミットを取り消す
*/
package commit

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// undoLastCommitCmd は直前のコミットを取り消すコマンドです。
// git reset --soft HEAD^ を実行し、コミットのみを取り消して
// 変更内容はステージングエリアに残します。
var undoLastCommitCmd = &cobra.Command{
	Use:   "undo-last-commit",
	Short: "直前のコミットを取り消し",
	Long: `git reset --soft HEAD^ を実行し、直前のコミットだけを取り消します。
作業ツリーとステージング内容はそのまま残ります。

注意:
  - コミットは取り消されますが、変更内容は保持されます
  - ステージングエリアに変更が残ります
  - コミットメッセージを修正したい時などに便利です`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// git reset --soft HEAD^ を実行して直前のコミットのみを取り消す
		// --soft オプションにより、変更内容はステージングエリアに保持される
		if err := gitcmd.RunWithIO("reset", "--soft", "HEAD^"); err != nil {
			return fmt.Errorf("コミットの取り消しに失敗しました: %w", err)
		}
		fmt.Println("最後のコミットを取り消しました（変更は残っています）")
		return nil
	},
}

// init はコマンドの初期化を行います。
// undoLastCommitCmd を RootCmd に登録することで、CLI から実行可能にします。
func init() {
	cmd.RootCmd.AddCommand(undoLastCommitCmd)
}
