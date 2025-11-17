// ================================================================================
// pr_issue_link.go
// ================================================================================
// このファイルは git の拡張コマンド pr-issue-link コマンドを実装しています。
// pr パッケージはPR関連のコマンドを提供します。
//
// 【概要】
// pr-issue-link コマンドは、PRを作成する際にGitHubのIssueと紐づける機能を
// 提供します。PRの説明欄に「Closes #<issue番号>」を自動的に追加することで、
// PRがマージされた際に関連するIssueが自動的にクローズされます。
//
// 【主な機能】
// - オープンなIssueの一覧表示と選択
// - 複数のIssueを選択可能
// - PRの説明欄に「Closes #番号」を自動追加
// - GitHub CLI (gh) を使用したPR作成
//
// 【使用例】
//   git pr-issue-link                    # 対話的にIssueを選択してPR作成
//   git pr-issue-link --base main        # mainブランチへのPR作成
//   git pr-issue-link --issue 123        # Issue #123 を指定してPR作成
//   git pr-issue-link --issue 123,456    # 複数のIssueを指定
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package pr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// IssueInfo はissueの詳細情報を表す構造体
type IssueInfo struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// prIssueLinkCmd は pr-issue-link コマンドの定義です。
// PRとIssueを紐づけて作成します。
var prIssueLinkCmd = &cobra.Command{
	Use:   "pr-issue-link [--base ベースブランチ] [--issue イシュー番号]",
	Short: "PRとIssueを紐づけて作成（Closes #番号を自動追加）",
	Long: `PRを作成する際にGitHubのIssueと紐づけます。
PRの説明欄に「Closes #<issue番号>」を自動的に追加することで、
PRがマージされた際に関連するIssueが自動的にクローズされます。

対話的にIssueを選択するか、--issue オプションで直接指定できます。
複数のIssueを紐づける場合は、カンマ区切りで指定します。`,
	Example: `  git pr-issue-link                    # 対話的にIssueを選択してPR作成
  git pr-issue-link --base main        # mainブランチへのPR作成
  git pr-issue-link --issue 123        # Issue #123 を指定してPR作成
  git pr-issue-link --issue 123,456    # 複数のIssueを指定
  git pr-issue-link -b develop -i 42   # developブランチへ、Issue #42 と紐づけ`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLIForPR() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// フラグの取得
		baseBranch, err := cobraCmd.Flags().GetString("base")
		if err != nil {
			return fmt.Errorf("baseフラグの取得に失敗: %w", err)
		}

		issueStr, err := cobraCmd.Flags().GetString("issue")
		if err != nil {
			return fmt.Errorf("issueフラグの取得に失敗: %w", err)
		}

		title, err := cobraCmd.Flags().GetString("title")
		if err != nil {
			return fmt.Errorf("titleフラグの取得に失敗: %w", err)
		}

		body, err := cobraCmd.Flags().GetString("body")
		if err != nil {
			return fmt.Errorf("bodyフラグの取得に失敗: %w", err)
		}

		// 現在のブランチを取得
		currentBranch, err := getCurrentBranchForPR()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗しました: %w", err)
		}

		if currentBranch == "" {
			return fmt.Errorf("ブランチが見つかりません。detached HEAD 状態かもしれません")
		}

		fmt.Printf("現在のブランチ: %s\n", currentBranch)

		// ベースブランチの取得
		if baseBranch == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("マージ先のベースブランチを入力してください (デフォルト: main): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
			}
			baseBranch = strings.TrimSpace(input)
			if baseBranch == "" {
				baseBranch = "main"
			}
		}

		fmt.Printf("ベースブランチ: %s\n\n", baseBranch)

		// Issueの選択
		var selectedIssues []int
		if issueStr != "" {
			// コマンドラインから指定された場合
			parts := strings.Split(issueStr, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				num, err := strconv.Atoi(p)
				if err != nil {
					return fmt.Errorf("無効なIssue番号: %s", p)
				}
				selectedIssues = append(selectedIssues, num)
			}
		} else {
			// 対話的に選択
			issues, err := getOpenIssueListForPR()
			if err != nil {
				return fmt.Errorf("issueの一覧取得に失敗しました: %w", err)
			}

			if len(issues) == 0 {
				fmt.Println("オープンなIssueが存在しません。")
				if !ui.Confirm("Issueと紐づけずにPRを作成しますか？", false) {
					fmt.Println("キャンセルしました。")
					return nil
				}
			} else {
				// Issue一覧を表示
				fmt.Printf("オープンなIssue一覧 (%d 個):\n\n", len(issues))
				for i, issue := range issues {
					fmt.Printf("%d. #%d: %s\n", i+1, issue.Number, issue.Title)
					bodyPreview := strings.TrimSpace(issue.Body)
					if len(bodyPreview) > 50 {
						bodyPreview = bodyPreview[:50] + "..."
					}
					if bodyPreview != "" {
						fmt.Printf("   %s\n", bodyPreview)
					}
					fmt.Println()
				}

				// Issueを選択
				selectedIssues, err = selectIssuesForPR(issues)
				if err != nil {
					return err
				}
			}
		}

		// タイトルの取得
		if title == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("PRのタイトルを入力してください (空の場合はコミットから自動生成): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
			}
			title = strings.TrimSpace(input)
		}

		// 本文の構築
		finalBody := buildPRBody(body, selectedIssues)

		// 確認
		fmt.Println("\n========================================")
		fmt.Printf("ベースブランチ: %s\n", baseBranch)
		fmt.Printf("ヘッドブランチ: %s\n", currentBranch)
		if title != "" {
			fmt.Printf("タイトル: %s\n", title)
		} else {
			fmt.Println("タイトル: (コミットから自動生成)")
		}
		if len(selectedIssues) > 0 {
			fmt.Printf("紐づけるIssue: ")
			for i, num := range selectedIssues {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("#%d", num)
			}
			fmt.Println()
		}
		if finalBody != "" {
			fmt.Println("\n--- PR本文 ---")
			fmt.Println(finalBody)
		}
		fmt.Println("========================================")

		if !ui.Confirm("\nPRを作成しますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// PRを作成
		fmt.Println("\nPRを作成しています...")
		prURL, err := createPRWithIssueLink(baseBranch, currentBranch, title, finalBody)
		if err != nil {
			return fmt.Errorf("PRの作成に失敗しました: %w", err)
		}

		fmt.Printf("\n✓ PRを作成しました\n")
		fmt.Printf("URL: %s\n", prURL)

		if len(selectedIssues) > 0 {
			fmt.Println("\n紐づけられたIssue:")
			for _, num := range selectedIssues {
				fmt.Printf("  - Issue #%d (PRマージ時に自動クローズされます)\n", num)
			}
		}

		return nil
	},
}

