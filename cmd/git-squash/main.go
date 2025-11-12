package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// commit はコミット情報を表す構造体
//
// フィールド:
//  - hash: コミットハッシュ（SHA-1）。例: "a1b2c3d4e5f6..."
//  - subject: コミットメッセージの1行目（サブジェクト）。例: "Add new feature"
//
// 使用箇所:
//  - getRecentCommits: コミット履歴の取得結果を格納
//  - runSquash: スカッシュ対象のコミット情報を表示
type commit struct {
	hash    string
	subject string
}

// main は複数のコミットを1つにスカッシュするメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. コミット数の取得（引数指定 or 対話的選択）
//  3. 最近のコミット履歴を取得
//  4. スカッシュ対象のコミットを表示
//  5. ユーザー確認を求める
//  6. git reset --soft でコミットを取り消す
//  7. 新しいコミットメッセージを入力
//  8. 1つの新しいコミットを作成
//
// 使用するgitコマンド:
//  - git log --oneline -n <数> --format=%H %s: コミット履歴を取得
//  - git reset --soft HEAD~<数>: 指定数のコミットを取り消し（変更は保持）
//  - git commit -m <メッセージ>: 新しいコミットを作成
//
// 実装の詳細:
//  - 引数なしの場合は対話的にコミット数を選択可能
//  - git reset --soft を使用することで変更内容は保持される
//  - 元のコミットメッセージを参照として表示
//
// 終了コード:
//  - 0: 正常終了（スカッシュ成功またはキャンセル）
//  - 1: エラー発生（引数不正、コミット取得失敗、スカッシュ失敗など）
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

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

	// 破壊的操作なのでEnterでno
	if !ui.Confirm("実行しますか？", false) {
		fmt.Println("スカッシュを中止しました。")
		return
	}

	if err := runSquash(numCommits, commits); err != nil {
		fmt.Fprintln(os.Stderr, "スカッシュに失敗しました:", err)
		os.Exit(1)
	}
}

// selectCommitsInteractively は対話的にスカッシュするコミット数を選択する
//
// 戻り値:
//  - int: ユーザーが選択したコミット数（0の場合はキャンセル）
//  - error: エラー情報（コミット履歴取得失敗、入力エラーなど）
//
// 使用するgitコマンド:
//  - getRecentCommits 経由で git log を実行
//
// 実装の詳細:
//  - 最近の10件のコミットを表示
//  - ユーザーに数値入力を促す
//  - 0が入力された場合はキャンセル扱い
//  - 範囲外の値が入力された場合はエラーを返す
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

// getRecentCommits は最近のコミット履歴を取得する
//
// パラメータ:
//  - count: 取得するコミット数
//
// 戻り値:
//  - []commit: コミット情報のスライス（新しい順）
//  - error: エラー情報（git log の実行失敗、パースエラーなど）
//
// 使用するgitコマンド:
//  - git log --oneline -n <count> --format=%H %s: ハッシュとサブジェクトを取得
//
// 実装の詳細:
//  - %H: フルハッシュ（40文字）
//  - %s: コミットメッセージの1行目
//  - 出力フォーマット: "<ハッシュ> <サブジェクト>"
//  - 空行は無視
func getRecentCommits(count int) ([]commit, error) {
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

// runSquash はコミットのスカッシュを実行する
//
// パラメータ:
//  - numCommits: スカッシュするコミット数
//  - commits: 対象のコミット情報スライス
//
// 戻り値:
//  - error: エラー情報（git reset 失敗、コミット作成失敗など）
//
// 使用するgitコマンド:
//  - git reset --soft HEAD~<数>: 指定数のコミットを取り消し（変更は保持）
//  - git commit -m <メッセージ>: 新しいコミットを作成
//
// 実装の詳細:
//  - git reset --soft を使用して変更内容を保持したままコミットのみ取り消す
//  - 元のコミットメッセージを参考として表示（古い順）
//  - ユーザーに新しいコミットメッセージの入力を求める
//  - 空のメッセージはエラーとして扱う
func runSquash(numCommits int, commits []commit) error {
	// git reset --soft を使用してコミットを取り消し
	resetTarget := fmt.Sprintf("HEAD~%d", numCommits)
	if err := gitcmd.RunQuiet("reset", "--soft", resetTarget); err != nil {
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
	if err := gitcmd.RunWithIO("commit", "-m", newMessage); err != nil {
		return fmt.Errorf("新しいコミットの作成に失敗しました: %w", err)
	}

	fmt.Printf("スカッシュが完了しました。%d個のコミットが1つにまとめられました。\n", numCommits)
	return nil
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - コミットのスカッシュ機能の使い方を説明
//  - 対話的モードと引数指定モードの両方を説明
//  - 最低2つ以上のコミットが必要であることを明記
func printHelp() {
	help := `git squash - 複数のコミットをスカッシュ

使い方:
  git squash           # 対話的にコミット数を選択
  git squash <数>      # 指定した数のコミットをスカッシュ

説明:
  直近の複数コミットを1つにまとめます。
  引数なしで実行すると、最近の10個のコミットを表示し、
  スカッシュするコミット数を入力で指定できます。

オプション:
  -h                   このヘルプを表示

例:
  git squash           # 対話的に選択
  git squash 3         # 直近3つのコミットをスカッシュ

注意:
  - 最低2つ以上のコミットが必要です
  - 新しいコミットメッセージの入力が必要です
`
	fmt.Print(help)
}
