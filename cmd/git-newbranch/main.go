package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ブランチ名を指定してください。")
		os.Exit(1)
	}
	branch := os.Args[1]

	exists, err := branchExists(branch)
	if err != nil {
		fmt.Println("ブランチの存在確認に失敗しました:", err)
		os.Exit(1)
	}

	if exists {
		proceed, err := askForConfirmation(branch)
		if err != nil {
			fmt.Println("入力の読み込みに失敗しました:", err)
			os.Exit(1)
		}
		if !proceed {
			fmt.Println("処理を中止しました。")
			return
		}
	}

	delCmd := exec.Command("git", "branch", "-D", branch)
	delCmd.Stdout = os.Stdout
	delCmd.Stderr = os.Stderr
	if err := delCmd.Run(); err != nil && !isNotFound(err) {
		fmt.Println("ブランチの削除に失敗しました:", err)
		os.Exit(1)
	}

	createCmd := exec.Command("git", "checkout", "-b", branch)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		fmt.Println("ブランチ作成に失敗しました:", err)
		os.Exit(1)
	}

	fmt.Printf("ブランチ %s を作成しました。\n", branch)
}

func branchExists(name string) (bool, error) {
	ref := fmt.Sprintf("refs/heads/%s", name)
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func askForConfirmation(branch string) (bool, error) {
	fmt.Printf("ブランチ %s は既に存在します。削除して作り直しますか？ [y/N]: ", branch)
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

func isNotFound(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	return ok && exitErr.ExitCode() == 1
}
