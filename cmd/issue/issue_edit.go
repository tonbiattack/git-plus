// ================================================================================
// issue_edit.go
// ================================================================================
// このファイルは git の拡張コマンド issue-edit コマンドを実装しています。
// issue パッケージの一部として、GitHub issue 関連の機能を提供します。
//
// 【概要】
// issue-edit コマンドは、GitHubのopenしているissueの一覧を表示し、
// 選択したissueをエディタで編集する機能を提供します。
//
// 【主な機能】
// - GitHubのopenしているissueの一覧取得
// - 番号での選択
// - ユーザーが設定しているエディタ（VSCode等）での編集
// - 題名（title）と本文（body）の両方を編集可能
// - -v/--view オプションで閲覧モード（編集せずに表示のみ）
// - -m/--comment オプションでコメントを追加（クローズしない）
// - -c/--close オプションでコメント入力後にissueをクローズ
//
// 【使用例】
//   git issue-edit              # issueの一覧を表示して選択・編集
//   git issue-edit 123          # Issue #123 を直接編集
//   git issue-edit -v           # issueの一覧を表示して選択・閲覧のみ
//   git issue-edit 123 -v       # Issue #123 を直接閲覧
//   git issue-edit -m           # コメントを追加（クローズしない）
//   git issue-edit -c           # コメント入力後にissueをクローズ
//
// 【内部仕様】
// - GitHub CLI (gh) の gh issue list / gh issue edit を使用
// - git config core.editor または環境変数 EDITOR/VISUAL でエディタを取得
// - 一時ファイルに issue 本文を書き出してエディタで編集
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package issue

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/terminal"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// IssueEntry はissueの詳細情報を表す構造体
type IssueEntry struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// IssueContent はエディタで編集されたissueの内容を表す構造体
type IssueContent struct {
	Title string
	Body  string
}

