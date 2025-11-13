package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// commit はコミット情報を表す構造体
type commit struct {
	hash    string
	subject string
}

var squashCmd = &cobra.Command{
	Use:   "squash [コミット数]",
	Short: "複数のコミットをスカッシュ",
	Long: `直近の複数コミットを1つにまとめます。
引数なしで実行すると、最近の10個のコミットを表示し、
スカッシュするコミット数を入力で指定できます。`,
	Example: `  git-plus squash           # 対話的に選択
  git-plus squash 3         # 直近3つのコミットをスカッシュ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var numCommits int
		var err error

		if len(args) >= 1 {
			numCommits, err = strconv.Atoi(args[0])
			if err != nil || numCommits <= 0 {
				return fmt.Errorf("不正なコミット数です: %s", args[0])
			}
		} else {
			// 引数がない場合は対話的に決定
			numCommits, err = selectCommitsCount()
			if err != nil {
				return err
			}
			if numCommits == 0 {
				fmt.Println("スカッシュを中止しました。")
				return nil
			}
		}

		if numCommits < 2 {
			return fmt.Errorf("スカッシュするには2つ以上のコミットが必要です")
		}

		commits, err := getRecentCommitsList(numCommits)
		if err != nil {
			return fmt.Errorf("コミット履歴の取得に失敗しました: %w", err)
		}

		if len(commits) < numCommits {
			return fmt.Errorf("指定された数のコミットが存在しません（実際: %d個）", len(commits))
		}

		fmt.Printf("以下の %d 個のコミットをスカッシュします:\n", numCommits)
		for i, c := range commits {
			fmt.Printf("  %d. %s %s\n", i+1, c.hash[:8], c.subject)
		}

		if !ui.Confirm("実行しますか？", false) {
			fmt.Println("スカッシュを中止しました。")
			return nil
		}

		if err := executeSquash(numCommits, commits); err != nil {
			return fmt.Errorf("スカッシュに失敗しました: %w", err)
		}

		return nil
	},
}

func selectCommitsCount() (int, error) {
	commits, err := getRecentCommitsList(10)
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

func getRecentCommitsList(count int) ([]commit, error) {
	output, err := gitcmd.Run("log", "--oneline", "-n", strconv.Itoa(count), "--format=%H %s")
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

func executeSquash(numCommits int, commits []commit) error {
	// git reset --soft を使用してコミットを取り消し
	resetTarget := fmt.Sprintf("HEAD~%d", numCommits)
	if err := gitcmd.RunQuiet("reset", "--soft", resetTarget); err != nil {
		return fmt.Errorf("git reset --soft の実行に失敗しました: %w", err)
	}

	// 既存のコミットメッセージを表示
	fmt.Println("\n元のコミットメッセージ:")
	for i := len(commits) - 1; i >= 0; i-- {
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
	if err := gitcmd.RunWithIO("commit", "-m", newMessage); err != nil {
		return fmt.Errorf("新しいコミットの作成に失敗しました: %w", err)
	}

	fmt.Printf("スカッシュが完了しました。%d個のコミットが1つにまとめられました。\n", numCommits)
	return nil
}

func init() {
	rootCmd.AddCommand(squashCmd)
}
