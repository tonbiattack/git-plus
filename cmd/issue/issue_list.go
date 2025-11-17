// ================================================================================
// issue_list.go
// ================================================================================
// このファイルは git の拡張コマンド issue-list コマンドを実装しています。
// issue パッケージの一部として、GitHub issue 関連の機能を提供します。
//
// 【概要】
// issue-list コマンドは、GitHubのopenしているissueの一覧を表示し、
// 選択したissueの詳細を確認し、そこから編集・コメント追加・クローズ・
// 新規作成など様々なアクションを実行できる統合的なインターフェースを提供します。
//
// 【主な機能】
// - GitHubのopenしているissueの一覧取得
// - 番号での選択と詳細表示
// - アクションメニュー（編集/コメント追加/クローズ/新規作成/戻る）
// - 各種モード間のシームレスな切り替え
//
// 【使用例】
//
//	git issue-list              # issueの一覧を表示して選択・操作
//
// 【内部仕様】
// - GitHub CLI (gh) の gh issue list / gh issue edit / gh issue create を使用
// - git config core.editor または環境変数 EDITOR/VISUAL でエディタを取得
// - 一時ファイルに issue 本文を書き出してエディタで編集
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================
package issue

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// issueListCmd は issue-list コマンドの定義です。
var issueListCmd = &cobra.Command{
	Use:   "issue-list",
	Short: "GitHubのissueを一覧表示して選択・操作",
	Long: `GitHubのopenしているissueの一覧を表示し、番号で選択して詳細を確認できます。
詳細表示後、様々なアクションを選択できます：
- 編集: issueの題名と本文を編集
- コメント追加: issueにコメントを追加
- クローズ: issueをクローズ（コメント追加後）
- 新規作成: 新しいissueを作成
- 一覧に戻る: issue一覧に戻る
- 一括クローズ: 複数のissueを同じコメントでクローズ

内部的に GitHub CLI (gh) を使用してissueと連携します。`,
	Example: `  git issue-list              # issueの一覧を表示して選択・操作`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		return runIssueListLoop()
	},
}

// runIssueListLoop はissue一覧表示とアクション選択のメインループを実行します。
func runIssueListLoop() error {
	for {
		// issue一覧を取得
		issues, err := getOpenIssueList()
		if err != nil {
			return fmt.Errorf("issueの一覧取得に失敗しました: %w", err)
		}

		if len(issues) == 0 {
			fmt.Println("openしているissueが存在しません。")
			fmt.Println()

			// issueがない場合でも新規作成のオプションを提供
			shouldCreate, err := promptForNewIssueWhenEmpty()
			if err != nil {
				return err
			}
			if shouldCreate {
				if err := createNewIssue(); err != nil {
					return err
				}
				continue // 作成後、一覧に戻る
			}
			return nil
		}

		// issue一覧を表示
		displayIssueList(issues)

		// issueを選択または新規作成
		selectedIssue, shouldExit, err := selectIssueOrCreate(issues)
		if err != nil {
			// 一括クローズ完了後は一覧を再表示
			if err.Error() == "BULK_CLOSE_DONE" {
				continue
			}
			return err
		}
		if shouldExit {
			return nil
		}
		if selectedIssue == nil {
			// 新規作成が選択された
			if err := createNewIssue(); err != nil {
				return err
			}
			continue // 作成後、一覧に戻る
		}

		// 選択したissueの詳細を表示
		displayIssueDetail(selectedIssue)

		// アクションメニューを表示
		shouldContinue, err := showActionMenu(selectedIssue)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}
		// shouldContinueがtrueなら一覧に戻る
	}
}

// displayIssueList はissue一覧を表示します。
func displayIssueList(issues []IssueEntry) {
	fmt.Printf("Open Issue一覧 (%d 個):\n\n", len(issues))
	for i, issue := range issues {
		fmt.Printf("%d. #%d: %s\n", i+1, issue.Number, issue.Title)
		// bodyの最初の50文字を表示（プレビュー）
		bodyPreview := strings.TrimSpace(issue.Body)
		if len(bodyPreview) > 50 {
			bodyPreview = bodyPreview[:50] + "..."
		}
		if bodyPreview != "" {
			fmt.Printf("   %s\n", bodyPreview)
		}
		fmt.Println()
	}
}

