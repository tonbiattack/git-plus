package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// BranchInfo はブランチの情報を保持する構造体
type BranchInfo struct {
	Name         string
	LastCommitAt string
}

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "最近使用したブランチを表示して切り替え",
	Long: `最近コミットがあったブランチを時系列順（最新順）に最大10件表示します。
番号を入力することで、選択したブランチに即座に切り替えられます。
現在のブランチは一覧から除外されます。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("最近使用したブランチを取得しています...")

		// 最近のブランチを取得
		branches, err := getRecentBranchesList()
		if err != nil {
			return fmt.Errorf("ブランチ一覧の取得に失敗しました: %w", err)
		}

		if len(branches) == 0 {
			fmt.Println("ブランチが見つかりませんでした。")
			return nil
		}

		// 現在のブランチを取得
		currentBranch, err := getCurrentBranchNow()
		if err != nil {
			fmt.Printf("警告: 現在のブランチの取得に失敗しました: %v\n", err)
		}

		// ブランチ一覧を表示
		fmt.Println("\n最近使用したブランチ:")
		displayCount := 0
		for _, branch := range branches {
			if branch.Name == currentBranch {
				continue
			}
			displayCount++
			fmt.Printf("%d. %s\n", displayCount, branch.Name)

			if displayCount >= 10 {
				break
			}
		}

		if displayCount == 0 {
			fmt.Println("切り替え可能なブランチがありません。")
			return nil
		}

		// ブランチ選択
		fmt.Print("\nSelect branch (番号を入力): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("キャンセルしました。")
			return nil
		}

		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > displayCount {
			return fmt.Errorf("無効な番号です。1から%dの範囲で入力してください", displayCount)
		}

		// 選択されたブランチを取得
		selectedBranch := ""
		count := 0
		for _, branch := range branches {
			if branch.Name == currentBranch {
				continue
			}
			count++
			if count == selection {
				selectedBranch = branch.Name
				break
			}
		}

		if selectedBranch == "" {
			return fmt.Errorf("ブランチの選択に失敗しました")
		}

		// 選択されたブランチに切り替え
		fmt.Printf("\nブランチ '%s' に切り替えています...\n", selectedBranch)
		if err := switchToSelectedBranch(selectedBranch); err != nil {
			return fmt.Errorf("ブランチの切り替えに失敗しました: %w", err)
		}

		fmt.Printf("✓ ブランチ '%s' に切り替えました。\n", selectedBranch)
		return nil
	},
}

func getRecentBranchesList() ([]BranchInfo, error) {
	output, err := gitcmd.Run("for-each-ref",
		"--sort=-committerdate",
		"--format=%(refname:short)|%(committerdate:relative)",
		"refs/heads/")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]BranchInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		branches = append(branches, BranchInfo{
			Name:         parts[0],
			LastCommitAt: parts[1],
		})
	}

	return branches, nil
}

func getCurrentBranchNow() (string, error) {
	output, err := gitcmd.Run("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func switchToSelectedBranch(branch string) error {
	return gitcmd.RunWithIO("switch", branch)
}

func init() {
	rootCmd.AddCommand(recentCmd)
}
