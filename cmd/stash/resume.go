// ================================================================================
// resume.go
// ================================================================================
// このファイルは git の拡張コマンド resume コマンドを実装しています。
// stash パッケージ内に配置され、スタッシュ関連の機能を提供します。
//
// 【概要】
// resume コマンドは、pause コマンドで一時停止した作業を再開する機能を提供します。
// 保存されていたブランチに戻り、stash に保存されていた変更を復元します。
//
// 【主な機能】
// - pause コマンドで保存された状態ファイルの読み込み
// - 元のブランチへの自動切り替え
// - stash に保存された変更の自動復元
// - 復元後の状態ファイルの自動削除
//
// 【使用例】
//   git resume  # pause で保存した状態を復元
//
// 【内部仕様】
// - 状態ファイルは $HOME/.config/git-plus/pause-state.json から読み込まれます
// - stash が存在する場合は git stash pop で復元されます
// - 復元が成功すると状態ファイルは自動的に削除されます
// ================================================================================

package stash

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/pausestate"
)

// resumeCmd は resume コマンドの定義です。
// git pause で保存した作業を再開し、元のブランチと変更を復元します。
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "git pause で保存した作業を再開します",
	Long: `git pause が記録したブランチへ戻り、必要であれば保存された stash を適用して、
休止状態のメタデータを削除します。`,
	Example: `  git resume`,
	RunE: func(c *cobra.Command, args []string) error {
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

// getCurrentBranchName は現在チェックアウトされているブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名（空白や改行は除去されます）
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git branch --show-current コマンドを実行してブランチ名を取得します。
func getCurrentBranchName() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// switchBranchTo は指定されたブランチに切り替えます。
//
// パラメータ:
//   - branch: 切り替え先のブランチ名
//
// 戻り値:
//   - error: ブランチの切り替えに失敗した場合のエラー情報
//
// 内部処理:
//   git switch <ブランチ名> コマンドを実行します。
func switchBranchTo(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	return cmd.Run()
}

// popStashRef は指定された stash 参照を復元します。
//
// パラメータ:
//   - stashRef: 復元する stash の参照（SHA-1 ハッシュ）
//              注: 現在の実装では参照は使用されず、最新の stash を pop します
//
// 戻り値:
//   - error: stash の復元に失敗した場合のエラー情報
//
// 内部処理:
//   1. git stash list でスタッシュの存在を確認
//   2. git stash pop で最新のスタッシュを復元
//
// 備考:
//   現在の実装では stashRef パラメータは使用されず、常に最新の stash を復元します。
//   将来的には特定の stash を復元できるように改善される可能性があります。
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

// init は resume コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	cmd.RootCmd.AddCommand(resumeCmd)
}
