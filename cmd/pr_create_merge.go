// ================================================================================
// pr_create_merge.go
// ================================================================================
// このファイルは git の拡張コマンド pr-create-merge コマンドを実装しています。
//
// 【概要】
// pr-create-merge コマンドは、PRの作成からマージ、ブランチ削除、ベースブランチへの
// 切り替えまでを一気に実行する自動化機能を提供します。GitHub CLI (gh) を使用します。
//
// 【主な機能】
// - タイトル・本文なしでのPR作成（--fill オプションで自動生成）
// - PRの即座のマージとブランチ削除
// - ベースブランチへの自動切り替え
// - 最新の変更の自動取得（git pull）
// - 対話的なベースブランチの選択
// - 各ステップの進行状況表示
//
// 【使用例】
//   git pr-create-merge              # 対話的にベースブランチを入力
//   git pr-create-merge main         # main ブランチへマージ
//   git pr-create-merge develop      # develop ブランチへマージ
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

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

// prCreateMergeCmd は pr-create-merge コマンドの定義です。
// PRの作成からマージ、ブランチ削除までを一気に実行します。
var prCreateMergeCmd = &cobra.Command{
	Use:   "pr-create-merge [ベースブランチ名]",
	Short: "PRの作成からマージ、ブランチ削除までを一気に実行",
	Long: `以下の処理を自動的に実行します:
  1. タイトル・本文なしでPRを作成（--fillオプション使用）
  2. PRをマージしてブランチを削除（--merge --delete-branch）
  3. ベースブランチに切り替え（git switch）
  4. 最新の変更を取得（git pull）`,
	Example: `  git pr-create-merge              # 対話的にベースブランチを入力
  git pr-create-merge main         # mainブランチへマージ
  git pr-create-merge develop      # developブランチへマージ`,
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

// getCurrentBranchForPR は現在チェックアウトされているブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名（空白や改行は除去されます）
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git branch --show-current コマンドを実行してブランチ名を取得します。
func getCurrentBranchForPR() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// createPRForMerge は指定されたベースブランチとヘッドブランチでPRを作成します。
//
// パラメータ:
//   - base: マージ先のベースブランチ名
//   - head: マージ元のヘッドブランチ名
//
// 戻り値:
//   - error: PR作成に失敗した場合のエラー情報
//
// 内部処理:
//   gh pr create --base <base> --head <head> --fill コマンドを実行します。
//   --fill オプションにより、タイトルと本文はコミット履歴から自動生成されます。
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

// mergePRAndDeleteBranch は最新のPRをマージしてブランチを削除します。
//
// 戻り値:
//   - error: マージまたはブランチ削除に失敗した場合のエラー情報
//
// 内部処理:
//   gh pr merge --merge --delete-branch コマンドを実行します。
//   --merge オプションでマージコミットを作成し、
//   --delete-branch オプションでリモートブランチを自動削除します。
func mergePRAndDeleteBranch() error {
	cmd := exec.Command("gh", "pr", "merge",
		"--merge",
		"--delete-branch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// switchToBranch は指定されたブランチに切り替えます。
//
// パラメータ:
//   - branch: 切り替え先のブランチ名
//
// 戻り値:
//   - error: ブランチの切り替えに失敗した場合のエラー情報
//
// 内部処理:
//   git switch <ブランチ名> コマンドを実行します。
func switchToBranch(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// pullLatestChanges はリモートから最新の変更を取得します。
//
// 戻り値:
//   - error: pull に失敗した場合のエラー情報
//
// 内部処理:
//   git pull コマンドを実行して、リモートの変更をローカルに統合します。
func pullLatestChanges() error {
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// init は pr-create-merge コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(prCreateMergeCmd)
}
