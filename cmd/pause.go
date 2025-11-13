package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/pausestate"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var pauseCmd = &cobra.Command{
	Use:   "pause <branch>",
	Short: "作業中の変更を stash して別ブランチへ移動します",
	Long: `現在のブランチ名とコミットしていない変更を保存し、指定したブランチへ切り替えます。
作業へ戻る準備ができたら git resume を実行して保存内容を復元してください。`,
	Example: `  git-plus pause main
  git-plus pause feature/login`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetBranch := args[0]

		// 既に pause 状態かチェック
		exists, err := pausestate.Exists()
		if err != nil {
			return fmt.Errorf("状態の確認に失敗: %w", err)
		}

		if exists {
			state, err := pausestate.Load()
			if err != nil {
				return fmt.Errorf("既存の状態の読み込みに失敗: %w", err)
			}

			fmt.Printf("警告: 既に pause 状態です（%s → %s）\n", state.FromBranch, state.ToBranch)

			if !ui.Confirm("上書きしますか？", false) {
				fmt.Println("キャンセルしました")
				return nil
			}
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

		var stashRef string
		stashMessage := fmt.Sprintf("git-pause: from %s", currentBranch)

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

		// ブランチを切り替え
		fmt.Printf("ブランチを切り替え中: %s → %s\n", currentBranch, targetBranch)
		if err := checkoutBranch(targetBranch); err != nil {
			pausestate.Delete()
			return fmt.Errorf("ブランチの切り替えに失敗: %w", err)
		}

		fmt.Printf("✓ %s に切り替えました\n", targetBranch)
		fmt.Println("\n元のブランチに戻るには: git-plus resume")
		return nil
	},
}

func getBranchCurrent() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func checkUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func createStashWithMessage(message string) (string, error) {
	cmd := exec.Command("git", "stash", "push", "-m", message)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	cmd = exec.Command("git", "rev-parse", "stash@{0}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func checkoutBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