// selectIssueOrCreate はユーザーにissueを選択させるか、新規作成を選ばせます。
// 戻り値: (選択されたissue, 終了フラグ, エラー)
// - issueがnilで終了フラグがfalseの場合は新規作成が選択された
// - 終了フラグがtrueの場合は終了が選択された
func selectIssueOrCreate(issues []IssueEntry) (*IssueEntry, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("選択してください:\n")
		fmt.Printf("  番号: issueの詳細を表示\n")
		fmt.Printf("  n: 新規issueを作成\n")
		fmt.Printf("  bc: 複数issueを一括クローズ\n")
		fmt.Printf("  q: 終了\n")
		fmt.Print("\n入力: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, false, fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		input = strings.TrimSpace(input)

		// 終了
		if input == "q" || input == "Q" || input == "ｑ" || input == "Ｑ" {
			fmt.Println("終了します。")
			return nil, true, nil
		}

		// 新規作成
		if input == "n" || input == "N" || input == "ｎ" || input == "Ｎ" {
			return nil, false, nil
		}

		// 一括クローズ
		if input == "bc" || input == "BC" || input == "ｂｃ" || input == "ＢＣ" {
			if err := performBulkClose(issues); err != nil {
				return nil, false, err
			}
			// 一括クローズ後は一覧を再表示するため、特別な値を返す
			// nilを返して新規作成モードに入らないように、特別なフラグを使う
			return nil, false, fmt.Errorf("BULK_CLOSE_DONE")
		}

		// 番号選択
		input = ui.NormalizeNumberInput(input)
		if input == "" {
			fmt.Println("無効な入力です。")
			fmt.Println()
			continue
		}

		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(issues) {
			fmt.Printf("無効な番号です。1から%dの範囲で入力してください。\n", len(issues))
			fmt.Println()
			continue
		}

		return &issues[selection-1], false, nil
	}
}

// displayIssueDetail は選択されたissueの詳細を表示します。
func displayIssueDetail(issue *IssueEntry) {
	fmt.Printf("\n========================================\n")
	fmt.Printf("Issue: #%d\n", issue.Number)
	fmt.Printf("タイトル: %s\n", issue.Title)
	fmt.Printf("URL: %s\n", issue.URL)
	fmt.Printf("========================================\n")
	fmt.Println("--- 本文 ---")
	if issue.Body != "" {
		fmt.Println(issue.Body)
	} else {
		fmt.Println("(本文なし)")
	}
	fmt.Printf("========================================\n\n")
}

// showActionMenu はアクションメニューを表示し、選択されたアクションを実行します。
// 戻り値: (一覧に戻るかどうか, エラー)
func showActionMenu(issue *IssueEntry) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("アクションを選択してください:")
		fmt.Println("  e: 編集（題名・本文）")
		fmt.Println("  m: コメント追加")
		fmt.Println("  c: クローズ（コメント追加後）")
		fmt.Println("  n: 新規issue作成")
		fmt.Println("  b: 一覧に戻る")
		fmt.Println("  q: 終了")
		fmt.Print("\n入力: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "e", "ｅ":
			// 編集モード
			if err := editIssue(issue); err != nil {
				return false, fmt.Errorf("issueの編集に失敗しました: %w", err)
			}
			fmt.Println("✓ issueを更新しました")
			return true, nil // 一覧に戻る

		case "m", "ｍ":
			// コメント追加モード
			comment, err := promptForComment(issue)
			if err != nil {
				return false, fmt.Errorf("コメントの入力に失敗しました: %w", err)
			}
			if strings.TrimSpace(comment) != "" {
				if err := postComment(issue.Number, comment); err != nil {
					return false, fmt.Errorf("コメントの投稿に失敗: %w", err)
				}
				fmt.Println("✓ コメントを追加しました")
			} else {
				fmt.Println("コメントが空のため、投稿をスキップしました。")
			}
			return true, nil // 一覧に戻る

		case "c", "ｃ":
			// クローズモード
			comment, err := promptForComment(issue)
			if err != nil {
				return false, fmt.Errorf("コメントの入力に失敗しました: %w", err)
			}
			if strings.TrimSpace(comment) != "" {
				if err := postComment(issue.Number, comment); err != nil {
					return false, fmt.Errorf("コメントの投稿に失敗: %w", err)
				}
				fmt.Println("✓ コメントを追加しました")
			}
			if err := closeGitHubIssue(issue.Number); err != nil {
				return false, fmt.Errorf("issueのクローズに失敗しました: %w", err)
			}
			fmt.Println("✓ issueをクローズしました")
			return true, nil // 一覧に戻る

		case "n", "ｎ":
			// 新規作成モード
			if err := createNewIssue(); err != nil {
				return false, err
			}
			return true, nil // 一覧に戻る

		case "b", "ｂ":
			// 一覧に戻る
			fmt.Println()
			return true, nil

		case "q", "ｑ":
			// 終了
			fmt.Println("終了します。")
			return false, nil

		default:
			fmt.Println("無効な入力です。e, m, c, n, b, q のいずれかを入力してください。")
			fmt.Println()
		}
	}
}

