package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	if len(os.Args) < 2 {
		fmt.Println("タグ名を指定してください。")
		os.Exit(1)
	}
	tagName := os.Args[1]

	// ローカルタグ削除（既に存在しない場合があるためエラーは無視）
	if err := runGitCommand(true, "tag", "-d", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "ローカルタグの削除で予期せぬエラーが発生しました: %v\n", err)
	}

	// リモートタグ削除（存在しないこともあるので警告として扱う）
	if err := runGitCommand(true, "push", "--delete", "origin", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "リモートタグの削除に失敗しました: %v\n", err)
	}

	// 最新コミットにタグを再付与
	if err := runGitCommand(false, "tag", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "タグの再作成に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// リモートにタグをプッシュ
	if err := runGitCommand(false, "push", "origin", tagName); err != nil {
		fmt.Fprintf(os.Stderr, "タグのプッシュに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("タグ %s をリセットして再作成しました。\n", tagName)
}

func runGitCommand(ignoreError bool, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ignoreError {
			return nil
		}
		return err
	}

	return nil
}

func printHelp() {
	help := `git reset-tag - タグをリセットして再作成

使い方:
  git reset-tag <タグ名>

説明:
  指定したタグをローカルとリモートから削除し、
  最新コミットに同名のタグを再作成してリモートにプッシュします。

オプション:
  -h                    このヘルプを表示

例:
  git reset-tag v1.2.3    # v1.2.3 タグをリセット

注意:
  - ローカルとリモート（origin）のタグが削除されます
  - 最新コミットに新しいタグが作成されます
`
	fmt.Print(help)
}
