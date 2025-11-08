package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

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
		fmt.Printf("git push --set-upstream %s %s を実行します...\n\n", remote, remoteBranch)

		// git push --set-upstream を実行
		cmd := exec.Command("git", "push", "--set-upstream", remote, remoteBranch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println("\nプッシュに失敗しました:", err)
			os.Exit(1)
		}

		fmt.Printf("\nブランチ '%s' を '%s' にプッシュし、トラッキングブランチを設定しました。\n", currentBranch, remoteRef)
		return
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

func printHelp() {
	help := `git track - トラッキングブランチを設定

使い方:
  git track                    # origin/<現在のブランチ名> をトラッキング
  git track <リモート名>       # <リモート名>/<現在のブランチ名> をトラッキング
  git track <リモート名> <ブランチ名>  # <リモート名>/<ブランチ名> をトラッキング

説明:
  現在のブランチに対してトラッキングブランチを設定します。
  リモートブランチが存在しない場合は、自動的に
  git push --set-upstream を実行してリモートブランチを作成し、
  トラッキング設定を行います。

オプション:
  -h                    このヘルプを表示

例:
  git track                    # origin/<現在のブランチ> をトラッキング
  git track upstream           # upstream/<現在のブランチ> をトラッキング
  git track origin feature-123 # origin/feature-123 をトラッキング

注意:
  - リモートブランチがない場合は自動でプッシュされます
  - git pull 実行時のトラッキング情報エラーを解決できます
`
	fmt.Print(help)
}
