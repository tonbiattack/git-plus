package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tonbiattack/git-plus/internal/pausestate"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// git-pr-checkout は最新のPRまたは指定されたPRをチェックアウトするコマンド。
// 現在の作業を自動的に保存し、git resume で復元できるようにする。
//
// 主な処理フロー:
//  1. 引数の検証とヘルプ表示。
//  2. GitHub CLI (gh) の利用可否確認。
//  3. PR番号の取得（引数で指定 or 最新のPRを自動取得）。
//  4. 現在の作業を pause と同様に保存（stash + pausestate）。
//  5. gh pr checkout でPRブランチをチェックアウト。
//  6. 元の作業に戻るには git resume を案内。
//
// 使用する主な Git/GitHub CLI コマンド:
//   - gh pr list --limit 1 --json number: 最新のPR番号取得。
//   - gh pr checkout <number>: PRブランチをチェックアウト。
//   - git branch --show-current: 現在のブランチ名取得。
//   - git status --porcelain: 未コミット変更の有無確認。
//   - git stash push: 変更を stash に保存。
func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		os.Exit(0)
	}

	// GitHub CLI の確認
	if !isGitHubCLIAvailable() {
		fmt.Fprintln(os.Stderr, "エラー: GitHub CLI (gh) がインストールされていません")
		fmt.Fprintln(os.Stderr, "インストール方法: https://cli.github.com/")
		os.Exit(1)
	}

	// PR番号を取得
	var prNumber string
	var err error

	if len(os.Args) > 1 {
		// 引数でPR番号が指定されている場合
		prNumber = os.Args[1]
		fmt.Printf("PR #%s をチェックアウトします\n", prNumber)
	} else {
		// 最新のPRを取得
		fmt.Println("最新のPRを取得中...")
		prNumber, err = getLatestPRNumber()
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: 最新のPRの取得に失敗: %v\n", err)
			os.Exit(1)
		}
		if prNumber == "" {
			fmt.Println("エラー: オープンなPRが見つかりません")
			os.Exit(1)
		}
		fmt.Printf("最新のPR #%s をチェックアウトします\n", prNumber)
	}

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
		if !ui.Confirm("上書きしてPRをチェックアウトしますか？", false) {
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
	stashMessage := fmt.Sprintf("git-pr-checkout: from %s", currentBranch)

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

	// PRをチェックアウト
	fmt.Printf("PR #%s をチェックアウト中...\n", prNumber)
	targetBranch, err := checkoutPR(prNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: PRのチェックアウトに失敗: %v\n", err)
		// エラー時はスタッシュを戻す
		if stashRef != "" {
			fmt.Println("スタッシュを復元中...")
			if popErr := popStash(); popErr != nil {
				fmt.Fprintf(os.Stderr, "警告: スタッシュの復元に失敗: %v\n", popErr)
				fmt.Println("手動で復元してください: git stash pop")
			}
		}
		os.Exit(1)
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

	fmt.Printf("✓ PR #%s のブランチ '%s' にチェックアウトしました\n", prNumber, targetBranch)
	fmt.Printf("\n元のブランチに戻るには: git resume\n")
}

// isGitHubCLIAvailable は GitHub CLI が利用可能かチェック
func isGitHubCLIAvailable() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// getLatestPRNumber は最新のPR番号を取得
func getLatestPRNumber() (string, error) {
	cmd := exec.Command("gh", "pr", "list", "--limit", "1", "--json", "number")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// JSONをパース
	var prs []struct {
		Number int `json:"number"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		return "", fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	if len(prs) == 0 {
		return "", nil
	}

	return fmt.Sprintf("%d", prs[0].Number), nil
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

// checkoutPR は指定されたPRをチェックアウトし、チェックアウトしたブランチ名を返す
func checkoutPR(prNumber string) (string, error) {
	cmd := exec.Command("gh", "pr", "checkout", prNumber)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// チェックアウト後のブランチ名を取得
	branch, err := getCurrentBranch()
	if err != nil {
		return "", fmt.Errorf("チェックアウト後のブランチ名取得に失敗: %w", err)
	}

	return branch, nil
}

// popStash はスタッシュを復元
func popStash() error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// printHelp はヘルプメッセージを表示
func printHelp() {
	help := `git pr-checkout - checkout the latest or specified pull request

Usage:
  git pr-checkout [PR番号]

Description:
  Fetches and checks out the latest or specified pull request branch.
  Automatically saves your current work (like git pause) so you can return to it later with git resume.

  When run without arguments, it checks out the latest open PR.
  When given a PR number, it checks out that specific PR.

  Uses GitHub CLI (gh) internally to interact with pull requests.

Arguments:
  [PR番号]    Optional. The PR number to checkout. If omitted, checks out the latest PR.

Options:
  -h, --help  Show this help message.

Examples:
  git pr-checkout          # Checkout the latest PR
  git pr-checkout 123      # Checkout PR #123

Return to your original branch:
  git resume

Prerequisites:
  - GitHub CLI (gh) must be installed and authenticated
  - Install: https://cli.github.com/
  - Authenticate: gh auth login

---
git pr-checkout - 最新または指定されたプルリクエストをチェックアウト

使用方法:
  git pr-checkout [PR番号]

説明:
  最新または指定されたプルリクエストのブランチを取得してチェックアウトします。
  現在の作業を自動的に保存（git pause と同様）するため、後で git resume で戻ることができます。

  引数なしで実行すると、最新のオープンなPRをチェックアウトします。
  PR番号を指定すると、その番号のPRをチェックアウトします。

  内部的に GitHub CLI (gh) を使用してプルリクエストと連携します。

引数:
  [PR番号]    省略可能。チェックアウトするPR番号。省略すると最新のPRをチェックアウト。

オプション:
  -h, --help  このヘルプを表示します。

使用例:
  git pr-checkout          # 最新のPRをチェックアウト
  git pr-checkout 123      # PR #123 をチェックアウト

元のブランチに戻るには:
  git resume

前提条件:
  - GitHub CLI (gh) がインストールされ、認証済みであること
  - インストール: https://cli.github.com/
  - 認証: gh auth login
`
	fmt.Print(help)
}