// issueEditCmd は issue-edit コマンドの定義です。
var issueEditCmd = &cobra.Command{
	Use:   "issue-edit [ISSUE番号]",
	Short: "GitHubのissueを一覧表示して選択・編集",
	Long: `GitHubのopenしているissueの一覧を表示し、番号で選択して編集できます。
ISSUE番号を直接指定すると、そのissueを直接編集できます。
編集はユーザーが設定しているエディタ（VSCode等）で行えます。
題名（title）と本文（body）の両方を編集できます。

-v/--view オプションを使用すると、編集せずに閲覧のみ行えます。
-m/--comment オプションを使用すると、コメントを追加できます（クローズしない）。
-c/--close オプションを使用すると、コメント入力後にissueをクローズできます。

内部的に GitHub CLI (gh) を使用してissueと連携します。`,
	Example: `  git issue-edit              # issueの一覧を表示して選択・編集
  git issue-edit 123          # Issue #123 を直接編集
  git issue-edit -v           # issueの一覧を表示して選択・閲覧のみ
  git issue-edit 123 -v       # Issue #123 を直接閲覧
  git issue-edit -m           # コメントを追加（クローズしない）
  git issue-edit -c           # コメント入力後にissueをクローズ`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// フラグの取得
		viewOnly, err := cobraCmd.Flags().GetBool("view")
		if err != nil {
			return fmt.Errorf("viewフラグの取得に失敗: %w", err)
		}
		addComment, err := cobraCmd.Flags().GetBool("comment")
		if err != nil {
			return fmt.Errorf("commentフラグの取得に失敗: %w", err)
		}
		closeIssue, err := cobraCmd.Flags().GetBool("close")
		if err != nil {
			return fmt.Errorf("closeフラグの取得に失敗: %w", err)
		}

		// viewフラグと他のフラグの組み合わせチェック
		if viewOnly && (addComment || closeIssue) {
			return fmt.Errorf("viewフラグは他のフラグと同時に使用できません")
		}

		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		var selectedIssue *IssueEntry

		// 直接Issue番号が指定された場合
		if len(args) > 0 {
			issueNumberStr := ui.NormalizeNumberInput(args[0])
			issueNumber, err := strconv.Atoi(issueNumberStr)
			if err != nil {
				return fmt.Errorf("無効なIssue番号です: %s", args[0])
			}

			fmt.Printf("Issue #%d を取得中...\n", issueNumber)
			issue, err := getIssueByNumber(issueNumber)
			if err != nil {
				return fmt.Errorf("issue #%d の取得に失敗しました: %w", issueNumber, err)
			}

			selectedIssue = issue

			// 選択したissueの詳細を表示
			fmt.Printf("\nIssue: #%d\n", selectedIssue.Number)
			fmt.Printf("タイトル: %s\n", selectedIssue.Title)
			fmt.Printf("URL: %s\n\n", selectedIssue.URL)
		} else {
			// issue一覧を取得
			issues, err := getOpenIssueList()
			if err != nil {
				return fmt.Errorf("issueの一覧取得に失敗しました: %w", err)
			}

			if len(issues) == 0 {
				fmt.Println("openしているissueが存在しません。")
				return nil
			}

			// issue一覧を表示
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

			// issueを選択
			reader := bufio.NewReader(os.Stdin)
			for {
				if viewOnly {
					fmt.Print("閲覧するissueを選択してください (番号を入力、Enterでキャンセル): ")
				} else {
					fmt.Print("編集するissueを選択してください (番号を入力、Enterでキャンセル): ")
				}
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
				}

				input = ui.NormalizeNumberInput(input)
				if input == "" {
					fmt.Println("キャンセルしました。")
					return nil
				}

				selection, err := strconv.Atoi(input)
				if err != nil || selection < 1 || selection > len(issues) {
					fmt.Printf("無効な番号です。1から%dの範囲で入力してください。\n", len(issues))
					continue
				}

				selectedIssue = &issues[selection-1]
				break
			}

			// 選択したissueの詳細を表示
			fmt.Printf("\n選択されたissue: #%d\n", selectedIssue.Number)
			fmt.Printf("タイトル: %s\n", selectedIssue.Title)
			fmt.Printf("URL: %s\n\n", selectedIssue.URL)
		}

		// viewOnlyフラグの確認
		if viewOnly {
			// 閲覧モード: 本文を表示するだけ
			fmt.Println("--- 本文 ---")
			if selectedIssue.Body != "" {
				fmt.Println(selectedIssue.Body)
			} else {
				fmt.Println("(本文なし)")
			}
			fmt.Println()
			return nil
		}

		// コメント追加フラグまたはクローズフラグが指定された場合
		if addComment || closeIssue {
			// コメント入力画面を開く
			comment, err := promptForComment(selectedIssue)
			if err != nil {
				return fmt.Errorf("コメントの入力に失敗しました: %w", err)
			}

			// コメントが空でない場合は投稿
			if strings.TrimSpace(comment) != "" {
				if err := postComment(selectedIssue.Number, comment); err != nil {
					return fmt.Errorf("コメントの投稿に失敗: %w", err)
				}
				fmt.Println("✓ コメントを追加しました")
			}

			// issueをクローズ（-c が指定された場合）
			if closeIssue {
				if err := closeGitHubIssue(selectedIssue.Number); err != nil {
					return fmt.Errorf("issueのクローズに失敗しました: %w", err)
				}
				fmt.Println("✓ issueをクローズしました")
			}

			return nil
		}

		// エディタで題名と本文を編集
		if err := editIssue(selectedIssue); err != nil {
			return fmt.Errorf("issueの編集に失敗しました: %w", err)
		}

		fmt.Println("✓ issueを更新しました")
		return nil
	},
}

// checkGitHubCLIInstalled は GitHub CLI (gh) がインストールされているかを確認します。
func checkGitHubCLIInstalled() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// getOpenIssueList はopenしているissueの一覧を取得します。
func getOpenIssueList() ([]IssueEntry, error) {
	cmd := exec.Command("gh", "issue", "list", "--state", "open", "--json", "number,title,body,state,url", "--limit", "100")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue listの実行に失敗: %w", err)
	}

	var issues []IssueEntry
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	return issues, nil
}

// getIssueByNumber は指定されたIssue番号のissueを取得します。
func getIssueByNumber(issueNumber int) (*IssueEntry, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber), "--json", "number,title,body,state,url")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue viewの実行に失敗: %w", err)
	}

	var issue IssueEntry
	if err := json.Unmarshal(output, &issue); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %w", err)
	}

	return &issue, nil
}

