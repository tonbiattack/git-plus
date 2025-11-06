package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	branches, err := mergedBranches()
	if err != nil {
		fmt.Println("マージ済みブランチの取得に失敗しました:", err)
		os.Exit(1)
	}

	if len(branches) == 0 {
		fmt.Println("削除対象のブランチはありません。")
		return
	}

	fmt.Println("以下のブランチを削除します:")
	for _, b := range branches {
		fmt.Println(b)
	}

	proceed, err := askForConfirmation()
	if err != nil {
		fmt.Println("入力の読み込みに失敗しました:", err)
		os.Exit(1)
	}
	if !proceed {
		fmt.Println("キャンセルしました。")
		return
	}

	var deleteErrors bool
	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "-d", branch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			deleteErrors = true
			fmt.Fprintf(os.Stderr, "ブランチ %s の削除に失敗しました: %v\n", branch, err)
		}
	}

	if deleteErrors {
		fmt.Println("一部のブランチの削除に失敗しました。")
		os.Exit(1)
	}

	fmt.Println("削除しました。")
}

func mergedBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--merged")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	var branches []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		}
		if line == "" {
			continue
		}
		if shouldSkipBranch(line) {
			continue
		}
		branches = append(branches, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return branches, nil
}

func shouldSkipBranch(branch string) bool {
	switch branch {
	case "main", "master", "develop":
		return true
	default:
		return false
	}
}

func askForConfirmation() (bool, error) {
	fmt.Print("本当に削除しますか？ (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			input = ""
		} else {
			return false, err
		}
	}

	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes", nil
}
