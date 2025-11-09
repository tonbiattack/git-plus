package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// BranchInfo はブランチの情報を保持する構造体
//
// フィールド:
//   - Name: ブランチ名（例: "feature/awesome", "main"）
//   - LastCommitAt: 最終コミット日時の相対表記（例: "2 hours ago", "3 days ago"）
//
// この構造体は git for-each-ref の出力をパースして生成される
type BranchInfo struct {
	Name         string // ブランチ名
	LastCommitAt string // 最終コミット日時（相対表記）
}

// main は最近使用したブランチを表示し、選択したブランチに切り替えるメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. 最近コミットがあったブランチを最終コミット日時順に取得
//  3. 現在のブランチを除外してブランチ一覧を表示（最大10件）
//  4. ユーザーに番号入力を促し、選択されたブランチに git checkout で切り替え
//
// 表示されるブランチ:
//  - 最近コミットがあったブランチを新しい順に表示
//  - 現在のブランチは自動的に除外される
//  - 最大10件まで表示
//
// 終了コード:
//  - 0: 正常終了（切り替え成功またはキャンセル）
//  - 1: エラー発生（ブランチ取得失敗、切り替え失敗など）
func main() {
	// -h オプションのチェック
	// コマンドライン引数に -h が含まれている場合はヘルプを表示して終了
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	fmt.Println("最近使用したブランチを取得しています...")

	// 最近コミットがあったブランチを最終コミット日時順（新しい順）に取得
	branches, err := getRecentBranches()
	if err != nil {
		fmt.Printf("エラー: ブランチ一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// ブランチが1つも見つからない場合は終了
	if len(branches) == 0 {
		fmt.Println("ブランチが見つかりませんでした。")
		os.Exit(0)
	}

	// 現在のブランチを取得
	// 現在のブランチは一覧から除外するため
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Printf("警告: 現在のブランチの取得に失敗しました: %v\n", err)
		// 警告のみで処理は継続
	}

	// ブランチ一覧を表示
	// 現在のブランチを除外し、最大10件まで番号付きで表示
	fmt.Println("\n最近使用したブランチ:")
	displayCount := 0
	for _, branch := range branches {
		// 現在のブランチはスキップ
		if branch.Name == currentBranch {
			continue
		}
		displayCount++
		fmt.Printf("%d. %s\n", displayCount, branch.Name)

		// 最大10件まで表示
		if displayCount >= 10 {
			break
		}
	}

	// 現在のブランチを除外した結果、表示可能なブランチがない場合は終了
	if displayCount == 0 {
		fmt.Println("切り替え可能なブランチがありません。")
		os.Exit(0)
	}

	// ブランチ選択
	// ユーザーに番号入力を促し、標準入力から1行読み取る
	fmt.Print("\nSelect branch (番号を入力): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("エラー: 入力の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 入力の前後の空白（改行含む）を削除
	input = strings.TrimSpace(input)
	// 空入力の場合はキャンセルとして扱う
	if input == "" {
		fmt.Println("キャンセルしました。")
		os.Exit(0)
	}

	// 入力を整数に変換し、範囲チェック
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > displayCount {
		fmt.Printf("エラー: 無効な番号です。1から%dの範囲で入力してください。\n", displayCount)
		os.Exit(1)
	}

	// 選択されたブランチを取得
	// 表示時と同じロジックで、現在のブランチをスキップしながら番号をカウント
	selectedBranch := ""
	count := 0
	for _, branch := range branches {
		if branch.Name == currentBranch {
			continue
		}
		count++
		if count == selection {
			selectedBranch = branch.Name
			break
		}
	}

	// 番号に対応するブランチが見つからない場合（通常は発生しない）
	if selectedBranch == "" {
		fmt.Println("エラー: ブランチの選択に失敗しました。")
		os.Exit(1)
	}

	// 選択されたブランチに切り替え
	// git checkout コマンドを実行
	fmt.Printf("\nブランチ '%s' に切り替えています...\n", selectedBranch)
	if err := switchBranch(selectedBranch); err != nil {
		fmt.Printf("エラー: ブランチの切り替えに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 切り替え成功メッセージを表示
	fmt.Printf("✓ ブランチ '%s' に切り替えました。\n", selectedBranch)
}

// getRecentBranches は最近使用したブランチを取得する
//
// git for-each-ref コマンドを使用して、ローカルブランチを最終コミット日時順（新しい順）に取得する。
//
// 戻り値:
//  - []BranchInfo: ブランチ情報のスライス（最新のコミット順にソート済み）
//  - error: git コマンドの実行エラー
//
// 実装の詳細:
//  使用する git コマンド:
//    git for-each-ref --sort=-committerdate \
//      --format=%(refname:short)|%(committerdate:relative) \
//      refs/heads/
//
//  オプションの説明:
//    --sort=-committerdate: 最終コミット日時の降順（新しい順）でソート
//    --format: 出力フォーマットを "ブランチ名|相対日時" に指定
//    refs/heads/: ローカルブランチのみを対象（リモートブランチは除外）
//
//  出力例:
//    feature/awesome|2 hours ago
//    main|3 days ago
//    develop|1 week ago
func getRecentBranches() ([]BranchInfo, error) {
	// git for-each-ref で最終コミット日時順にブランチを取得
	cmd := exec.Command("git", "for-each-ref",
		"--sort=-committerdate",                                 // 最終コミット日時の降順（新しい順）
		"--format=%(refname:short)|%(committerdate:relative)", // ブランチ名|相対日時
		"refs/heads/")                                           // ローカルブランチのみ

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 出力を行ごとに分割
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]BranchInfo, 0, len(lines))

	// 各行をパースしてBranchInfo構造体に変換
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// "ブランチ名|相対日時" の形式で分割
		// 例: "feature/awesome|2 hours ago" → ["feature/awesome", "2 hours ago"]
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			// フォーマットが期待と異なる場合はスキップ
			continue
		}

		branches = append(branches, BranchInfo{
			Name:         parts[0], // ブランチ名
			LastCommitAt: parts[1], // 最終コミット日時（相対表記）
		})
	}

	return branches, nil
}

