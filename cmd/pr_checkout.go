// ================================================================================
// pr_checkout.go
// ================================================================================
// このファイルは git-plus の pr-checkout コマンドを実装しています。
//
// 【概要】
// pr-checkout コマンドは、最新または指定されたプルリクエストのブランチを
// チェックアウトする機能を提供します。GitHub CLI (gh) を使用してPRと連携します。
//
// 【主な機能】
// - 最新のオープンなPRの自動取得とチェックアウト
// - PR番号を指定したチェックアウト
// - 現在の作業内容の自動保存（pause コマンドと同様）
// - resume コマンドでの復元に対応
// - 既存の pause 状態の検出と上書き確認
//
// 【使用例】
//   git-plus pr-checkout          # 最新のPRをチェックアウト
//   git-plus pr-checkout 123      # PR #123 をチェックアウト
//
// 【内部仕様】
// - GitHub CLI (gh) の gh pr checkout コマンドを使用
// - 現在の状態は $HOME/.config/git-plus/pause-state.json に保存
// - 未コミットの変更は stash に自動保存
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/pausestate"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// prCheckoutCmd は pr-checkout コマンドの定義です。
// 最新または指定されたプルリクエストをチェックアウトします。
var prCheckoutCmd = &cobra.Command{
	Use:   "pr-checkout [PR番号]",
	Short: "最新または指定されたプルリクエストをチェックアウト",
	Long: `最新または指定されたプルリクエストのブランチを取得してチェックアウトします。
現在の作業を自動的に保存（git pause と同様）するため、後で git resume で戻ることができます。

引数なしで実行すると、最新のオープンなPRをチェックアウトします。
PR番号を指定すると、その番号のPRをチェックアウトします。

内部的に GitHub CLI (gh) を使用してプルリクエストと連携します。`,
	Example: `  git-plus pr-checkout          # 最新のPRをチェックアウト
  git-plus pr-checkout 123      # PR #123 をチェックアウト`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLI() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// PR番号を取得
		var prNumber string
		var err error

		if len(args) > 0 {
			prNumber = args[0]
			fmt.Printf("PR #%s をチェックアウトします\n", prNumber)
		} else {
			fmt.Println("最新のPRを取得中...")
			prNumber, err = fetchLatestPRNumber()
			if err != nil {
				return fmt.Errorf("最新のPRの取得に失敗: %w", err)
			}
			if prNumber == "" {
				return fmt.Errorf("オープンなPRが見つかりません")
			}
			fmt.Printf("最新のPR #%s をチェックアウトします\n", prNumber)
		}

		// 現在のブランチを取得
		currentBranch, err := getBranchCurrent()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗: %w", err)
		}

		// 変更があるかチェック
		hasChanges, err := checkUncommittedChanges()
		if err != nil {
			return fmt.Errorf("変更の確認に失敗: %w", err)
		}

		// 既に pause 状態かチェック（変更がある場合のみ確認プロンプトを表示）
		exists, err := pausestate.Exists()
		if err != nil {
			return fmt.Errorf("状態の確認に失敗: %w", err)
		}

		if exists && hasChanges {
			state, err := pausestate.Load()
			if err != nil {
				return fmt.Errorf("既存の状態の読み込みに失敗: %w", err)
			}

			fmt.Printf("警告: 既に pause 状態です（%s → %s）\n", state.FromBranch, state.ToBranch)

			if !ui.Confirm("上書きしてPRをチェックアウトしますか？", false) {
				fmt.Println("キャンセルしました")
				return nil
			}
		}

		var stashRef string
		stashMessage := fmt.Sprintf("git-pr-checkout: from %s", currentBranch)

		if hasChanges {
			fmt.Println("変更を保存中...")
			stashRef, err = createStashWithMessage(stashMessage)
			if err != nil {
				return fmt.Errorf("スタッシュの作成に失敗: %w", err)
			}
			fmt.Printf("✓ 変更を保存しました: %s\n", stashRef)
		} else {
			fmt.Println("変更がないため、スタッシュはスキップします")
			stashRef = ""
		}

		// PRをチェックアウト
		fmt.Printf("PR #%s をチェックアウト中...\n", prNumber)
		targetBranch, err := performPRCheckout(prNumber)
		if err != nil {
			// エラー時はスタッシュを戻す
			if stashRef != "" {
				fmt.Println("スタッシュを復元中...")
				if popErr := popStashNow(); popErr != nil {
					fmt.Printf("警告: スタッシュの復元に失敗: %v\n", popErr)
					fmt.Println("手動で復元してください: git stash pop")
				}
			}
			return fmt.Errorf("PRのチェックアウトに失敗: %w", err)
		}

		// 状態を保存
		state := &pausestate.PauseState{
			FromBranch:   currentBranch,
			ToBranch:     targetBranch,
			StashRef:     stashRef,
			StashMessage: stashMessage,
			Timestamp:    time.Now(),
		}

		if err := pausestate.Save(state); err != nil {
			return fmt.Errorf("状態の保存に失敗: %w", err)
		}

		fmt.Printf("✓ PR #%s のブランチ '%s' にチェックアウトしました\n", prNumber, targetBranch)
		fmt.Println("\n元のブランチに戻るには: git-plus resume")
		return nil
	},
}

// checkGitHubCLI は GitHub CLI (gh) がインストールされているかを確認します。
//
// 戻り値:
//   - bool: gh コマンドが利用可能な場合は true、そうでない場合は false
//
// 内部処理:
//   gh --version コマンドを実行し、成功するかどうかで判定します。
func checkGitHubCLI() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// fetchLatestPRNumber は最新のオープンなPR番号を取得します。
//
// 戻り値:
//   - string: PR番号（数値の文字列表現、PRがない場合は空文字列）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   gh pr list --limit 1 --json number コマンドでJSON形式で最新のPR番号を取得し、
//   パースして返します。
func fetchLatestPRNumber() (string, error) {
	cmd := exec.Command("gh", "pr", "list", "--limit", "1", "--json", "number")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var prs []struct {
		Number int `json:"number"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		return "", fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	if len(prs) == 0 {
		return "", nil
	}

	return fmt.Sprintf("%d", prs[0].Number), nil
}

// performPRCheckout は指定されたPR番号のブランチをチェックアウトします。
//
// パラメータ:
//   - prNumber: チェックアウトするPRの番号
//
// 戻り値:
//   - string: チェックアウトしたブランチ名
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   1. gh pr checkout <prNumber> でPRのブランチをチェックアウト
//   2. チェックアウト後の現在のブランチ名を取得して返す
func performPRCheckout(prNumber string) (string, error) {
	cmd := exec.Command("gh", "pr", "checkout", prNumber)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	branch, err := getBranchCurrent()
	if err != nil {
		return "", fmt.Errorf("チェックアウト後のブランチ名取得に失敗: %w", err)
	}

	return branch, nil
}

// popStashNow は最新のスタッシュを復元します。
//
// 戻り値:
//   - error: スタッシュの復元に失敗した場合のエラー情報
//
// 内部処理:
//   git stash pop コマンドで最新のスタッシュを復元します。
//   エラー発生時（PRチェックアウト失敗など）のロールバックに使用されます。
func popStashNow() error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// init は pr-checkout コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(prCheckoutCmd)
}
