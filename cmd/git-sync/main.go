package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// ヘルプオプションのチェック
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		printHelp()
		return
	}

	// --continue オプションの処理
	if len(os.Args) > 1 && os.Args[1] == "--continue" {
		if err := continueRebase(); err != nil {
			fmt.Fprintf(os.Stderr, "rebase の続行に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("同期が完了しました。")
		return
	}

	// --abort オプションの処理
	if len(os.Args) > 1 && os.Args[1] == "--abort" {
		if err := abortRebase(); err != nil {
			fmt.Fprintf(os.Stderr, "rebase の中止に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("同期を中止しました。")
		return
	}

	// ターゲットブランチの決定
	targetBranch := ""
	if len(os.Args) > 1 {
		targetBranch = os.Args[1]
	} else {
		// 引数がない場合は main/master を自動判定
		branch, err := detectDefaultBranch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "デフォルトブランチの検出に失敗しました: %v\n", err)
			os.Exit(1)
		}
		targetBranch = branch
	}

	// git fetch origin を実行
	fmt.Printf("origin から最新の変更を取得しています...\n")
	if err := runCommand("git", "fetch", "origin"); err != nil {
		fmt.Fprintf(os.Stderr, "fetch に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// git rebase origin/<ブランチ> を実行
	remoteBranch := fmt.Sprintf("origin/%s", targetBranch)
	fmt.Printf("%s にリベースしています...\n", remoteBranch)
	if err := runCommand("git", "rebase", remoteBranch); err != nil {
		// rebase 中にコンフリクトが発生した場合
		if isRebaseInProgress() {
			fmt.Println("\nコンフリクトが発生しました。")
			fmt.Println("コンフリクトを解決した後、以下のコマンドを実行してください:")
			fmt.Println("  git sync --continue    # 同期を続行")
			fmt.Println("  git sync --abort       # 同期を中止")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "rebase に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("同期が完了しました。(%s)\n", remoteBranch)
}

// detectDefaultBranch は origin のデフォルトブランチ (main または master) を検出する
func detectDefaultBranch() (string, error) {
	// origin/main の存在確認
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/main")
	if err := cmd.Run(); err == nil {
		return "main", nil
	}

	// origin/master の存在確認
	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/master")
	if err := cmd.Run(); err == nil {
		return "master", nil
	}

	return "", fmt.Errorf("origin/main も origin/master も見つかりませんでした")
}

// runCommand は指定されたコマンドを実行し、出力を表示する
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// isRebaseInProgress は rebase が進行中かどうかをチェックする
func isRebaseInProgress() bool {
	// .git/rebase-merge または .git/rebase-apply ディレクトリの存在を確認
	// rebase-merge: 通常の rebase 時に作成される一時ディレクトリ
	// rebase-apply: git am や古い形式の rebase で使用される一時ディレクトリ
	// これらのディレクトリが存在する = rebase が進行中（コンフリクト等で中断している状態）
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return true
	}
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return true
	}
	return false
}

// continueRebase は rebase を続行する
func continueRebase() error {
	return runCommand("git", "rebase", "--continue")
}

// abortRebase は rebase を中止する
func abortRebase() error {
	return runCommand("git", "rebase", "--abort")
}

// printHelp はヘルプメッセージを表示する
func printHelp() {
	help := `git sync - 現在のブランチを最新のリモートブランチと同期

使い方:
  git sync [ブランチ名]      # 指定されたブランチと同期（デフォルト: main/master）
  git sync --continue       # コンフリクト解決後に続行
  git sync --abort          # 同期を中止

説明:
  現在のブランチを最新の origin/<ブランチ> と同期します。
  内部的に git rebase を使用するため、履歴がきれいに保たれます。

引数:
  [ブランチ名]              同期先のブランチ（省略時は main/master を自動判定）

オプション:
  --continue                コンフリクト解決後に rebase を続行
  --abort                   同期を中止して元の状態に戻す
  -h                        このヘルプを表示

例:
  git sync                  # origin/main (または origin/master) と同期
  git sync develop          # origin/develop と同期
  git sync --continue       # コンフリクト解決後に続行
  git sync --abort          # 同期を中止

内部動作:
  1. git fetch origin を実行
  2. 指定されたブランチ（デフォルト: main/master）を自動判定
  3. git rebase origin/<ブランチ> を実行
  4. コンフリクト発生時は修正方法を案内

注意:
  - rebase を使用するため、すでにプッシュ済みのコミットがある場合は
    force push が必要になる可能性があります
  - 共有ブランチでは使用に注意してください
`
	fmt.Print(help)
}
