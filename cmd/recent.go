// ================================================================================
// recent.go
// ================================================================================
// このファイルは git-plus の recent コマンドを実装しています。
//
// 【概要】
// recent コマンドは、最近使用したブランチを時系列順に表示し、
// 番号を選択することで即座にブランチを切り替える機能を提供します。
//
// 【主な機能】
// - 最近コミットがあったブランチを最大10件表示
// - コミット日時順（最新順）での表示
// - 現在のブランチは一覧から除外
// - 番号入力による対話的なブランチ切り替え
// - 相対的なコミット日時の表示（例: "2 hours ago"）
//
// 【使用例】
//   git-plus recent  # 最近使用したブランチを表示して選択
//
// 【内部仕様】
// - git for-each-ref --sort=-committerdate を使用してブランチを取得
// - 表示件数は最大10件に制限
// ================================================================================

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

// BranchInfo はブランチの情報を保持する構造体です。
type BranchInfo struct {
	Name         string // ブランチ名
	LastCommitAt string // 最後のコミット日時（相対表記、例: "2 hours ago"）
}

// recentCmd は recent コマンドの定義です。
// 最近使用したブランチを表示して切り替えます。
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

// getRecentBranchesList は最近使用したブランチの一覧を取得します。
//
// 戻り値:
//   - []BranchInfo: ブランチ情報のスライス（コミット日時の降順）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git for-each-ref --sort=-committerdate コマンドで
//   コミット日時順にソートされたブランチ一覧を取得します。
//   出力形式: <ブランチ名>|<相対日時>
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

// getCurrentBranchNow は現在チェックアウトされているブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名（空白や改行は除去されます）
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git branch --show-current コマンドを実行してブランチ名を取得します。
func getCurrentBranchNow() (string, error) {
	output, err := gitcmd.Run("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// switchToSelectedBranch は指定されたブランチに切り替えます。
//
// パラメータ:
//   - branch: 切り替え先のブランチ名
//
// 戻り値:
//   - error: ブランチの切り替えに失敗した場合のエラー情報
//
// 内部処理:
//   git switch <ブランチ名> コマンドを実行します。
func switchToSelectedBranch(branch string) error {
	return gitcmd.RunWithIO("switch", branch)
}

// init は recent コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(recentCmd)
}
