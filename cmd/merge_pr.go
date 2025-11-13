// ================================================================================
// merge_pr.go
// ================================================================================
// このファイルは git-plus の merge-pr コマンドを実装しています。
//
// 【概要】
// merge-pr コマンドは、GitHub CLI の `gh pr merge` をラップして、
// Git コマンドとして実行できるようにします。
//
// 【主な機能】
// - PRのマージ（merge commit / squash / rebase）
// - ブランチの削除
// - 自動マージの設定
// - すべての gh pr merge のオプションをサポート
//
// 【使用例】
//   git-plus merge-pr              # 対話的にPRをマージ
//   git-plus merge-pr 89           # PR #89 をマージ
//   git-plus merge-pr --squash     # スカッシュマージ
//   git-plus merge-pr --delete-branch  # ブランチを削除
//
// 【内部仕様】
// - GitHub CLI (gh) の gh pr merge コマンドをそのままラップ
// - すべての引数とオプションを gh に転送
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// mergePRCmd は merge-pr コマンドの定義です。
// gh pr merge をラップして Git コマンドとして実行できるようにします。
var mergePRCmd = &cobra.Command{
	Use:   "merge-pr [PR番号] [オプション...]",
	Short: "プルリクエストをマージ（gh pr merge のラッパー）",
	Long: `GitHub CLI の gh pr merge をラップして、Git コマンドとして実行できるようにします。

引数なしで実行すると、対話的にマージ方法やブランチ削除を選択できます。
PR番号を指定すると、その番号のPRをマージします。

すべての gh pr merge のオプションがそのまま使用できます：
  --merge           マージコミットを作成（デフォルト）
  --squash          スカッシュマージ
  --rebase          リベースマージ
  --delete-branch   マージ後にブランチを削除
  --auto            ステータスチェック通過後に自動マージ
  --body <text>     マージコミットのボディ
  --subject <text>  マージコミットのサブジェクト

内部的に GitHub CLI (gh) を使用してプルリクエストをマージします。`,
	Example: `  git-plus merge-pr                    # 対話的にPRをマージ
  git-plus merge-pr 89                 # PR #89 をマージ
  git-plus merge-pr --squash           # スカッシュマージ
  git-plus merge-pr --delete-branch    # ブランチを削除
  git-plus merge-pr 89 --squash --delete-branch --auto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLI() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// gh pr merge コマンドを構築
		ghArgs := []string{"pr", "merge"}
		ghArgs = append(ghArgs, args...)

		// gh コマンドを実行
		ghCmd := exec.Command("gh", ghArgs...)
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		ghCmd.Stdin = os.Stdin

		if err := ghCmd.Run(); err != nil {
			return fmt.Errorf("gh pr merge の実行に失敗: %w", err)
		}

		return nil
	},
}

// init は merge-pr コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(mergePRCmd)
}
