package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// 現在のブランチ名を取得
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Println("現在のブランチの取得に失敗しました:", err)
		os.Exit(1)
	}

	// リモート名を引数から取得、デフォルトは origin
	remote := "origin"
	if len(os.Args) >= 2 {
		remote = os.Args[1]
	}

	// ブランチ名を引数から取得、デフォルトは現在のブランチ名
	remoteBranch := currentBranch
	if len(os.Args) >= 3 {
		remoteBranch = os.Args[2]
	}

	// リモートブランチが存在するか確認
	remoteRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
	exists, err := remoteRefExists(remoteRef)
	if err != nil {
		fmt.Println("リモートブランチの確認に失敗しました:", err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("リモートブランチ %s が見つかりません。\n", remoteRef)
		fmt.Println("git fetch を実行してリモートの最新情報を取得してください。")
		os.Exit(1)
	}

	// upstream を設定
	upstreamRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
	cmd := exec.Command("git", "branch", "--set-upstream-to="+upstreamRef, currentBranch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("トラッキングブランチの設定に失敗しました:", err)
		os.Exit(1)
	}

	fmt.Printf("ブランチ '%s' のトラッキングブランチを '%s' に設定しました。\n", currentBranch, upstreamRef)
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func remoteRefExists(ref string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/%s", ref))
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
