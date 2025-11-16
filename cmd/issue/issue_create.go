// ================================================================================
// issue_create.go
// ================================================================================
// このファイルは git の拡張コマンド issue-create コマンドを実装しています。
// issue パッケージの一部として、GitHub issue 関連の機能を提供します。
//
// 【概要】
// issue-create コマンドは、GitHubに新しいissueを作成する機能を提供します。
//
// 【主な機能】
// - ユーザーが設定しているエディタ（VSCode等）での題名と本文の入力
// - GitHubへの新しいissue作成
//
// 【使用例】
//   git issue-create          # エディタでissueの題名と本文を入力して作成
//
// 【内部仕様】
// - GitHub CLI (gh) の gh issue create を使用
// - git config core.editor または環境変数 EDITOR/VISUAL でエディタを取得
// - 一時ファイルに issue の題名と本文を書き出してエディタで編集
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package issue

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
)

// issueCreateCmd は issue-create コマンドの定義です。
var issueCreateCmd = &cobra.Command{
	Use:   "issue-create",
	Short: "GitHubに新しいissueを作成",
	Long: `エディタで題名と本文を入力し、GitHubに新しいissueを作成します。
ユーザーが設定しているエディタ（VSCode等）で題名と本文を編集できます。

内部的に GitHub CLI (gh) を使用してissueを作成します。`,
	Example: `  git issue-create          # エディタでissueを作成`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// エディタで題名と本文を作成
		content, err := createIssueInEditor()
		if err != nil {
			return fmt.Errorf("issueの作成に失敗しました: %w", err)
		}

		// 題名と本文が空でないことを確認
		if strings.TrimSpace(content.Title) == "" {
			return fmt.Errorf("題名が空です。issueの作成をキャンセルしました")
		}

		// issueを作成
		issueURL, err := createIssue(content.Title, content.Body)
		if err != nil {
			return fmt.Errorf("issueの作成に失敗しました: %w", err)
		}

		fmt.Printf("✓ issueを作成しました\n")
		fmt.Printf("URL: %s\n", issueURL)
		return nil
	},
}

// createIssueInEditor はエディタで新しいissueの題名と本文を入力します。
func createIssueInEditor() (*IssueContent, error) {
	// エディタを取得
	editor, err := getEditor()
	if err != nil {
		return nil, fmt.Errorf("エディタの取得に失敗: %w", err)
	}

	// 一時ファイルを作成
	tmpFile, err := createTempNewIssueFile()
	if err != nil {
		return nil, fmt.Errorf("一時ファイルの作成に失敗: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// エディタで編集
	fmt.Printf("エディタで新しいissueを作成中... (%s)\n", editor)
	if err := openEditor(editor, tmpFile); err != nil {
		return nil, fmt.Errorf("エディタの起動に失敗: %w", err)
	}

	// 編集後の内容を読み込み
	content, err := readFileContent(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("編集内容の読み込みに失敗: %w", err)
	}

	return content, nil
}

// createTempNewIssueFile は新しいissueの題名と本文を入力するための一時ファイルを作成します。
func createTempNewIssueFile() (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "new-issue.md")

	// ヘッダーコメントとテンプレートを書き込み
	content := `# 新しいIssueを作成
#
# 以下のissueの題名と本文を入力してください。
# '#' で始まる行はコメントとして無視されます。
# 'Title:' の後に題名を記載し、'---' の区切り線の後に本文を記載してください。
# ファイルを保存して閉じると、issueが作成されます。
# ========================================

Title:

---

`

	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// createIssue は指定された題名と本文で新しいissueを作成します。
func createIssue(title, body string) (string, error) {
	cmd := exec.Command("gh", "issue", "create", "--title", title, "--body", body)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// gh issue create の出力からURLを取得
	issueURL := strings.TrimSpace(string(output))
	return issueURL, nil
}

// init は issue-create コマンドを root コマンドに登録します。
func init() {
	cmd.RootCmd.AddCommand(issueCreateCmd)
}
