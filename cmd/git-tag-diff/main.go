package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("使用方法: git tag-diff <古いタグ> <新しいタグ>")
		fmt.Println("例: git tag-diff V4.2.00.00 V4.3.00.00")
		os.Exit(1)
	}

	oldTag := os.Args[1]
	newTag := os.Args[2]

	// 出力ファイル名の自動生成
	outputFile := fmt.Sprintf("tag_diff_%s_to_%s.txt", oldTag, newTag)

	// タグの存在確認
	if err := validateTag(oldTag); err != nil {
		fmt.Printf("エラー: タグ '%s' が存在しません\n", oldTag)
		os.Exit(1)
	}
	if err := validateTag(newTag); err != nil {
		fmt.Printf("エラー: タグ '%s' が存在しません\n", newTag)
		os.Exit(1)
	}

	// git logコマンドの実行
	tagRange := fmt.Sprintf("%s..%s", oldTag, newTag)
	cmd := exec.Command("git", "log", tagRange, "--no-merges", "--pretty=format:- %s (%an, %ad)", "--date=iso")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("git logの実行に失敗しました: %s\n", string(exitErr.Stderr))
		} else {
			fmt.Printf("git logの実行に失敗しました: %v\n", err)
		}
		os.Exit(1)
	}

	// 出力が空の場合
	if len(output) == 0 {
		fmt.Printf("タグ %s と %s の間に差分はありません。\n", oldTag, newTag)
		os.Exit(0)
	}

	// ファイルに書き込み
	absPath, err := filepath.Abs(outputFile)
	if err != nil {
		fmt.Printf("ファイルパスの取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		fmt.Printf("ファイルへの書き込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// サマリー表示
	commits := strings.Split(string(output), "\n")
	fmt.Printf("✓ タグ %s と %s の差分を %s に出力しました。\n", oldTag, newTag, absPath)
	fmt.Printf("  コミット数: %d\n", len(commits))
}

// validateTag はタグが存在するかを確認する
func validateTag(tag string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", tag)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
