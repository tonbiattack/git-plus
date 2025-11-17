// ================================================================================
// worktree_delete.go
// ================================================================================
// このファイルは git の拡張コマンド worktree-delete を実装しています。
//
// 【概要】
// worktree-delete コマンドは、既存の git worktree を一覧表示し、
// インタラクティブに選択して削除する機能を提供します。
//
// 【主な機能】
// - 既存の worktree を一覧表示
// - 番号入力による対話的な選択
// - 確認プロンプトによる安全な削除
// - --force フラグによる強制削除のサポート
//
// 【使用例】
//   git worktree-delete           # worktree 一覧を表示して選択・削除
//   git worktree-delete --force   # 未コミットの変更があっても強制削除
// ================================================================================

package worktree

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var forceDelete bool

var worktreeDeleteCmd = &cobra.Command{
	Use:   "worktree-delete",
	Short: "既存の worktree を一覧表示して削除",
	Long: `既存の git worktree を一覧表示し、インタラクティブに選択して削除します。
削除前に確認プロンプトが表示されます。

--force フラグを使用すると、未コミットの変更があっても強制削除できます。`,
	Example: `  git worktree-delete
  git worktree-delete --force`,
	RunE: func(c *cobra.Command, args []string) error {
		// worktree 一覧を取得
		worktrees, err := getWorktreeList()
		if err != nil {
			return fmt.Errorf("worktree 一覧の取得に失敗しました: %w", err)
		}

		if len(worktrees) == 0 {
			fmt.Println("worktree が存在しません。")
			return nil
		}

		// 現在の worktree を取得
		currentPath, err := getCurrentWorktreePath()
		if err != nil {
			fmt.Printf("警告: 現在の worktree の取得に失敗しました: %v\n", err)
		}

		// worktree 一覧を表示（現在の worktree とメインの worktree を除く）
		fmt.Println("Worktree 一覧:")
		displayCount := 0
		validWorktrees := make([]WorktreeInfo, 0)

		for _, wt := range worktrees {
			// 現在の worktree は削除対象から除外
			if wt.Path == currentPath {
				continue
			}
			// bare リポジトリも除外
			if wt.Branch == "(bare)" {
				continue
			}
			displayCount++
			validWorktrees = append(validWorktrees, wt)
			fmt.Printf("\n%d. %s\n", displayCount, wt.Path)
			fmt.Printf("   ブランチ: %s\n", wt.Branch)
			fmt.Printf("   コミット: %s\n", wt.Commit)
		}

		if displayCount == 0 {
			fmt.Println("\n削除可能な worktree がありません。")
			fmt.Println("現在の worktree のみが存在します。")
			return nil
		}

		// worktree を選択
		reader := bufio.NewReader(os.Stdin)
		var selectedWorktree WorktreeInfo
		for {
			fmt.Print("\n削除する worktree を選択してください (番号を入力、Enterでキャンセル): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
			}

			input = ui.NormalizeNumberInput(input)
			if input == "" {
				fmt.Println("キャンセルしました。")
				return nil
			}

			selection, err := strconv.Atoi(input)
			if err != nil || selection < 1 || selection > displayCount {
				fmt.Printf("無効な番号です。1から%dの範囲で入力してください。\n", displayCount)
				continue
			}

			selectedWorktree = validWorktrees[selection-1]
			break
		}

		fmt.Printf("\n選択された worktree:\n")
		fmt.Printf("  パス: %s\n", selectedWorktree.Path)
		fmt.Printf("  ブランチ: %s\n", selectedWorktree.Branch)

		// 削除確認
		fmt.Printf("\n本当にこの worktree を削除しますか？ (y/N): ")
		confirmInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		confirmInput = strings.TrimSpace(strings.ToLower(confirmInput))
		if confirmInput != "y" && confirmInput != "yes" {
			fmt.Println("削除をキャンセルしました。")
			return nil
		}

		// worktree を削除
		fmt.Printf("\nworktree を削除しています...\n")

		var removeArgs []string
		if forceDelete {
			removeArgs = []string{"worktree", "remove", "--force", selectedWorktree.Path}
		} else {
			removeArgs = []string{"worktree", "remove", selectedWorktree.Path}
		}

		if err := gitcmd.RunWithIO(removeArgs...); err != nil {
			if !forceDelete {
				fmt.Printf("\n削除に失敗しました。未コミットの変更がある場合は --force オプションを使用してください。\n")
			}
			return fmt.Errorf("worktree の削除に失敗しました: %w", err)
		}

		fmt.Printf("\n✓ worktree を削除しました: %s\n", selectedWorktree.Path)
		fmt.Printf("  ブランチ '%s' は保持されています。\n", selectedWorktree.Branch)
		fmt.Printf("  ブランチも削除する場合は 'git branch -d %s' を実行してください。\n", selectedWorktree.Branch)

		return nil
	},
}

func init() {
	worktreeDeleteCmd.Flags().BoolVar(&forceDelete, "force", false, "未コミットの変更があっても強制削除")
	cmd.RootCmd.AddCommand(worktreeDeleteCmd)
}
