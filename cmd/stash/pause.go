// ================================================================================
// pause.go
// ================================================================================
// このファイルは git の拡張コマンド pause コマンドを実装しています。
// stash パッケージ内に配置され、スタッシュ関連の機能を提供します。
//
// 【概要】
// pause コマンドは、現在のブランチでの作業を一時停止し、別のブランチに切り替える
// 機能を提供します。未コミットの変更は自動的に stash に保存され、後で resume コマンドで
// 復元できます。
//
// 【主な機能】
// - 現在の作業内容（未コミット変更）を stash に保存
// - 現在のブランチ情報と移動先ブランチ情報を状態ファイルに保存
// - 指定されたブランチへの自動切り替え
// - 既存の pause 状態の検出と上書き確認
//
// 【使用例】
//   git pause main              # main ブランチに一時切り替え
//   git pause feature/login     # feature/login ブランチに一時切り替え
//
// 【内部仕様】
// - 状態は $HOME/.config/git-plus/pause-state.json に保存されます
// - stash メッセージには "git-pause: from <ブランチ名>" の形式が使用されます
// - 変更がない場合は stash をスキップします
// ================================================================================

package stash

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/pausestate"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// pauseCmd は pause コマンドの定義です。
// 作業中の変更を stash に保存し、別のブランチへ移動します。
// 後で resume コマンドで元のブランチと変更を復元できます。
var pauseCmd = &cobra.Command{
	Use:   "pause <branch>",
	Short: "作業中の変更を stash して別ブランチへ移動します",
	Long: `現在のブランチ名とコミットしていない変更を保存し、指定したブランチへ切り替えます。
作業へ戻る準備ができたら git resume を実行して保存内容を復元してください。`,
	Example: `  git pause main
  git pause feature/login`,
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
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
			_ = pausestate.Delete()
			return fmt.Errorf("ブランチの切り替えに失敗: %w", err)
		}

		fmt.Printf("✓ %s に切り替えました\n", targetBranch)
		fmt.Println("\n元のブランチに戻るには: git resume")
		return nil
	},
}

// getBranchCurrent は現在チェックアウトされているブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名（空白や改行は除去されます）
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git branch --show-current コマンドを実行してブランチ名を取得します。
func getBranchCurrent() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkUncommittedChanges は未コミットの変更があるかどうかを確認します。
//
// 戻り値:
//   - bool: 未コミットの変更がある場合は true、ない場合は false
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git status --porcelain コマンドを実行し、出力があるかどうかで判定します。
//   変更がない場合は空の出力が返されます。
func checkUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// createStashWithMessage は指定されたメッセージで stash を作成します。
//
// パラメータ:
//   - message: stash に付けるメッセージ
//
// 戻り値:
//   - string: 作成された stash の参照（SHA-1 ハッシュ）
//   - error: stash の作成に失敗した場合のエラー情報
//
// 内部処理:
//   1. git stash push -m "<メッセージ>" で変更を stash に保存
//   2. git rev-parse stash@{0} で最新の stash の参照を取得
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

// checkoutBranch は指定されたブランチにチェックアウトします。
//
// パラメータ:
//   - branch: チェックアウトするブランチ名
//
// 戻り値:
//   - error: チェックアウトに失敗した場合のエラー情報
//
// 内部処理:
//   git checkout <ブランチ名> コマンドを実行します。
func checkoutBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	return cmd.Run()
}

// init は pause コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	cmd.RootCmd.AddCommand(pauseCmd)
}
