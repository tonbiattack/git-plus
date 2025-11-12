package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tonbiattack/git-plus/internal/ui"
)

// main はPRの作成からマージ、ブランチ削除、pull までを一気に行うメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. 現在のブランチを取得
//  3. ベースブランチをユーザーに確認（引数で指定も可能）
//  4. PRを作成（タイトル・本文なしで--fillオプション使用）
//  5. PRを自動マージし、ブランチ削除
//  6. ベースブランチに切り替え
//  7. 最新の変更を取得（git pull）
//
// 終了コード:
//   - 0: 正常終了
//   - 1: エラー発生
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// 現在のブランチを取得
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Printf("エラー: 現在のブランチの取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if currentBranch == "" {
		fmt.Println("エラー: ブランチが見つかりません。detached HEAD 状態かもしれません。")
		os.Exit(1)
	}

	fmt.Printf("現在のブランチ: %s\n", currentBranch)

	// ベースブランチの取得
	var baseBranch string
	if len(os.Args) > 1 && os.Args[1] != "-h" {
		// 引数で指定された場合
		baseBranch = os.Args[1]
	} else {
		// 引数がない場合は入力を受け付ける
		fmt.Print("マージ先のベースブランチを入力してください (デフォルト: main): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("エラー: 入力の読み込みに失敗しました: %v\n", err)
			os.Exit(1)
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
		os.Exit(0)
	}

	// Step 1: PRを作成
	fmt.Println("\n[1/5] PRを作成しています...")
	if err := createPR(baseBranch, currentBranch); err != nil {
		fmt.Printf("エラー: PRの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ PRを作成しました")

	// Step 2: PRをマージしてブランチを削除
	fmt.Println("\n[2/5] PRをマージしてブランチを削除しています...")
	if err := mergePR(); err != nil {
		fmt.Printf("エラー: PRのマージに失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ PRをマージしてブランチを削除しました")

	// Step 3: ベースブランチに切り替え
	fmt.Printf("\n[3/5] ブランチ '%s' に切り替えています...\n", baseBranch)
	if err := switchBranch(baseBranch); err != nil {
		fmt.Printf("エラー: ブランチの切り替えに失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ ブランチ '%s' に切り替えました\n", baseBranch)

	// Step 4: 最新の変更を取得
	fmt.Println("\n[4/5] 最新の変更を取得しています...")
	if err := gitPull(); err != nil {
		fmt.Printf("エラー: git pull に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 最新の変更を取得しました")

	fmt.Println("\n✓ すべての処理が完了しました！")
}

// getCurrentBranch は現在のブランチ名を取得する
//
// 戻り値:
//   - string: 現在のブランチ名
//   - error: git コマンドの実行エラー
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// createPR はPRを作成する
//
// パラメータ:
//   - base: マージ先のベースブランチ
//   - head: マージ元のヘッドブランチ
//
// 戻り値:
//   - error: gh コマンドの実行エラー
func createPR(base, head string) error {
	cmd := exec.Command("gh", "pr", "create",
		"--base", base,
		"--head", head,
		"--fill")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// mergePR はPRをマージしてブランチを削除する
//
// 戻り値:
//   - error: gh コマンドの実行エラー
func mergePR() error {
	cmd := exec.Command("gh", "pr", "merge",
		"--merge",
		"--delete-branch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// switchBranch は指定したブランチに切り替える
//
// パラメータ:
//   - branch: 切り替え先のブランチ名
//
// 戻り値:
//   - error: git コマンドの実行エラー
func switchBranch(branch string) error {
	cmd := exec.Command("git", "switch", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// gitPull は git pull を実行する
//
// 戻り値:
//   - error: git コマンドの実行エラー
func gitPull() error {
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// printHelp はコマンドのヘルプメッセージを表示する
func printHelp() {
	help := `git pr-merge - PRの作成からマージ、ブランチ削除までを一気に実行

使い方:
  git pr-merge [ベースブランチ名]

引数:
  ベースブランチ名      マージ先のブランチ名（省略時は対話的に入力）

説明:
  以下の処理を自動的に実行します:
  1. タイトル・本文なしでPRを作成（--fillオプション使用）
  2. PRをマージしてブランチを削除（--merge --delete-branch）
  3. ベースブランチに切り替え（git switch）
  4. 最新の変更を取得（git pull）

オプション:
  -h                    このヘルプを表示

使用例:
  git pr-merge              # 対話的にベースブランチを入力
  git pr-merge main         # mainブランチへマージ
  git pr-merge develop      # developブランチへマージ

使用方法:
  1. 作業ブランチから git pr-merge を実行
  2. ベースブランチを引数で指定するか、入力（デフォルト: main）
  3. 確認メッセージで y を入力
  4. 自動的にPR作成→マージ→ブランチ切り替え→pull を実行

注意事項:
  - GitHub CLI (gh) がインストールされている必要があります
  - gh auth login でログイン済みである必要があります
  - リモートリポジトリへのプッシュ権限が必要です
`
	fmt.Print(help)
}
