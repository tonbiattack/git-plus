package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

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
		if err := gitcmd.RunWithIO("reset", "--soft", "HEAD^"); err != nil {
			return fmt.Errorf("コミットの取り消しに失敗しました: %w", err)
		}
		fmt.Println("最後のコミットを取り消しました（変更は残っています）")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoLastCommitCmd)
}