// editIssue は指定されたissueの題名と本文をエディタで編集します。
// エディタがキャンセルされた場合は操作を中断します。
func editIssue(issue *IssueEntry) error {
	// エディタを取得
	editor, err := getEditor()
	if err != nil {
		return fmt.Errorf("エディタの取得に失敗: %w", err)
	}

	// 一時ファイルを作成
	tmpFile, err := createTempIssueFile(issue)
	if err != nil {
		return fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// エディタで編集
	fmt.Printf("エディタで編集中... (%s)\n", editor)
	fmt.Println("ヒント: エディタを保存せずに閉じると変更はキャンセルされます")
	if err := openEditor(editor, tmpFile); err != nil {
		// エディタがキャンセルされた場合
		if strings.Contains(err.Error(), "キャンセル") {
			fmt.Println("編集がキャンセルされました。")
			return nil
		}
		return fmt.Errorf("エディタの起動に失敗: %w", err)
	}

	// 編集後の内容を読み込み
	newContent, err := readFileContent(tmpFile)
	if err != nil {
		return fmt.Errorf("編集内容の読み込みに失敗: %w", err)
	}

	// 変更がない場合はスキップ
	titleChanged := strings.TrimSpace(newContent.Title) != strings.TrimSpace(issue.Title)
	bodyChanged := strings.TrimSpace(newContent.Body) != strings.TrimSpace(issue.Body)

	if !titleChanged && !bodyChanged {
		fmt.Println("変更がないため、更新をスキップしました。")
		return nil
	}

	// issueを更新
	if err := updateIssue(issue.Number, newContent.Title, newContent.Body); err != nil {
		return fmt.Errorf("issueの更新に失敗: %w", err)
	}

	return nil
}

// getEditor はユーザーが設定しているエディタを取得します。
// 優先順位: git config core.editor > VISUAL > EDITOR > デフォルト(vi)
func getEditor() (string, error) {
	// git config core.editor を確認
	output, err := gitcmd.Run("config", "--get", "core.editor")
	if err == nil && len(output) > 0 {
		editor := strings.TrimSpace(string(output))
		if editor != "" {
			return editor, nil
		}
	}

	// 環境変数 VISUAL を確認
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual, nil
	}

	// 環境変数 EDITOR を確認
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, nil
	}

	// デフォルトは vi
	return "vi", nil
}

// createTempIssueFile はissueの題名と本文を含む一時ファイルを作成します。
func createTempIssueFile(issue *IssueEntry) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("issue-%d.md", issue.Number))

	// ヘッダーコメント、題名、本文を書き込み
	content := fmt.Sprintf(`# Issue #%d
# URL: %s
#
# 以下のissueの題名と本文を編集してください。
# '#' で始まる行はコメントとして無視されます。
# 'Title:' の後に題名を記載し、'---' の区切り線の後に本文を記載してください。
# ファイルを保存して閉じると、issueが更新されます。
# ========================================

Title: %s

---

%s
`, issue.Number, issue.URL, issue.Title, issue.Body)

	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// openEditor は指定されたエディタでファイルを開きます。
