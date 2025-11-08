package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type BranchInfo struct {
	Name         string
	LastCommitAt string
}

func main() {
	fmt.Println("最近使用したブランチを取得しています...")

	branches, err := getRecentBranches()
	if err != nil {
		fmt.Printf("エラー: ブランチ一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(branches) == 0 {
		fmt.Println("ブランチが見つかりませんでした。")
		os.Exit(0)
	}

	// 現在のブランチを取得
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Printf("警告: 現在のブランチの取得に失敗しました: %v\n", err)
	}

	// ブランチ一覧を表示
	fmt.Println("\n最近使用したブランチ:")
	displayCount := 0
	for _, branch := range branches {
		// 現在のブランチはスキップ
		if branch.Name == currentBranch {
			continue
		}
		displayCount++
		fmt.Printf("%d. %s\n", displayCount, branch.Name)

		// 最大10件まで表示
		if displayCount >= 10 {
			break
		}
	}

	if displayCount == 0 {
		fmt.Println("切り替え可能なブランチがありません。")
		os.Exit(0)
	}

	// ブランチ選択
	fmt.Print("\nSelect branch (番号を入力): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("エラー: 入力の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("キャンセルしました。")
		os.Exit(0)
	}

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > displayCount {
		fmt.Printf("エラー: 無効な番号です。1から%dの範囲で入力してください。\n", displayCount)
		os.Exit(1)
	}

	// 選択されたブランチを取得（現在のブランチをスキップした番号に対応）
	selectedBranch := ""
	count := 0
	for _, branch := range branches {
		if branch.Name == currentBranch {
			continue
		}
		count++
		if count == selection {
			selectedBranch = branch.Name
			break
		}
	}

	if selectedBranch == "" {
		fmt.Println("エラー: ブランチの選択に失敗しました。")
		os.Exit(1)
	}

	// ブランチを切り替え
	fmt.Printf("\nブランチ '%s' に切り替えています...\n", selectedBranch)
	if err := switchBranch(selectedBranch); err != nil {
		fmt.Printf("エラー: ブランチの切り替えに失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ ブランチ '%s' に切り替えました。\n", selectedBranch)
}

// getRecentBranches は最近使用したブランチを取得する
func getRecentBranches() ([]BranchInfo, error) {
	// git for-each-ref で最終コミット日時順にブランチを取得
	cmd := exec.Command("git", "for-each-ref",
		"--sort=-committerdate",
		"--format=%(refname:short)|%(committerdate:relative)",
		"refs/heads/")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]BranchInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		branches = append(branches, BranchInfo{
			Name:         parts[0],
			LastCommitAt: parts[1],
		})
	}

	return branches, nil
}

// getCurrentBranch は現在のブランチ名を取得する
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// switchBranch は指定したブランチに切り替える
func switchBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
