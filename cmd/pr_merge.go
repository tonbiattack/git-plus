// ================================================================================
// pr_merge.go
// ================================================================================
// このファイルは git-plus の pr-merge コマンドを実装しています。
//
// 【概要】
// pr-merge コマンドは、GitHub CLI の `gh pr merge` をラップして、
// Git コマンドとして実行できるようにします。
//
// 【主な機能】
// - PRのマージ（merge commit / squash / rebase）
// - デフォルト: マージコミットで対話なしで直接実行
// - ブランチの削除（デフォルトで有効）
// - 自動マージの設定
// - すべての gh pr merge のオプションをサポート
//
// 【使用例】
//   git-plus pr-merge              # PR をマージコミットで直接マージ（ブランチも削除）
//   git-plus pr-merge 89           # PR #89 をマージコミットで直接マージ（ブランチも削除）
//   git-plus pr-merge --squash     # スカッシュマージ（ブランチも削除）
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

// prMergeCmd は pr-merge コマンドの定義です。
// gh pr merge をラップして Git コマンドとして実行できるようにします。
var prMergeCmd = &cobra.Command{
	Use:   "pr-merge [PR番号] [オプション...]",
	Short: "プルリクエストをマージ（gh pr merge のラッパー）",
	Long: `GitHub CLI の gh pr merge をラップして、Git コマンドとして実行できるようにします。

デフォルトの動作：
  - マージコミットで対話なしで直接実行（--merge が自動適用）
  - マージ後にブランチを削除（--delete-branch が自動適用）

引数なしで実行すると、カレントブランチのPRをマージします。
PR番号を指定すると、その番号のPRをマージします。

すべての gh pr merge のオプションがそのまま使用できます：
  --merge           マージコミットを作成（デフォルト）
  --squash          スカッシュマージ（デフォルトを上書き）
  --rebase          リベースマージ（デフォルトを上書き）
  --delete-branch   マージ後にブランチを削除（デフォルト）
  --auto            ステータスチェック通過後に自動マージ
  --body <text>     マージコミットのボディ
  --subject <text>  マージコミットのサブジェクト

内部的に GitHub CLI (gh) を使用してプルリクエストをマージします。`,
	Example: `  git-plus pr-merge                    # カレントブランチのPRをマージコミットで直接マージ
  git-plus pr-merge 89                 # PR #89 をマージコミットで直接マージ
  git-plus pr-merge --squash           # スカッシュマージで直接マージ
  git-plus pr-merge --rebase           # リベースマージで直接マージ
  git-plus pr-merge 89 --squash --auto # PR #89 をスカッシュマージで自動マージ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLI() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// gh pr merge コマンドを構築
		ghArgs := []string{"pr", "merge"}

		// デフォルトで --merge と --delete-branch を追加
		// ユーザーが既に指定している場合は追加しない
		hasDeleteBranch := false
		hasMergeMethod := false
		for _, arg := range args {
			if arg == "--delete-branch" || arg == "-d" {
				hasDeleteBranch = true
			}
			if arg == "--merge" || arg == "--squash" || arg == "--rebase" {
				hasMergeMethod = true
			}
		}

		// デフォルトでマージコミットを使用（対話なしで直接実行）
		if !hasMergeMethod {
			ghArgs = append(ghArgs, "--merge")
		}

		// デフォルトでブランチを削除
		if !hasDeleteBranch {
			ghArgs = append(ghArgs, "--delete-branch")
		}

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

// init は pr-merge コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(prMergeCmd)
}
