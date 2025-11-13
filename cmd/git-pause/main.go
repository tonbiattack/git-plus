package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tonbiattack/git-plus/internal/pausestate"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// git-pause は作業中の変更を一時的に退避し、後で git-resume で復元できるよう
// 別ブランチへ切り替えるためのコマンド。
//
// 主な処理フロー:
//  1. 引数の検証とヘルプ表示。
//  2. 既存の pause 状態があれば上書き可否を確認。
//  3. 現在のブランチ名取得と未コミット変更の検知。
//  4. 変更があれば識別しやすいメッセージで stash へ退避。
//  5. pausestate に元/先ブランチ・stash 情報・タイムスタンプを保存。
//  6. 指定ブランチへスイッチし、復帰方法を案内。
//
// 使用する主な Git コマンド:
//  - git branch --show-current: 現在のブランチ名取得。
//  - git status --porcelain: 未コミット変更の有無を確認。
//  - git stash push / rev-parse: 変更を stash し参照を取得。
//  - git switch <branch>: 対象ブランチへ切り替え。
func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		os.Exit(0)
	}

	targetBranch := os.Args[1]

	// 既に pause 状態かチェック
	exists, err := pausestate.Exists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 状態の確認に失敗: %v\n", err)
		os.Exit(1)
	}

	if exists {
		state, err := pausestate.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: 既存の状態の読み込みに失敗: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("警告: 既に pause 状態です（%s → %s）\n", state.FromBranch, state.ToBranch)

		// 上書きは破壊的操作なのでEnterでno
		if !ui.Confirm("上書きしますか？", false) {
			fmt.Println("キャンセルしました")
			os.Exit(0)
		}
	}

	// 現在のブランチを取得
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 現在のブランチの取得に失敗: %v\n", err)
		os.Exit(1)
	}

	// 変更があるかチェック
	hasChanges, err := hasUncommittedChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 変更の確認に失敗: %v\n", err)
		os.Exit(1)
	}

	var stashRef string
	stashMessage := fmt.Sprintf("git-pause: from %s", currentBranch)

	if hasChanges {
		// スタッシュに保存
		fmt.Printf("変更を保存中...\n")
		stashRef, err = createStash(stashMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: スタッシュの作成に失敗: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ 変更を保存しました: %s\n", stashRef)
	} else {
		fmt.Println("変更がないため、スタッシュはスキップします")
		stashRef = "" // 変更がない場合は空文字列
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
		fmt.Fprintf(os.Stderr, "エラー: 状態の保存に失敗: %v\n", err)
		os.Exit(1)
	}

	// ブランチを切り替え
	fmt.Printf("ブランチを切り替え中: %s → %s\n", currentBranch, targetBranch)
	if err := switchBranch(targetBranch); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: ブランチの切り替えに失敗: %v\n", err)
		// エラー時は状態ファイルを削除
		pausestate.Delete()
		os.Exit(1)
	}

	fmt.Printf("✓ %s に切り替えました\n", targetBranch)
	fmt.Printf("\n元のブランチに戻るには: git resume\n")
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

// hasUncommittedChanges はコミットされていない変更があるかチェック
func hasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// createStash はスタッシュを作成し、スタッシュ参照を返す
func createStash(message string) (string, error) {
	cmd := exec.Command("git", "stash", "push", "-m", message)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// 最新のスタッシュ参照を取得
	cmd = exec.Command("git", "rev-parse", "stash@{0}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// switchBranch は指定されたブランチまたはタグに切り替え
func switchBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// printHelp はヘルプメッセージを表示
func printHelp() {
	help := `git pause - stash work in progress and hop to another branch

Usage:
  git pause <branch>

Description:
  Saves the current branch name and uncommitted changes, then switches to the specified branch.
  Use git resume to restore the saved work when you are ready to return.

Arguments:
  <branch>    Branch to switch to after pausing the current work.

Options:
  -h, --help  Show this help message.

Examples:
  git pause main
  git pause feature/login

Resume the paused branch with:
  git resume

---
git pause - 作業中の変更を stash して別ブランチへ移動します

使用方法:
  git pause <branch>

説明:
  現在のブランチ名とコミットしていない変更を保存し、指定したブランチへ切り替えます。
  作業へ戻る準備ができたら git resume を実行して保存内容を復元してください。

引数:
  <branch>    休止後に切り替えるブランチ名。

オプション:
  -h, --help  このヘルプを表示します。

使用例:
  git pause main
  git pause feature/login

休止したブランチへ戻るには:
  git resume
`
	fmt.Print(help)
}

