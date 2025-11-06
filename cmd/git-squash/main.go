package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type commit struct {
	hash    string
	subject string
}

func main() {
	var numCommits int
	var err error

	if len(os.Args) >= 2 {
		numCommits, err = strconv.Atoi(os.Args[1])
		if err != nil || numCommits <= 0 {
			fmt.Printf("不正なコミット数です: %s\n", os.Args[1])
			os.Exit(1)
		}
	} else {
		// 引数がない場合は対話的に決定
		numCommits, err = selectCommitsInteractively()
		if err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(1)
		}
		if numCommits == 0 {
			fmt.Println("スカッシュを中止しました。")
			return
		}
	}

	if numCommits < 2 {
		fmt.Println("スカッシュするには2つ以上のコミットが必要です。")
		os.Exit(1)
	}

	commits, err := getRecentCommits(numCommits)
	if err != nil {
		fmt.Fprintln(os.Stderr, "コミット履歴の取得に失敗しました:", err)
		os.Exit(1)
	}

	if len(commits) < numCommits {
		fmt.Printf("指定された数のコミットが存在しません（実際: %d個）。\n", len(commits))
		os.Exit(1)
	}

	fmt.Printf("以下の %d 個のコミットをスカッシュします:\n", numCommits)
	for i, c := range commits {
		fmt.Printf("  %d. %s %s\n", i+1, c.hash[:8], c.subject)
	}

	proceed, err := askForConfirmation()
	if err != nil {
		fmt.Fprintln(os.Stderr, "入力の読み込みに失敗しました:", err)
		os.Exit(1)
	}
	if !proceed {
		fmt.Println("スカッシュを中止しました。")
		return
	}

	if err := runSquash(numCommits, commits); err != nil {
		fmt.Fprintln(os.Stderr, "スカッシュに失敗しました:", err)
		os.Exit(1)
	}
}

func selectCommitsInteractively() (int, error) {
	commits, err := getRecentCommits(10) // 最近の10件を表示
	if err != nil {
		return 0, fmt.Errorf("コミット履歴の取得に失敗しました: %w", err)
	}

	if len(commits) < 2 {
		return 0, fmt.Errorf("スカッシュ可能なコミットが不足しています（最低2つ必要）")
	}

	fmt.Println("最近のコミット:")
	maxDisplay := len(commits)
	if maxDisplay > 10 {
		maxDisplay = 10
	}
	for i := 0; i < maxDisplay; i++ {
		c := commits[i]
		fmt.Printf("  %d. %s %s\n", i+1, c.hash[:8], c.subject)
	}

	fmt.Print("\nスカッシュするコミット数を入力してください (2以上、0で中止): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			input = "0"
		} else {
			return 0, fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}
	}

	input = strings.TrimSpace(input)
	if input == "0" || input == "" {
		return 0, nil
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("数値として解釈できません: %s", input)
	}

	if num < 0 {
		return 0, nil
	}

	if num > len(commits) {
		return 0, fmt.Errorf("指定された数が利用可能なコミット数を超えています（最大: %d）", len(commits))
	}

	return num, nil
}

func getRecentCommits(count int) ([]commit, error) {
	cmd := exec.Command("git", "log", "--oneline", "-n", strconv.Itoa(count), "--format=%H %s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log の実行に失敗しました: %w", err)
	}

	var commits []commit
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		commits = append(commits, commit{
			hash:    parts[0],
			subject: parts[1],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("コミット履歴の解析に失敗しました: %w", err)
	}

	return commits, nil
}

func askForConfirmation() (bool, error) {
	fmt.Print("実行しますか？ (y/N): ")
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

func runSquash(numCommits int, commits []commit) error {
	// git reset --soft を使用してコミットを取り消し
	resetTarget := fmt.Sprintf("HEAD~%d", numCommits)
	cmd := exec.Command("git", "reset", "--soft", resetTarget)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git reset --soft の実行に失敗しました: %w", err)
	}

	// 既存のコミットメッセージを表示
	fmt.Println("\n元のコミットメッセージ:")
	for i := len(commits) - 1; i >= 0; i-- { // 古い順に表示
		c := commits[i]
		fmt.Printf("  - %s\n", c.subject)
	}

	// ユーザーから新しいコミットメッセージを取得
	fmt.Print("\n新しいコミットメッセージを入力してください: ")
	reader := bufio.NewReader(os.Stdin)
	newMessage, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("コミットメッセージの読み込みに失敗しました: %w", err)
	}

	newMessage = strings.TrimSpace(newMessage)
	if newMessage == "" {
		return fmt.Errorf("コミットメッセージが空です")
	}

	// 新しいコミットを作成
	commitCmd := exec.Command("git", "commit", "-m", newMessage)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr

	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("新しいコミットの作成に失敗しました: %w", err)
	}

	fmt.Printf("スカッシュが完了しました。%d個のコミットが1つにまとめられました。\n", numCommits)
	return nil
}
