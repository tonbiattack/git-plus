/*
Package cmd は git の拡張コマンド各種コマンドを定義します。

このファイル (delete_local_branches.go) は、マージ済みのローカルブランチを
一括削除するコマンドを提供します。

主な機能:
  - git branch --merged に含まれるブランチの取得
  - 保護ブランチ（main, master, develop）の自動除外
  - 現在のブランチの自動除外
  - 削除前の確認プロンプト

使用例:
  git delete-local-branches  # マージ済みのブランチを削除
*/
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

// deleteLocalBranchesCmd はマージ済みのローカルブランチを削除するコマンドです。
// 保護対象のブランチ（main, master, develop）と現在のブランチは削除対象から除外されます。
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

// getMergedBranches はマージ済みのブランチ一覧を取得します。
// 保護ブランチと現在のブランチは除外されます。
//
// 戻り値:
//   - []string: 削除対象のブランチ名のスライス
//   - error: エラーが発生した場合はエラーオブジェクト
func getMergedBranches() ([]string, error) {
	// git branch --merged を実行してマージ済みブランチを取得
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

// shouldSkipProtectedBranch は保護対象のブランチかどうかを判定します。
//
// パラメータ:
//   branch: 判定するブランチ名
//
// 戻り値:
//   保護対象の場合は true、そうでない場合は false
func shouldSkipProtectedBranch(branch string) bool {
	// main, master, develop は保護対象として削除しない
	switch branch {
	case "main", "master", "develop":
		return true
	default:
		return false
	}
}

// init はコマンドの初期化を行います。
// deleteLocalBranchesCmd を rootCmd に登録することで、CLI から実行可能にします。
func init() {
	rootCmd.AddCommand(deleteLocalBranchesCmd)
}
