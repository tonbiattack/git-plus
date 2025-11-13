package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-plus",
	Short: "Git の日常操作を少しだけ楽にするためのカスタムコマンド集",
	Long: `git-plus は Git の日常操作をより簡単にするためのカスタムコマンド集です。

ブランチ管理、コミット操作、PR管理、統計分析など、
様々な便利な機能を提供します。`,
}

// Execute は rootCmd を実行します
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
