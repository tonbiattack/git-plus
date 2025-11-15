// ================================================================================
// pr_list.go
// ================================================================================
// このファイルは git-plus の pr-list コマンドを実装しています。
//
// 【概要】
// pr-list コマンドは、GitHub CLI の `gh pr list` をラップして、
// Git コマンドとして実行できるようにします。
//
// 【主な機能】
// - PRの一覧表示
// - すべての gh pr list のオプションをサポート
// - シェル補完機能の有効化
//
// 【使用例】
//   git-plus pr-list                    # PR一覧を表示
//   git-plus pr-list --state open       # オープンなPRのみ表示
//   git-plus pr-list --author @me       # 自分が作成したPRを表示
//   git-plus pr-list --assignee @me     # 自分にアサインされたPRを表示
//
// 【内部仕様】
// - GitHub CLI (gh) の gh pr list コマンドをそのままラップ
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

// prListCmd は pr-list コマンドの定義です。
// gh pr list をラップして Git コマンドとして実行できるようにします。
var prListCmd = &cobra.Command{
	Use:   "pr-list [オプション...]",
	Short: "プルリクエスト一覧を表示（gh pr list のラッパー）",
	Long: `GitHub CLI の gh pr list をラップして、Git コマンドとして実行できるようにします。

すべての gh pr list のオプションがそのまま使用できます：
  --state <state>        PR の状態でフィルタ（open, closed, merged, all）
  --author <user>        作成者でフィルタ（@me で自分のPR）
  --assignee <user>      アサイン先でフィルタ（@me で自分がアサインされたPR）
  --label <name>         ラベルでフィルタ
  --limit <int>          表示件数を制限（デフォルト: 30）
  --base <branch>        ベースブランチでフィルタ
  --head <branch>        ヘッドブランチでフィルタ
  --json <fields>        JSON形式で出力
  --jq <expression>      jq式でフィルタ
  --template <string>    Go template形式で出力
  --web                  ブラウザで開く

内部的に GitHub CLI (gh) を使用してプルリクエストの一覧を取得します。`,
	Example: `  git-plus pr-list                      # PR一覧を表示
  git-plus pr-list --state open         # オープンなPRのみ表示
  git-plus pr-list --state closed       # クローズされたPRのみ表示
  git-plus pr-list --state merged       # マージされたPRのみ表示
  git-plus pr-list --author @me         # 自分が作成したPRを表示
  git-plus pr-list --assignee @me       # 自分にアサインされたPRを表示
  git-plus pr-list --limit 10           # 最新10件のPRを表示
  git-plus pr-list --label bug          # "bug" ラベルが付いたPRを表示
  git-plus pr-list --base main          # mainブランチへのPRを表示`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLI() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// gh pr list コマンドを構築
		ghArgs := []string{"pr", "list"}
		ghArgs = append(ghArgs, args...)

		// gh コマンドを実行
		ghCmd := exec.Command("gh", ghArgs...)
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		ghCmd.Stdin = os.Stdin

		if err := ghCmd.Run(); err != nil {
			return fmt.Errorf("gh pr list の実行に失敗: %w", err)
		}

		return nil
	},
}

// init は pr-list コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(prListCmd)
}
