package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var prMergeCmd = &cobra.Command{
	Use:   "pr-merge [ベースブランチ名]",
	Short: "PRの作成からマージ、ブランチ削除までを一気に実行",
	Long: `以下の処理を自動的に実行します:
  1. タイトル・本文なしでPRを作成（--fillオプション使用）
  2. PRをマージしてブランチを削除（--merge --delete-branch）
  3. ベースブランチに切り替え（git switch）
  4. 最新の変更を取得（git pull）`,
	Example: `  git-plus pr-merge              # 対話的にベースブランチを入力
  git-plus pr-merge main         # mainブランチへマージ
  git-plus pr-merge develop      # developブランチへマージ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 現在のブランチを取得
		currentBranch, err := getCurrentBranchForPR()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗しました: %w", err)
		}

		if currentBranch == "" {
			return fmt.Errorf("ブランチが見つかりません。detached HEAD 状態かもしれません")
		}

		fmt.Printf("現在のブランチ: %s\n", currentBranch)

		// ベースブランチの取得
		var baseBranch string
		if len(args) > 0 {
			baseBranch = args[0]
		} else {
			fmt.Print("マージ先のベースブランチを入力してください (デフォルト: main): ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
			}

			baseBranch = strings.TrimSpace(input)
			if baseBranch == "" {
				baseBranch = "main"
			}
		}

		fmt.Printf("\nベースブランチ: %s\n", baseBranch)
		fmt.Printf("ヘッドブランチ: %s\n", currentBranch)

		if !ui.Confirm("\nPRを作成してマージしますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// Step 1: PRを作成
		fmt.Println("\n[1/4] PRを作成しています...")
		if err := createPRForMerge(baseBranch, currentBranch); err != nil {
			return fmt.Errorf("PRの作成に失敗しました: %w", err)
		}
		fmt.Println("✓ PRを作成しました")

		// Step 2: PRをマージしてブランチを削除
		fmt.Println("\n[2/4] PRをマージしてブランチを削除しています...")
		if err := mergePRAndDeleteBranch(); err != nil {
			return fmt.Errorf("PRのマージに失敗しました: %w", err)
		}
		fmt.Println("✓ PRをマージしてブランチを削除しました")

		// Step 3: ベースブランチに切り替え
		fmt.Printf("\n[3/4] ブランチ '%s' に切り替えています...\n", baseBranch)
		if err := switchToBranch(baseBranch); err != nil {
			return fmt.Errorf("ブランチの切り替えに失敗しました: %w", err)
		}
		fmt.Printf("✓ ブランチ '%s' に切り替えました\n", baseBranch)

		// Step 4: 最新の変更を取得
		fmt.Println("\n[4/4] 最新の変更を取得しています...")
		if err := pullLatestChanges(); err != nil {
			return fmt.Errorf("git pull に失敗しました: %w", err)
		}
		fmt.Println("✓ 最新の変更を取得しました")

		fmt.Println("\n✓ すべての処理が完了しました！")
		return nil
	},
}

func getCurrentBranchForPR() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func createPRForMerge(base, head string) error {
	cmd := exec.Command("gh", "pr", "create",
		"--base", base,
		"--head", head,
		"--fill")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func mergePRAndDeleteBranch() error {
	cmd := exec.Command("gh", "pr", "merge",
		"--merge",
		"--delete-branch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func switchToBranch(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func pullLatestChanges() error {
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(prMergeCmd)
}