// promptForNewIssueWhenEmpty はissueが存在しない場合に新規作成するかを確認します。
func promptForNewIssueWhenEmpty() (bool, error) {
	return ui.Confirm("新しいissueを作成しますか？", false), nil
}

// createNewIssue は新しいissueを作成します。
func createNewIssue() error {
	// エディタで題名と本文を作成
	content, err := createIssueInEditor()
	if err != nil {
		return fmt.Errorf("issueの作成に失敗しました: %w", err)
	}

	// 題名と本文が空でないことを確認
	if strings.TrimSpace(content.Title) == "" {
		fmt.Println("題名が空です。issueの作成をキャンセルしました。")
		return nil
	}

	// issueを作成
	issueURL, err := createIssue(content.Title, content.Body)
	if err != nil {
		return fmt.Errorf("issueの作成に失敗しました: %w", err)
	}

	// イシュー番号を抽出
	issueNumber := extractIssueNumber(issueURL)
	if issueNumber != "" {
		fmt.Printf("✓ Issue #%s: %s を作成しました\n", issueNumber, content.Title)
	} else {
		fmt.Printf("✓ issueを作成しました\n")
	}
	fmt.Printf("URL: %s\n", issueURL)
	if content.Body != "" {
		fmt.Printf("\n--- 本文 ---\n%s\n", content.Body)
	}
	fmt.Println()
	return nil
}

// performBulkClose は複数のissueを一括でクローズします。
func performBulkClose(issues []IssueEntry) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== 一括クローズ ===")
	fmt.Println("クローズするissueの番号をスペース区切りで入力してください。")
	fmt.Println("例: 1 3 5")
	fmt.Print("\n入力: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("キャンセルしました。")
		return nil
	}

	// 番号をパース
	parts := strings.Fields(input)
	var selectedIssues []*IssueEntry
	seen := make(map[int]bool)

	for _, part := range parts {
		normalized := ui.NormalizeNumberInput(part)
		num, err := strconv.Atoi(normalized)
		if err != nil || num < 1 || num > len(issues) {
			fmt.Printf("無効な番号をスキップ: %s\n", part)
			continue
		}
		if !seen[num] {
			seen[num] = true
			selectedIssues = append(selectedIssues, &issues[num-1])
		}
	}

	if len(selectedIssues) == 0 {
		fmt.Println("有効なissueが選択されませんでした。")
		return nil
	}

	// 選択されたissueを表示
	fmt.Printf("\n以下の %d 個のissueをクローズします:\n", len(selectedIssues))
	for _, issue := range selectedIssues {
		fmt.Printf("  - #%d: %s\n", issue.Number, issue.Title)
	}
	fmt.Println()

	// 確認
	if !ui.Confirm("これらのissueをクローズしますか？", true) {
		fmt.Println("キャンセルしました。")
		return nil
	}

	// コメントの入力
	var issueNumbers []int
	for _, issue := range selectedIssues {
		issueNumbers = append(issueNumbers, issue.Number)
	}

	comment, err := promptForBulkCloseComment(issueNumbers)
	if err != nil {
		return fmt.Errorf("コメントの入力に失敗しました: %w", err)
	}

	// 各issueにコメント追加とクローズを実行
	fmt.Println()
	successCount := 0
	failCount := 0
	for _, issue := range selectedIssues {
		fmt.Printf("Issue #%d を処理中...\n", issue.Number)

		// コメントが空でない場合は投稿
		if strings.TrimSpace(comment) != "" {
			if err := postComment(issue.Number, comment); err != nil {
				fmt.Printf("✗ Issue #%d へのコメント追加に失敗: %v\n", issue.Number, err)
				failCount++
				continue
			}
			fmt.Printf("  ✓ コメントを追加しました\n")
		}

		// issueをクローズ
		if err := closeGitHubIssue(issue.Number); err != nil {
			fmt.Printf("✗ Issue #%d のクローズに失敗: %v\n", issue.Number, err)
			failCount++
			continue
		}
		fmt.Printf("  ✓ クローズしました\n")
		successCount++
	}

	// 結果サマリー
	fmt.Printf("\n完了: %d 個のissueをクローズしました", successCount)
	if failCount > 0 {
		fmt.Printf(" (%d 個失敗)", failCount)
	}
	fmt.Println()
	fmt.Println()

	return nil
}

// init は issue-list コマンドを root コマンドに登録します。
func init() {
	cmd.RootCmd.AddCommand(issueListCmd)
}
