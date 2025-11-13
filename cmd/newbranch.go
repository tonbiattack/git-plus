package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var newbranchCmd = &cobra.Command{
	Use:   "newbranch <ブランチ名>",
	Short: "ブランチを作成または再作成",
	Long: `指定したブランチ名でブランチを作成します。
既にブランチが存在する場合は、以下の選択肢が表示されます：
  [r]ecreate - ブランチを削除して作り直す
  [s]witch   - 既存のブランチに切り替える
  [c]ancel   - 処理を中止する`,
	Example: `  git-plus newbranch feature/awesome`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		// ブランチが存在するかチェック
		exists, err := checkBranchExists(branch)
		if err != nil {
			return fmt.Errorf("ブランチの存在確認に失敗しました: %w", err)
		}

		// ブランチが既に存在する場合の処理
		if exists {
			action, err := askUserAction(branch)
			if err != nil {
				return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
			}

			if action == "cancel" {
				fmt.Println("処理を中止しました。")
				return nil
			}

			if action == "switch" {
				if err := gitcmd.RunWithIO("switch", branch); err != nil {
					return fmt.Errorf("ブランチの切り替えに失敗しました: %w", err)
				}
				fmt.Printf("ブランチ %s に切り替えました。\n", branch)
				return nil
			}
			// action == "recreate" の場合は下に続く
		}

		// 既存ブランチを強制削除
		if err := gitcmd.RunWithIO("branch", "-D", branch); err != nil && !isBranchNotFound(err) {
			return fmt.Errorf("ブランチの削除に失敗しました: %w", err)
		}

		// 新しいブランチを作成して切り替え
		if err := gitcmd.RunWithIO("switch", "-c", branch); err != nil {
			return fmt.Errorf("ブランチ作成に失敗しました: %w", err)
		}

		fmt.Printf("ブランチ %s を作成しました。\n", branch)
		return nil
	},
}

func checkBranchExists(name string) (bool, error) {
	ref := fmt.Sprintf("refs/heads/%s", name)
	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", ref)

	if err != nil {
		if gitcmd.IsExitError(err, 1) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func askUserAction(branch string) (string, error) {
	fmt.Printf("ブランチ %s は既に存在します。どうしますか？ [r]ecreate/[s]witch/[c]ancel (r/s/c): ", branch)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			input = "c"
		} else {
			return "", err
		}
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	switch answer {
	case "r", "recreate":
		return "recreate", nil
	case "s", "switch":
		return "switch", nil
	case "c", "cancel", "":
		return "cancel", nil
	default:
		return "cancel", nil
	}
}

func isBranchNotFound(err error) bool {
	return gitcmd.IsExitError(err, 1)
}

func init() {
	rootCmd.AddCommand(newbranchCmd)
}
