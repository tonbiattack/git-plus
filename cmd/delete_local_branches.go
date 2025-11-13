package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var deleteLocalBranchesCmd = &cobra.Command{
	Use:   "delete-local-branches",
	Short: "マージ済みのローカルブランチを削除",
	Long: `git branch --merged に含まれるマージ済みブランチのうち、
main / master / develop / 現在のブランチ 以外をまとめて削除します。
削除前に確認プロンプトが表示されます。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// マージ済みブランチの一覧を取得
		branches, err := getMergedBranches()
		if err != nil {
			return fmt.Errorf("マージ済みブランチの取得に失敗しました: %w", err)
		}

		if len(branches) == 0 {
			fmt.Println("削除対象のブランチはありません。")
			return nil
		}

		// 削除対象のブランチ一覧を表示
		fmt.Println("以下のブランチを削除します:")
		for _, b := range branches {
			fmt.Println(b)
		}

		// ユーザーに削除の確認を求める
		if !ui.Confirm("本当に削除しますか？", false) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// 各ブランチを順番に削除
		var deleteErrors bool
		for _, branch := range branches {
			if err := gitcmd.RunWithIO("branch", "-d", branch); err != nil {
				deleteErrors = true
				fmt.Fprintf(os.Stderr, "ブランチ %s の削除に失敗しました: %v\n", branch, err)
			}
		}

		if deleteErrors {
			return fmt.Errorf("一部のブランチの削除に失敗しました")
		}

		fmt.Println("削除しました。")
		return nil
	},
}

func getMergedBranches() ([]string, error) {
	output, err := gitcmd.Run("branch", "--merged")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var branches []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 現在のブランチ（*が付いている）はスキップ
		if strings.HasPrefix(line, "*") {
			continue
		}

		if line == "" {
			continue
		}

		// 保護対象ブランチはスキップ
		if shouldSkipProtectedBranch(line) {
			continue
		}

		branches = append(branches, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}

func shouldSkipProtectedBranch(branch string) bool {
	switch branch {
	case "main", "master", "develop":
		return true
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(deleteLocalBranchesCmd)
}
