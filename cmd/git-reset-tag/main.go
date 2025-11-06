package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
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