// getCurrentBranch は現在のブランチ名を取得する
//
// git branch --show-current コマンドを実行して、現在チェックアウト中のブランチ名を取得する。
//
// 戻り値:
//  - string: 現在のブランチ名（例: "main", "feature/awesome"）
//  - error: git コマンドの実行エラー
//
// 注意:
//  - detached HEAD 状態の場合は空文字列を返す
//  - 出力の前後の空白（改行含む）は自動的に削除される
//
// 使用する git コマンド:
//  git branch --show-current
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// 前後の空白（改行含む）を削除して返す
	return strings.TrimSpace(string(output)), nil
}

// switchBranch は指定したブランチに切り替える
//
// git checkout コマンドを実行して、指定されたブランチにワーキングツリーを切り替える。
//
// パラメータ:
//  - branch: 切り替え先のブランチ名（例: "main", "feature/awesome"）
//
// 戻り値:
//  - error: git コマンドの実行エラー（ブランチが存在しない、未コミットの変更がある等）
//
// 動作:
//  - git checkout の標準出力と標準エラー出力は、そのまま親プロセスの出力にリダイレクトされる
//  - これにより、git が出力するメッセージ（例: "Switched to branch 'main'"）がユーザーに表示される
//
// 使用する git コマンド:
//  git checkout <branch>
func switchBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout // git の標準出力をそのまま表示
	cmd.Stderr = os.Stderr // git のエラー出力もそのまま表示
	return cmd.Run()
}

// printHelp はコマンドのヘルプメッセージを表示する
//
// 使い方、説明、オプション、使用手順を含む詳細なヘルプを標準出力に表示する。
// -h オプションが指定された場合や、使い方が分からない場合にユーザーに情報を提供する。
func printHelp() {
	help := `git recent - 最近使用したブランチを表示して切り替え

使い方:
  git recent

説明:
  最近コミットがあったブランチを時系列順（最新順）に最大10件表示します。
  番号を入力することで、選択したブランチに即座に切り替えられます。
  現在のブランチは一覧から除外されます。

オプション:
  -h                    このヘルプを表示

使用方法:
  1. git recent を実行
  2. ブランチ一覧が表示される（最大10件、番号付き）
  3. 切り替えたいブランチの番号を入力
  4. 空入力でキャンセル

実装詳細:
  git for-each-ref --sort=-committerdate を使用して、
  最終コミット日時順にブランチを取得しています。
  現在のブランチは git branch --show-current で取得し、一覧から除外します。
`
	fmt.Print(help)
}