// checkGitHubCLIForPR は GitHub CLI (gh) がインストールされているかを確認します。
func checkGitHubCLIForPR() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// getOpenIssueListForPR はopenしているissueの一覧を取得します。
func getOpenIssueListForPR() ([]IssueInfo, error) {
	ghCmd := exec.Command("gh", "issue", "list", "--state", "open", "--json", "number,title,body,state,url", "--limit", "100")
	output, err := ghCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue listの実行に失敗: %w", err)
	}

	var issues []IssueInfo
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	return issues, nil
}

// selectIssuesForPR はユーザーにIssueを選択させます。
// 複数選択が可能です。
func selectIssuesForPR(issues []IssueInfo) ([]int, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("紐づけるIssueを選択してください:")
	fmt.Println("  番号: 単一のIssueを選択")
	fmt.Println("  1,3,5: 複数のIssueを選択（カンマ区切り）")
	fmt.Println("  all: すべてのIssueを選択")
	fmt.Println("  none: Issueを紐づけない")
	fmt.Print("\n入力: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("入力の読み込みに失敗しました: %w", err)
	}

	input = strings.TrimSpace(input)

	// none の場合
	if input == "none" || input == "n" || input == "" {
		return nil, nil
	}

	// all の場合
	if input == "all" || input == "a" {
		var selected []int
		for _, issue := range issues {
			selected = append(selected, issue.Number)
		}
		return selected, nil
	}

	// 番号選択
	var selected []int
	parts := strings.Split(input, ",")
	for _, p := range parts {
		p = ui.NormalizeNumberInput(strings.TrimSpace(p))
		if p == "" {
			continue
		}

		idx, err := strconv.Atoi(p)
		if err != nil || idx < 1 || idx > len(issues) {
			return nil, fmt.Errorf("無効な番号: %s (1から%dの範囲で入力してください)", p, len(issues))
		}

		issueNumber := issues[idx-1].Number
		// 重複チェック
		found := false
		for _, num := range selected {
			if num == issueNumber {
				found = true
				break
			}
		}
		if !found {
			selected = append(selected, issueNumber)
		}
	}

	return selected, nil
}

// buildPRBody はPRの本文を構築します。
// 選択されたIssue番号に対して「Closes #番号」を追加します。
func buildPRBody(userBody string, issueNumbers []int) string {
	var parts []string

	// ユーザーが入力した本文
	if userBody != "" {
		parts = append(parts, userBody)
	}

	// Issue紐づけ
	if len(issueNumbers) > 0 {
		var closesLines []string
		for _, num := range issueNumbers {
			closesLines = append(closesLines, fmt.Sprintf("Closes #%d", num))
		}
		closesText := strings.Join(closesLines, "\n")
		parts = append(parts, closesText)
	}

	return strings.Join(parts, "\n\n")
}

// createPRWithIssueLink はPRを作成し、URLを返します。
func createPRWithIssueLink(baseBranch, headBranch, title, body string) (string, error) {
	args := []string{"pr", "create", "--base", baseBranch, "--head", headBranch}

	if title != "" {
		args = append(args, "--title", title)
	} else {
		// タイトルが指定されていない場合は --fill を使用
		args = append(args, "--fill")
	}

	if body != "" {
		args = append(args, "--body", body)
	}

	ghCmd := exec.Command("gh", args...)
	output, err := ghCmd.Output()
	if err != nil {
		// 標準エラー出力も含めてエラーを返す
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%w: %s", err, string(exitErr.Stderr))
		}
		return "", err
	}

	prURL := strings.TrimSpace(string(output))
	return prURL, nil
}

// init は pr-issue-link コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	cmd.RootCmd.AddCommand(prIssueLinkCmd)
	prIssueLinkCmd.Flags().StringP("base", "b", "", "マージ先のベースブランチ")
	prIssueLinkCmd.Flags().StringP("issue", "i", "", "紐づけるIssue番号（カンマ区切りで複数指定可能）")
	prIssueLinkCmd.Flags().StringP("title", "t", "", "PRのタイトル")
	prIssueLinkCmd.Flags().String("body", "", "PRの本文（Closes #番号 は自動追加されます）")
}
