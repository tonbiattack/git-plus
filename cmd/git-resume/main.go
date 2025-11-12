package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tonbiattack/git-plus/internal/pausestate"
)

// git-resume は git-pause で保存した作業状態を復元するコマンドで、
// 元のブランチへ戻り、stash された変更を適用し、pause 情報を掃除する。
//
// 主な処理フロー:
//  1. -h/--help の処理。
//  2. 保存済み pause 状態の読み込み（元/先ブランチ・stash 参照）。
//  3. 必要に応じて元ブランチへスイッチ。
//  4. stash が記録されていれば適用。
//  5. 復元成功後に pause 状態ファイルを削除。
//
// 使用する主な Git コマンド:
//  - git branch --show-current: 現在のブランチ確認。
//  - git switch <branch>: 元ブランチへ戻る。
//  - git stash list/pop: 退避していた変更を適用。
func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		os.Exit(0)
	}

	// 状態ファイルが存在するかチェック
	exists, err := pausestate.Exists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 状態の確認に失敗: %v\n", err)
		os.Exit(1)
	}

	if !exists {
		fmt.Println("エラー: pause 状態がありません")
		fmt.Println("git pause <branch> で作業を一時保存してください")
		os.Exit(1)
	}

	// 状態を読み込み
	state, err := pausestate.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 状態の読み込みに失敗: %v\n", err)
		os.Exit(1)
	}

	if state == nil {
		fmt.Println("エラー: pause 状態がありません")
		os.Exit(1)
	}

	fmt.Printf("元のブランチに戻ります: %s → %s\n", state.ToBranch, state.FromBranch)

	// 現在のブランチを確認
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 現在のブランチの取得に失敗: %v\n", err)
		os.Exit(1)
	}

	// ブランチを切り替え
	if currentBranch != state.FromBranch {
		fmt.Printf("ブランチを切り替え中: %s → %s\n", currentBranch, state.FromBranch)
		if err := switchBranch(state.FromBranch); err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ブランチの切り替えに失敗: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ %s に切り替えました\n", state.FromBranch)
	} else {
		fmt.Printf("既に %s にいます\n", state.FromBranch)
	}

	// スタッシュを復元（スタッシュが存在する場合のみ）
	if state.StashRef != "" {
		fmt.Println("変更を復元中...")
		if err := popStash(state.StashRef); err != nil {
			fmt.Fprintf(os.Stderr, "警告: スタッシュの復元に失敗: %v\n", err)
			fmt.Println("手動で復元してください: git stash list")
			// スタッシュの復元に失敗しても状態ファイルは削除しない
			os.Exit(1)
		}
		fmt.Println("✓ 変更を復元しました")
	} else {
		fmt.Println("復元するスタッシュがありません")
	}

	// 状態ファイルを削除
	if err := pausestate.Delete(); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 状態ファイルの削除に失敗: %v\n", err)
	}

	fmt.Println("\n✓ 作業の復元が完了しました")
}

// getCurrentBranch は現在のブランチ名を取得
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// switchBranch は指定されたブランチに切り替え
func switchBranch(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// popStash はスタッシュを復元
func popStash(stashRef string) error {
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
	// git stash pop は最新のスタッシュを復元するため、
	// 保存したスタッシュが最新であることを前提とする
	cmd = exec.Command("git", "stash", "pop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("スタッシュの適用に失敗: %w", err)
	}

	return nil
}

// printHelp はヘルプメッセージを表示する
func printHelp() {
	help := `git resume - restore a paused worktree

Usage:
  git resume

Description:
  Switches back to the branch recorded by git pause, reapplies the saved stash if any,
  and removes the pause metadata so the workflow can start anew.

Options:
  -h, --help  Show this help message.

Examples:
  git resume

---
git resume - git pause で保存した作業を再開します

使用方法:
  git resume

説明:
  git pause が記録したブランチへ戻り、必要であれば保存された stash を適用して、
  休止状態のメタデータを削除します。

オプション:
  -h, --help  このヘルプを表示します。

使用例:
  git resume
`
	fmt.Print(help)
}