// エディタが中断された場合や、ユーザーがキャンセルした場合はエラーを返します。
// ターミナルの状態は常に保護され、不正な状態になることはありません。
func openEditor(editor, filepath string) error {
	// VSCodeなどのエディタには --wait フラグを追加
	editor = terminal.AddWaitFlagIfNeeded(editor)

	// エディタコマンドをパースして引数を分割
	// 引用符を考慮したパースを行う
	parts, err := parseCommand(editor)
	if err != nil {
		return fmt.Errorf("エディタコマンドのパースに失敗: %w", err)
	}
	if len(parts) == 0 {
		return fmt.Errorf("エディタコマンドが空です")
	}

	args := append(parts[1:], filepath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// ターミナル状態を保護しながらエディタを実行
	result := terminal.RunEditorWithProtection(cmd)

	if result.Cancelled {
		return fmt.Errorf("操作がキャンセルされました")
	}
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// parseCommand はエディタコマンド文字列をパースして、引用符を考慮した引数リストに分割します。
func parseCommand(command string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case !inQuote && (r == '"' || r == '\''):
			// 引用符の開始
			inQuote = true
			quoteChar = r
		case inQuote && r == quoteChar:
			// 引用符の終了
			inQuote = false
			quoteChar = 0
		case !inQuote && r == ' ':
			// スペース: 引数の区切り
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		case r == '\\' && i+1 < len(runes):
			// バックスラッシュエスケープ
			next := runes[i+1]
			if inQuote && next == quoteChar {
				// 引用符内のエスケープされた引用符
				current.WriteRune(next)
				i++ // 次の文字をスキップ
			} else {
				// 通常のバックスラッシュ（Windowsのパス区切りなど）
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	// 最後の引数を追加
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	if inQuote {
		return nil, fmt.Errorf("引用符が閉じられていません")
	}

	return args, nil
}

// readFileContent はファイルの内容を読み込み、題名と本文を分離して返します。
func readFileContent(filepath string) (*IssueContent, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// コメント行（# で始まる行）を除去
	lines := strings.Split(string(content), "\n")
	var nonCommentLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			nonCommentLines = append(nonCommentLines, line)
		}
	}

	// 題名と本文を分離
	fullContent := strings.Join(nonCommentLines, "\n")

	// "Title:" を探す
	titlePrefix := "Title:"
	titleIndex := strings.Index(fullContent, titlePrefix)
	if titleIndex == -1 {
		return nil, fmt.Errorf("題名が見つかりませんでした。'Title:' の形式で記載してください")
	}

	// "---" の区切り線を探す
	separatorIndex := strings.Index(fullContent, "---")
	if separatorIndex == -1 {
		return nil, fmt.Errorf("区切り線 '---' が見つかりませんでした")
	}

	// 題名を抽出
	titleStart := titleIndex + len(titlePrefix)
	titleText := strings.TrimSpace(fullContent[titleStart:separatorIndex])

	// 本文を抽出
	bodyText := strings.TrimSpace(fullContent[separatorIndex+3:])

	return &IssueContent{
		Title: titleText,
		Body:  bodyText,
	}, nil
}

// updateIssue は指定されたissueの題名と本文を更新します。
func updateIssue(issueNumber int, newTitle, newBody string) error {
	cmd := exec.Command("gh", "issue", "edit", strconv.Itoa(issueNumber), "--title", newTitle, "--body", newBody)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// promptForComment はエディタでコメントを入力してもらい、その内容を返します。
// エディタがキャンセルされた場合は空文字列を返します。
func promptForComment(issue *IssueEntry) (string, error) {
	// エディタを取得
	editor, err := getEditor()
	if err != nil {
		return "", fmt.Errorf("エディタの取得に失敗: %w", err)
	}

	// 一時ファイルを作成
	tmpFile, err := createTempCommentFile(issue)
	if err != nil {
		return "", fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// エディタで編集
	fmt.Printf("エディタでコメントを入力中... (%s)\n", editor)
	fmt.Println("ヒント: エディタを保存せずに閉じるか、コメントを空のままにするとスキップされます")
	if err := openEditor(editor, tmpFile); err != nil {
		// エディタがキャンセルされた場合
		if strings.Contains(err.Error(), "キャンセル") {
			fmt.Println("コメント入力がキャンセルされました。")
			return "", nil
		}
		return "", fmt.Errorf("エディタの起動に失敗: %w", err)
	}

	// 編集後の内容を読み込み
	comment, err := readCommentFromFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("コメント内容の読み込みに失敗: %w", err)
	}

	return comment, nil
}

// createTempCommentFile はコメント入力用の一時ファイルを作成します。
func createTempCommentFile(issue *IssueEntry) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("issue-comment-%d.md", issue.Number))

	// ヘッダーコメントとコメント入力欄を書き込み
	content := fmt.Sprintf(`# Issue #%d へのコメント
# URL: %s
#
# このissueに追加するコメントを記載してください。
# '#' で始まる行はコメントとして無視されます。
# ファイルを保存して閉じると、コメントが投稿されます。
# ========================================

`, issue.Number, issue.URL)

	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// readCommentFromFile はファイルからコメント内容を読み込みます。
func readCommentFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// コメント行（# で始まる行）を除去
	lines := strings.Split(string(content), "\n")
	var nonCommentLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			nonCommentLines = append(nonCommentLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(nonCommentLines, "\n")), nil
}

// postComment は指定されたissueにコメントを投稿します。
func postComment(issueNumber int, comment string) error {
	cmd := exec.Command("gh", "issue", "comment", strconv.Itoa(issueNumber), "--body", comment)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// closeGitHubIssue は指定されたissueをクローズします。
func closeGitHubIssue(issueNumber int) error {
	cmd := exec.Command("gh", "issue", "close", strconv.Itoa(issueNumber))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// init は issue-edit コマンドを root コマンドに登録します。
func init() {
	cmd.RootCmd.AddCommand(issueEditCmd)
	issueEditCmd.Flags().BoolP("view", "v", false, "閲覧モード（編集せずに表示のみ）")
	issueEditCmd.Flags().BoolP("comment", "m", false, "コメントを追加")
	issueEditCmd.Flags().BoolP("close", "c", false, "issueをクローズ")
}
