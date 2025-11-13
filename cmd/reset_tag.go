package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var resetTagCmd = &cobra.Command{
	Use:   "reset-tag <タグ名>",
	Short: "タグをリセットして再作成",
	Long: `指定したタグをローカルとリモートから削除し、
最新コミットに同名のタグを再作成してリモートにプッシュします。`,
	Example: `  git-plus reset-tag v1.2.3    # v1.2.3 タグをリセット`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tagName := args[0]

		// ローカルタグ削除（既に存在しない場合があるためエラーは無視）
		if err := runGitCommandIgnoreError("tag", "-d", tagName); err != nil {
			fmt.Printf("ローカルタグの削除で予期せぬエラーが発生しました: %v\n", err)
		}

		// リモートタグ削除（存在しないこともあるので警告として扱う）
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

func runGitCommandIgnoreError(args ...string) error {
	err := gitcmd.RunWithIO(args...)
	if err != nil {
		// エラーを無視
		return nil
	}
	return nil
}

func init() {
	rootCmd.AddCommand(resetTagCmd)
}
