package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/pausestate"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "git pause で保存した作業を再開します",
	Long: `git pause が記録したブランチへ戻り、必要であれば保存された stash を適用して、
休止状態のメタデータを削除します。`,
	Example: `  git-plus resume`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 状態ファイルが存在するかチェック
		exists, err := pausestate.Exists()
		if err != nil {
			return fmt.Errorf("状態の確認に失敗: %w", err)
		}

		if !exists {
			fmt.Println("エラー: pause 状態がありません")
			fmt.Println("git pause <branch> で作業を一時保存してください")
			return fmt.Errorf("pause 状態がありません")
		}

		// 状態を読み込み
		state, err := pausestate.Load()
		if err != nil {
			return fmt.Errorf("状態の読み込みに失敗: %w", err)
		}

		if state == nil {
			return fmt.Errorf("pause 状態がありません")
		}

		fmt.Printf("元のブランチに戻ります: %s → %s\n", state.ToBranch, state.FromBranch)

		// 現在のブランチを確認
		currentBranch, err := getCurrentBranchName()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗: %w", err)
		}

		// ブランチを切り替え
		if currentBranch != state.FromBranch {
			fmt.Printf("ブランチを切り替え中: %s → %s\n", currentBranch, state.FromBranch)
			if err := switchBranchTo(state.FromBranch); err != nil {
				return fmt.Errorf("ブランチの切り替えに失敗: %w", err)
			}
			fmt.Printf("✓ %s に切り替えました\n", state.FromBranch)
		} else {
			fmt.Printf("既に %s にいます\n", state.FromBranch)
		}

		// スタッシュを復元（スタッシュが存在する場合のみ）
		if state.StashRef != "" {
			fmt.Println("変更を復元中...")
			if err := popStashRef(state.StashRef); err != nil {
				fmt.Println("警告: スタッシュの復元に失敗しました")
				fmt.Println("手動で復元してください: git stash list")
				return fmt.Errorf("スタッシュの復元に失敗: %w", err)
			}
			fmt.Println("✓ 変更を復元しました")
		} else {
			fmt.Println("復元するスタッシュがありません")
		}

		// 状態ファイルを削除
		if err := pausestate.Delete(); err != nil {
			fmt.Printf("警告: 状態ファイルの削除に失敗: %v\n", err)
		}

		fmt.Println("\n✓ 作業の復元が完了しました")
		return nil
	},
}

func getCurrentBranchName() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func switchBranchTo(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	return cmd.Run()
}

func popStashRef(stashRef string) error {
	// stash@{0} の形式でスタッシュを検索
	cmd := exec.Command("git", "stash", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("スタッシュ一覧の取得に失敗: %w", err)
	}

	stashList := strings.TrimSpace(string(output))
	if stashList == "" {
		return fmt.Errorf("スタッシュが見つかりません")
	}

	// スタッシュを pop
	cmd = exec.Command("git", "stash", "pop")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("スタッシュの適用に失敗: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
