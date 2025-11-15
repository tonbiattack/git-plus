// ================================================================================
// newbranch.go
// ================================================================================
// このファイルは git の拡張コマンド newbranch コマンドを実装しています。
//
// 【概要】
// newbranch コマンドは、新しいブランチを作成するか、既存のブランチを再作成するための
// 対話的なインターフェースを提供します。既にブランチが存在する場合は、ユーザーに
// 再作成、切り替え、キャンセルの選択肢を提示します。
//
// 【主な機能】
// - 新しいブランチの作成と自動チェックアウト
// - 既存ブランチの削除と再作成
// - 既存ブランチへの切り替え
// - 対話的なユーザー確認プロンプト
//
// 【使用例】
//   git newbranch feature/awesome  # 新しいブランチを作成
//   git newbranch main             # 既存ブランチの場合は選択肢を表示
// ================================================================================

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

// newbranchCmd は newbranch コマンドの定義です。
// ブランチを作成または再作成します。既にブランチが存在する場合は、
// ユーザーに対話的に選択肢（再作成/切り替え/キャンセル）を提示します。
var newbranchCmd = &cobra.Command{
	Use:   "newbranch <ブランチ名>",
	Short: "ブランチを作成または再作成",
	Long: `指定したブランチ名でブランチを作成します。
既にブランチが存在する場合は、以下の選択肢が表示されます：
  [r]ecreate - ブランチを削除して作り直す
  [s]witch   - 既存のブランチに切り替える
  [c]ancel   - 処理を中止する`,
	Example: `  git newbranch feature/awesome`,
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

// checkBranchExists は指定されたブランチが存在するかどうかを確認します。
//
// パラメータ:
//   - name: チェックするブランチ名
//
// 戻り値:
//   - bool: ブランチが存在する場合は true、存在しない場合は false
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git show-ref --verify --quiet refs/heads/<ブランチ名> を実行し、
//   終了コードで存在を判定します。終了コード 1 は「存在しない」を意味します。
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

// askUserAction はブランチが既に存在する場合にユーザーにアクションを尋ねます。
//
// パラメータ:
//   - branch: 既に存在するブランチ名
//
// 戻り値:
//   - string: ユーザーが選択したアクション ("recreate", "switch", "cancel")
//   - error: 入力の読み込みに失敗した場合のエラー情報
//
// ユーザー入力の処理:
//   - "r" または "recreate": ブランチを削除して再作成
//   - "s" または "switch": 既存のブランチに切り替え
//   - "c", "cancel", または空入力: 処理をキャンセル
//   - 上記以外: キャンセルとして扱う
//   - EOF の場合: 自動的にキャンセル
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

// isBranchNotFound はエラーがブランチ未発見エラーかどうかを判定します。
//
// パラメータ:
//   - err: チェックするエラー
//
// 戻り値:
//   - bool: git コマンドの終了コードが 1（ブランチ未発見）の場合は true
//
// 備考:
//   git branch -D コマンドで存在しないブランチを削除しようとした場合の
//   エラーを判定するために使用されます。
func isBranchNotFound(err error) bool {
	return gitcmd.IsExitError(err, 1)
}

// init は newbranch コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(newbranchCmd)
}
