// ================================================================================
// rename_branch.go
// ================================================================================
// このファイルでは git の拡張コマンド rename-branch コマンドを実装しています。
//
// 概要:
// rename-branch コマンドは現在チェックアウトしているローカルブランチの名前を変更し、
// 必要に応じてリモートブランチ（デフォルト: origin）も更新します。手作業での
// `git branch -m` や `git push --set-upstream` コマンドの入力ミスを防ぎ、
// 作業ブランチの整理を安全に進められるようにすることが目的です。
//
// 特徴:
// - ブランチ名の重複チェック（既存名との衝突を防止）
// - `--push` で rename 後のブランチをリモートにプッシュして upstream を再設定
// - `--delete-remote`（要 `--push`）で古いリモートブランチを安全に削除（確認プロンプト付き）
// - `--remote` フラグで `origin` 以外のリモート名にも対応
//
// 使い方:
//   git rename-branch feature/renamed         # ローカルのみリネーム
//   git rename-branch release/v2 --push       # リネーム後にリモートへプッシュ
//   git rename-branch hotfix/login --push --delete-remote
// ================================================================================

package branch

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	renamePush         bool   // --push フラグ: リモートに新しいブランチをプッシュ
	renameDeleteRemote bool   // --delete-remote フラグ: リモートの旧ブランチを削除
	renameRemoteName   string // --remote フラグ: 対象となるリモート名
)

// renameBranchCmd は rename-branch コマンドの定義です。
var renameBranchCmd = &cobra.Command{
	Use:   "rename-branch <新しいブランチ名>",
	Short: "現在のブランチ名を変更し、必要に応じてリモートも更新",
	Long: `現在チェックアウトしているブランチ名を変更します。

オプションでリモートブランチを更新し、古いブランチ名を削除できます。
手動での git branch -m / git push の入力ミスを防ぎ、安全に作業できます。`,
	Example: `  git rename-branch feature/renamed
  git rename-branch release/v2 --push
  git rename-branch hotfix/login --push --delete-remote`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRenameBranchCommand(args[0])
	},
}

// runRenameBranchCommand は rename-branch コマンドのメイン処理です。
//
// パラメータ:
//   - newName: 変更後のブランチ名
//
// 主な処理:
//  1. フラグの整合性チェック (--delete-remote は --push が必須)
//  2. 現在のブランチ名と新しいブランチ名の取得・検証
//  3. git branch -m でローカルブランチをリネーム
//  4. --push 指定時は `git push --set-upstream <remote> <new>` を実行
//  5. --delete-remote 指定時は確認後に `git push <remote> --delete <old>` を実行
func runRenameBranchCommand(newName string) error {
	if renameDeleteRemote && !renamePush {
		return fmt.Errorf("--delete-remote を使用するには --push も指定してください")
	}

	targetName := strings.TrimSpace(newName)
	if targetName == "" {
		return fmt.Errorf("新しいブランチ名を指定してください")
	}

	currentBranch, err := getCurrentBranchNow()
	if err != nil {
		return fmt.Errorf("現在のブランチ名の取得に失敗しました: %w", err)
	}
	if currentBranch == "" {
		return fmt.Errorf("現在のブランチを検出できません。HEAD が detach されている可能性があります")
	}
	if currentBranch == targetName {
		return fmt.Errorf("同じ名前には変更できません (%s)", targetName)
	}

	exists, err := checkBranchExists(targetName)
	if err != nil {
		return fmt.Errorf("新しいブランチ名の存在確認に失敗しました: %w", err)
	}
	if exists {
		return fmt.Errorf("ブランチ %s は既に存在します", targetName)
	}

	if err := renameLocalBranch(currentBranch, targetName); err != nil {
		return fmt.Errorf("ブランチ名の変更に失敗しました: %w", err)
	}
	fmt.Printf("ブランチ %s を %s に変更しました。\n", currentBranch, targetName)

	remote := strings.TrimSpace(renameRemoteName)
	if remote == "" {
		remote = "origin"
	}

	deleteRemote := renameDeleteRemote

	if renamePush {
		fmt.Printf("%s/%s に新しいブランチをプッシュしています...\n", remote, targetName)
		if err := pushRenamedBranch(remote, targetName); err != nil {
			return fmt.Errorf("新しいブランチのプッシュに失敗しました: %w", err)
		}
		fmt.Printf("プッシュが完了しました。(%s/%s)\n", remote, targetName)

		if deleteRemote {
			prompt := fmt.Sprintf("リモートブランチ %s/%s を削除しますか？", remote, currentBranch)
			if !ui.Confirm(prompt, false) {
				fmt.Println("リモートブランチの削除をキャンセルしました。")
				deleteRemote = false
			}
		}

		if deleteRemote {
			fmt.Printf("%s/%s を削除しています...\n", remote, currentBranch)
			if err := deleteRemoteBranch(remote, currentBranch); err != nil {
				return fmt.Errorf("リモートブランチの削除に失敗しました: %w", err)
			}
			fmt.Printf("リモートブランチ %s/%s を削除しました。\n", remote, currentBranch)
		} else {
			fmt.Printf("旧ブランチ %s/%s を残す場合は必要に応じて手動で削除してください。\n", remote, currentBranch)
		}
	} else {
		fmt.Println("リモートのブランチ名も変更する場合は以下を実行してください:")
		fmt.Printf("  git push %s --set-upstream %s\n", remote, targetName)
		fmt.Printf("  git push %s --delete %s   # 旧ブランチを削除する場合\n", remote, currentBranch)
	}

	return nil
}

// renameLocalBranch はローカルブランチの名前を変更します。
//
// パラメータ:
//   - oldName: 現在のブランチ名
//   - newName: 新しいブランチ名
//
// 実行する git コマンド:
//   - git branch -m <old> <new>
func renameLocalBranch(oldName, newName string) error {
	return gitcmd.RunWithIO("branch", "-m", oldName, newName)
}

// pushRenamedBranch はリネーム後のブランチをリモートへプッシュし upstream を設定します。
//
// パラメータ:
//   - remote: リモート名（例: origin）
//   - branch: 新しいブランチ名
//
// 実行する git コマンド:
//   - git push --set-upstream <remote> <branch>
func pushRenamedBranch(remote, branch string) error {
	return gitcmd.RunWithIO("push", "--set-upstream", remote, branch)
}

// deleteRemoteBranch はリモート上の旧ブランチを削除します。
//
// パラメータ:
//   - remote: リモート名
//   - branch: 削除したいブランチ名
//
// 実行する git コマンド:
//   - git push <remote> --delete <branch>
func deleteRemoteBranch(remote, branch string) error {
	return gitcmd.RunWithIO("push", remote, "--delete", branch)
}

// init は rename-branch コマンドを rootCmd に登録し、フラグを定義します。
func init() {
	renameBranchCmd.Flags().BoolVar(&renamePush, "push", false, "リモートにも新しいブランチをプッシュして upstream を更新する")
	renameBranchCmd.Flags().BoolVar(&renameDeleteRemote, "delete-remote", false, "リモートの旧ブランチを削除する（--push が必須）")
	renameBranchCmd.Flags().StringVar(&renameRemoteName, "remote", "origin", "リモートブランチを更新する際に使用するリモート名")
	cmd.RootCmd.AddCommand(renameBranchCmd)
}
