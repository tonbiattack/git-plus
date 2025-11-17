// ================================================================================
// issue_bulk_close.go
// ================================================================================
// このファイルは git の拡張コマンド issue-bulk-close コマンドを実装しています。
// issue パッケージの一部として、GitHub issue 関連の機能を提供します。
//
// 【概要】
// issue-bulk-close コマンドは、複数のGitHub issueを同じコメントで一括クローズする
// 機能を提供します。
//
// 【主な機能】
// - 複数のissue番号を引数で指定
// - 同じコメントを全issueに投稿
// - コメント入力後に一括でクローズ
// - -m/--message オプションでコメントを直接指定可能
//
// 【使用例】
//   git issue-bulk-close 1 2 3           # Issue #1, #2, #3 を一括クローズ（エディタでコメント入力）
//   git issue-bulk-close 1 2 -m "完了"   # Issue #1, #2 を "完了" コメントでクローズ
//
// 【内部仕様】
// - GitHub CLI (gh) の gh issue comment / gh issue close を使用
// - git config core.editor または環境変数 EDITOR/VISUAL でエディタを取得
// - 一時ファイルにコメントを書き出してエディタで編集
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package issue

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// issueBulkCloseCmd は issue-bulk-close コマンドの定義です。
var issueBulkCloseCmd = &cobra.Command{
	Use:   "issue-bulk-close <ISSUE番号...>",
	Short: "複数のGitHub issueを同じコメントで一括クローズ",
	Long: `複数のGitHub issueを同じコメントで一括クローズします。
Issue番号を複数指定して、同一のコメントを全てのissueに追加した後、クローズします。

-m/--message オプションでコメントを直接指定できます。
指定しない場合は、エディタが開いてコメントを入力できます。

内部的に GitHub CLI (gh) を使用してissueと連携します。`,
	Example: `  git issue-bulk-close 1 2 3           # Issue #1, #2, #3 を一括クローズ（エディタでコメント入力）
  git issue-bulk-close 1 2 -m "完了"   # Issue #1, #2 を "完了" コメントでクローズ
  git issue-bulk-close １ ２ ３         # 全角数字も対応`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// メッセージフラグの取得
		message, err := cobraCmd.Flags().GetString("message")
		if err != nil {
			return fmt.Errorf("messageフラグの取得に失敗: %w", err)
		}

		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// issue番号のパース
		var issueNumbers []int
		for _, arg := range args {
			normalizedArg := ui.NormalizeNumberInput(arg)
			num, err := strconv.Atoi(normalizedArg)
			if err != nil {
				return fmt.Errorf("無効なIssue番号です: %s", arg)
			}
			issueNumbers = append(issueNumbers, num)
		}

		// 重複チェック
		seen := make(map[int]bool)
		var uniqueNumbers []int
		for _, num := range issueNumbers {
			if !seen[num] {
				seen[num] = true
				uniqueNumbers = append(uniqueNumbers, num)
			}
		}
		issueNumbers = uniqueNumbers

		// 対象issueの確認と存在チェック
		fmt.Printf("以下の %d 個のissueをクローズします:\n", len(issueNumbers))
		var validIssues []*IssueEntry
		for _, num := range issueNumbers {
			issue, err := getIssueByNumber(num)
			if err != nil {
				return fmt.Errorf("issue #%d の取得に失敗しました: %w", num, err)
			}
			if issue.State != "OPEN" {
				fmt.Printf("⚠ Issue #%d は既にクローズされています。スキップします。\n", num)
				continue
			}
			fmt.Printf("  - #%d: %s\n", issue.Number, issue.Title)
			validIssues = append(validIssues, issue)
		}

		if len(validIssues) == 0 {
			fmt.Println("クローズ対象のopenなissueがありません。")
			return nil
		}

		fmt.Println()

		// 確認
		if !ui.Confirm("これらのissueをクローズしますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// コメントの取得
		var comment string
		if message != "" {
			comment = message
		} else {
			// エディタでコメントを入力
			var err error
			comment, err = promptForBulkCloseComment(issueNumbers)
			if err != nil {
				return fmt.Errorf("コメントの入力に失敗しました: %w", err)
			}
		}

		// 各issueにコメント追加とクローズを実行
		fmt.Println()
		successCount := 0
		failCount := 0
		for _, issue := range validIssues {
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

		return nil
	},
}

// promptForBulkCloseComment はエディタで一括クローズ用のコメントを入力してもらい、その内容を返します。
func promptForBulkCloseComment(issueNumbers []int) (string, error) {
	// エディタを取得
	editor, err := getEditor()
	if err != nil {
		return "", fmt.Errorf("エディタの取得に失敗: %w", err)
	}

	// 一時ファイルを作成
	tmpFile, err := createTempBulkCloseCommentFile(issueNumbers)
	if err != nil {
		return "", fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// エディタで編集
	fmt.Printf("エディタでコメントを入力中... (%s)\n", editor)
	if err := openEditor(editor, tmpFile); err != nil {
		return "", fmt.Errorf("エディタの起動に失敗: %w", err)
	}

	// 編集後の内容を読み込み
	comment, err := readCommentFromFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("コメント内容の読み込みに失敗: %w", err)
	}

	return comment, nil
}

// createTempBulkCloseCommentFile は一括クローズ用のコメント入力用一時ファイルを作成します。
func createTempBulkCloseCommentFile(issueNumbers []int) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "issue-bulk-close-comment.md")

	// issue番号のリストを作成
	var issueList strings.Builder
	for _, num := range issueNumbers {
		issueList.WriteString(fmt.Sprintf("# - Issue #%d\n", num))
	}

	// ヘッダーコメントとコメント入力欄を書き込み
	content := fmt.Sprintf(`# 一括クローズ用コメント
#
# 以下のissueに同じコメントを投稿してクローズします:
%s#
# このコメントは全てのissueに投稿されます。
# '#' で始まる行はコメントとして無視されます。
# ファイルを保存して閉じると、コメントが投稿されissueがクローズされます。
# コメントを空にすると、コメントなしでクローズされます。
# ========================================

`, issueList.String())

	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// init は issue-bulk-close コマンドを root コマンドに登録します。
func init() {
	cmd.RootCmd.AddCommand(issueBulkCloseCmd)
	issueBulkCloseCmd.Flags().StringP("message", "m", "", "クローズ時のコメント（指定しない場合はエディタで入力）")
}
