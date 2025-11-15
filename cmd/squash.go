// ================================================================================
// squash.go
// ================================================================================
// このファイルは git の拡張コマンド squash コマンドを実装しています。
//
// 【概要】
// squash コマンドは、複数のコミットを1つにまとめる（スカッシュする）機能を提供します。
// 対話的な UI により、簡単に複数のコミットを統合できます。
//
// 【主な機能】
// - 直近の複数コミットを1つにまとめる
// - 対話的なコミット数の選択（引数なしの場合）
// - コミット数を引数で直接指定（例: git squash 3）
// - 統合前のコミットメッセージ一覧の表示
// - 新しいコミットメッセージの入力
// - 統合前の確認プロンプト
//
// 【使用例】
//   git squash           # 対話的に選択（最近10件を表示）
//   git squash 3         # 直近3つのコミットをスカッシュ
//
// 【内部仕様】
// - git reset --soft HEAD~N でコミットを取り消し
// - 取り消されたコミットの変更は全てステージングエリアに残る
// - 新しいコミットメッセージで git commit を実行
// ================================================================================

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

// commit はコミット情報を表す構造体です。
type commit struct {
	hash    string // コミットのハッシュ値（SHA-1）
	subject string // コミットの件名（1行目のメッセージ）
}

// squashCmd は squash コマンドの定義です。
// 複数のコミットを1つにまとめます。
var squashCmd = &cobra.Command{
	Use:   "squash [コミット数]",
	Short: "複数のコミットをスカッシュ",
	Long: `直近の複数コミットを1つにまとめます。
引数なしで実行すると、最近の10個のコミットを表示し、
スカッシュするコミット数を入力で指定できます。`,
	Example: `  git squash           # 対話的に選択
  git squash 3         # 直近3つのコミットをスカッシュ`,
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

// selectCommitsCount は対話的にスカッシュするコミット数を選択します。
//
// 戻り値:
//   - int: ユーザーが選択したコミット数（0 の場合はキャンセル）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   1. 最近10件のコミットを取得して表示
//   2. ユーザーにコミット数の入力を求める
//   3. 0 または空入力の場合はキャンセル
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

	input = ui.NormalizeNumberInput(input)
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

// getRecentCommitsList は最近のコミット一覧を取得します。
//
// パラメータ:
//   - count: 取得するコミット数
//
// 戻り値:
//   - []commit: コミット情報のスライス（新しい順）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git log --oneline -n <count> --format=%H %s コマンドで
//   コミットハッシュと件名を取得します。
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

// executeSquash は実際のスカッシュ処理を実行します。
//
// パラメータ:
//   - numCommits: スカッシュするコミット数
//   - commits: コミット情報のスライス
//
// 戻り値:
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   1. git reset --soft HEAD~<numCommits> でコミットを取り消す
//   2. 元のコミットメッセージを表示
//   3. ユーザーに新しいコミットメッセージを入力してもらう
//   4. 新しいコミットメッセージで git commit を実行
//
// 備考:
//   git reset --soft を使用するため、変更はステージングエリアに保持されます。
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

// init は squash コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(squashCmd)
}
