package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := append([]string{"commit", "--amend"}, os.Args[1:]...)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "git commit --amend の実行に失敗しました: %v\n", err)
		os.Exit(1)
	}
}
