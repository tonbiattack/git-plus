package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var amendCmd = &cobra.Command{
	Use:   "amend",
	Short: "直前のコミットを修正",
	Long: `git commit --amend のショートカットです。
直前のコミットを修正します。引数はそのまま git commit --amend に渡されます。`,
	Example: `  git-plus amend                # 直前のコミットを修正（エディタを開く）
  git-plus amend --no-edit      # コミットメッセージを変更せずに修正
  git-plus amend --reset-author # 作成者情報をリセット`,
	DisableFlagParsing: true, // git commit --amend のフラグをそのまま渡すため
	RunE: func(cmd *cobra.Command, args []string) error {
		// git commit --amend に渡す引数を構築
		gitArgs := append([]string{"commit", "--amend"}, args...)

		if err := gitcmd.RunWithIO(gitArgs...); err != nil {
			if gitcmd.IsExitError(err, 1) {
				os.Exit(1)
			}
			return fmt.Errorf("git commit --amend の実行に失敗しました: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(amendCmd)
}
