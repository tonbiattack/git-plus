/*
Package cmd は git の拡張コマンド各種コマンドを定義します。

このファイル (back.go) は、前のブランチやタグに戻るコマンドを提供します。
git checkout - のショートカットとして機能し、
直前にいたブランチやタグに素早く戻ることができます。

主な機能:
  - 前のブランチ・タグへの切り替え（git checkout -）
  - ブランチやタグ間の素早い移動
  - 作業ブランチとメインブランチの往復、タグの確認などに便利

使用例:
  git back  # 前のブランチ/タグに戻る
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// backCmd は前のブランチやタグに戻るコマンドです。
// git checkout - を実行し、直前にチェックアウトしていたブランチやタグに切り替えます。
var backCmd = &cobra.Command{
	Use:   "back",
	Short: "前のブランチ/タグに戻る",
	Long: `git checkout - を実行し、直前にいたブランチやタグに戻ります。

このコマンドは、ブランチやタグ間を頻繁に移動する際に便利です。
例えば、作業ブランチとメインブランチの往復、タグの確認などに使用できます。

注意:
  - 少なくとも一度はブランチやタグを切り替えている必要があります
  - 最初のチェックアウトでは使用できません`,
	Example: `  git back  # 前のブランチ/タグに戻る
  git back       # git へのエイリアスがあれば使用可能`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// git checkout - を実行して前のブランチ/タグに切り替え
		// "-" は直前にチェックアウトしていたブランチ/タグを示す特殊な記号
		if err := gitcmd.RunWithIO("checkout", "-"); err != nil {
			return fmt.Errorf("前のブランチ/タグへの切り替えに失敗しました: %w", err)
		}
		return nil
	},
}

// init はコマンドの初期化を行います。
// backCmd を rootCmd に登録することで、CLI から実行可能にします。
func init() {
	rootCmd.AddCommand(backCmd)
}
