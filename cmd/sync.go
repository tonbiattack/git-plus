// ================================================================================
// sync.go
// ================================================================================
// このファイルは git-plus の sync コマンドを実装しています。
//
// 【概要】
// sync コマンドは、現在のブランチを最新のリモートブランチと同期する機能を提供します。
// 内部的に git rebase を使用するため、履歴がきれいに保たれます。
//
// 【主な機能】
// - リモートの最新変更の自動取得（git fetch origin）
// - リモートブランチへの自動リベース（git rebase origin/<ブランチ>）
// - デフォルトブランチ（main/master）の自動検出
// - コンフリクト発生時の適切な処理と復旧オプション
//   - --continue: コンフリクト解決後にリベースを続行
//   - --abort: リベースを中止して元の状態に戻す
//
// 【使用例】
//   git-plus sync                  # origin/main または origin/master と同期
//   git-plus sync develop          # origin/develop と同期
//   git-plus sync --continue       # コンフリクト解決後に続行
//   git-plus sync --abort          # 同期を中止
// ================================================================================

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var (
	syncContinue bool // コンフリクト解決後に rebase を続行するフラグ
	syncAbort    bool // rebase を中止するフラグ
)

// syncCmd は sync コマンドの定義です。
// 現在のブランチを最新のリモートブランチと同期します。
var syncCmd = &cobra.Command{
	Use:   "sync [ブランチ名]",
	Short: "現在のブランチを最新のリモートブランチと同期",
	Long: `現在のブランチを最新の origin/<ブランチ> と同期します。
内部的に git rebase を使用するため、履歴がきれいに保たれます。`,
	Example: `  git-plus sync                  # origin/main (または origin/master) と同期
  git-plus sync develop          # origin/develop と同期
  git-plus sync --continue       # コンフリクト解決後に続行
  git-plus sync --abort          # 同期を中止`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --continue オプションの処理
		if syncContinue {
			if err := continueRebaseOp(); err != nil {
				return fmt.Errorf("rebase の続行に失敗しました: %w", err)
			}
			fmt.Println("同期が完了しました。")
			return nil
		}

		// --abort オプションの処理
		if syncAbort {
			if err := abortRebaseOp(); err != nil {
				return fmt.Errorf("rebase の中止に失敗しました: %w", err)
			}
			fmt.Println("同期を中止しました。")
			return nil
		}

		// ターゲットブランチの決定
		targetBranch := ""
		if len(args) > 0 {
			targetBranch = args[0]
		} else {
			branch, err := detectDefaultRemoteBranch()
			if err != nil {
				return fmt.Errorf("デフォルトブランチの検出に失敗しました: %w", err)
			}
			targetBranch = branch
		}

		// git fetch origin を実行
		fmt.Println("origin から最新の変更を取得しています...")
		if err := gitcmd.RunWithIO("fetch", "origin"); err != nil {
			return fmt.Errorf("fetch に失敗しました: %w", err)
		}

		// git rebase origin/<ブランチ> を実行
		remoteBranch := fmt.Sprintf("origin/%s", targetBranch)
		fmt.Printf("%s にリベースしています...\n", remoteBranch)
		if err := gitcmd.RunWithIO("rebase", remoteBranch); err != nil {
			if checkRebaseInProgress() {
				fmt.Println("\nコンフリクトが発生しました。")
				fmt.Println("コンフリクトを解決した後、以下のコマンドを実行してください:")
				fmt.Println("  git-plus sync --continue    # 同期を続行")
				fmt.Println("  git-plus sync --abort       # 同期を中止")
				return fmt.Errorf("コンフリクトが発生しました")
			}
			return fmt.Errorf("rebase に失敗しました: %w", err)
		}

		fmt.Printf("同期が完了しました。(%s)\n", remoteBranch)
		return nil
	},
}

// detectDefaultRemoteBranch はリモートのデフォルトブランチを検出します。
//
// 戻り値:
//   - string: 検出されたブランチ名（"main" または "master"）
//   - error: どちらも見つからない場合のエラー情報
//
// 内部処理:
//   1. origin/main の存在を確認
//   2. 見つからない場合は origin/master の存在を確認
//   3. どちらも見つからない場合はエラーを返す
func detectDefaultRemoteBranch() (string, error) {
	// origin/main の存在確認
	if err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/remotes/origin/main"); err == nil {
		return "main", nil
	}

	// origin/master の存在確認
	if err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/remotes/origin/master"); err == nil {
		return "master", nil
	}

	return "", fmt.Errorf("origin/main も origin/master も見つかりませんでした")
}

// checkRebaseInProgress は現在リベース処理が進行中かどうかを確認します。
//
// 戻り値:
//   - bool: リベース中の場合は true、そうでない場合は false
//
// 内部処理:
//   .git/rebase-merge または .git/rebase-apply ディレクトリの存在を確認します。
//   これらのディレクトリが存在する場合、リベース処理が進行中です。
func checkRebaseInProgress() bool {
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return true
	}
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return true
	}
	return false
}

// continueRebaseOp はコンフリクト解決後にリベース処理を続行します。
//
// 戻り値:
//   - error: リベースの続行に失敗した場合のエラー情報
//
// 内部処理:
//   git rebase --continue コマンドを実行します。
//   このコマンドは、ユーザーがコンフリクトを手動で解決した後に呼び出されます。
func continueRebaseOp() error {
	return gitcmd.RunWithIO("rebase", "--continue")
}

// abortRebaseOp はリベース処理を中止し、元の状態に戻します。
//
// 戻り値:
//   - error: リベースの中止に失敗した場合のエラー情報
//
// 内部処理:
//   git rebase --abort コマンドを実行します。
//   これにより、リベース開始前の状態に完全に戻ります。
func abortRebaseOp() error {
	return gitcmd.RunWithIO("rebase", "--abort")
}

// init は sync コマンドを root コマンドに登録し、フラグを設定します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
//
// 設定されるフラグ:
//   --continue: コンフリクト解決後に rebase を続行
//   --abort: 同期を中止して元の状態に戻す
func init() {
	syncCmd.Flags().BoolVar(&syncContinue, "continue", false, "コンフリクト解決後に rebase を続行")
	syncCmd.Flags().BoolVar(&syncAbort, "abort", false, "同期を中止して元の状態に戻す")
	rootCmd.AddCommand(syncCmd)
}
