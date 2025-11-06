package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("git", "reset", "--soft", "HEAD^")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("コミットの取り消しに失敗しました:", err)
		os.Exit(1)
	}

	fmt.Println("最後のコミットを取り消しました（変更は残っています）")
}
