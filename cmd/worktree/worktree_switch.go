// ================================================================================
// worktree_switch.go
// ================================================================================
// このファイルは git の拡張コマンド worktree-switch を実装しています。
//
// 【概要】
// worktree-switch コマンドは、既存の git worktree を一覧表示し、
// インタラクティブに選択して移動する機能を提供します。
//
// 【主な機能】
// - 既存の worktree を一覧表示
// - 番号入力による対話的な選択
// - 選択した worktree のディレクトリに移動してVSCodeを開く
//
// 【使用例】
//   git worktree-switch           # worktree 一覧を表示して選択
//   git worktree-switch --no-code # VSCodeを開かない
// ================================================================================

package worktree

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// WorktreeInfo は worktree の情報を保持する構造体です
type WorktreeInfo struct {
	Path   string // worktree のパス
	Branch string // ブランチ名
	Commit string // コミットハッシュ（短縮形）
}

var noCodeSwitch bool

var worktreeSwitchCmd = &cobra.Command{
	Use:   "worktree-switch",
	Short: "既存の worktree を一覧表示して切り替え",
	Long: `既存の git worktree を一覧表示し、インタラクティブに選択して移動します。
選択した worktree のディレクトリでVSCodeを開きます。

これにより、複数のタスクを並行して作業する際の切り替えが容易になります。`,
	Example: `  git worktree-switch
  git worktree-switch --no-code`,
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

		// worktree 一覧を表示（現在の worktree を除く）
		fmt.Println("Worktree 一覧:")
		displayCount := 0
		validWorktrees := make([]WorktreeInfo, 0)

		for _, wt := range worktrees {
			if wt.Path == currentPath {
				continue
			}
			displayCount++
			validWorktrees = append(validWorktrees, wt)
			fmt.Printf("\n%d. %s\n", displayCount, wt.Path)
			fmt.Printf("   ブランチ: %s\n", wt.Branch)
			fmt.Printf("   コミット: %s\n", wt.Commit)
		}

		if displayCount == 0 {
			fmt.Println("\n切り替え可能な worktree がありません。")
			fmt.Println("現在の worktree のみが存在します。")
			return nil
		}

		// worktree を選択
		reader := bufio.NewReader(os.Stdin)
		var selectedWorktree WorktreeInfo
		for {
			fmt.Print("\n選択してください (番号を入力、Enterでキャンセル): ")
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

		fmt.Printf("\n選択された worktree: %s\n", selectedWorktree.Path)
		fmt.Printf("ブランチ: %s\n", selectedWorktree.Branch)

		// VSCodeを開く（--no-code フラグがない場合）
		if !noCodeSwitch {
			fmt.Printf("\nVSCode を開いています...\n")
			if err := openVSCodeForSwitch(selectedWorktree.Path); err != nil {
				fmt.Printf("警告: VSCode を開けませんでした: %v\n", err)
				fmt.Printf("手動で開いてください: code %s\n", selectedWorktree.Path)
			} else {
				fmt.Printf("✓ VSCode を開きました: %s\n", selectedWorktree.Path)
			}
		} else {
			fmt.Printf("\n✓ 選択した worktree: %s\n", selectedWorktree.Path)
			fmt.Printf("ディレクトリに移動してください: cd %s\n", selectedWorktree.Path)
		}

		return nil
	},
}

// getWorktreeList は全ての worktree 情報を取得します
func getWorktreeList() ([]WorktreeInfo, error) {
	output, err := gitcmd.Run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	worktrees := make([]WorktreeInfo, 0)

	var currentWorktree WorktreeInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if currentWorktree.Path != "" {
				worktrees = append(worktrees, currentWorktree)
				currentWorktree = WorktreeInfo{}
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			currentWorktree.Path = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "HEAD ") {
			commit := strings.TrimPrefix(line, "HEAD ")
			if len(commit) > 7 {
				currentWorktree.Commit = commit[:7]
			} else {
				currentWorktree.Commit = commit
			}
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			// refs/heads/ を除去
			branch = strings.TrimPrefix(branch, "refs/heads/")
			currentWorktree.Branch = branch
		} else if line == "detached" {
			currentWorktree.Branch = "(detached HEAD)"
		} else if line == "bare" {
			currentWorktree.Branch = "(bare)"
		}
	}

	// 最後の worktree を追加
	if currentWorktree.Path != "" {
		worktrees = append(worktrees, currentWorktree)
	}

	return worktrees, nil
}

// getCurrentWorktreePath は現在の worktree のパスを取得します
func getCurrentWorktreePath() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// openVSCodeForSwitch は指定されたディレクトリでVSCodeを開きます
func openVSCodeForSwitch(path string) error {
	cmd := exec.Command("code", path)
	return cmd.Start()
}

func init() {
	worktreeSwitchCmd.Flags().BoolVar(&noCodeSwitch, "no-code", false, "VSCode を開かない")
	cmd.RootCmd.AddCommand(worktreeSwitchCmd)
}
