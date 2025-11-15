// ================================================================================
// release_notes.go
// ================================================================================
// このファイルは git-plus の release-notes コマンドを実装しています。
//
// 【概要】
// release-notes コマンドは、既存のタグからGitHubのリリースノートを自動生成します。
// GitHub CLIのgh release createコマンドを使用して、タグ間の変更内容を
// 自動的に解析し、リリースノートを作成します。
//
// 【主な機能】
// - 既存のタグ一覧の表示と選択
// - GitHub CLIによるリリースノートの自動生成
// - ドラフトまたはプレリリースとしての作成
// - 最新タグまたは指定タグからの作成
//
// 【使用例】
//   git-plus release-notes                  # 対話的にタグを選択
//   git-plus release-notes --tag v1.2.3     # 指定したタグからリリース作成
//   git-plus release-notes --latest         # 最新タグからリリース作成
//   git-plus release-notes --draft          # ドラフトとして作成
//   git-plus release-notes --prerelease     # プレリリースとして作成
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	releaseTag        string // リリースを作成するタグ
	releaseDraft      bool   // ドラフトとして作成
	releasePrerelease bool   // プレリリースとして作成
	releaseLatest     bool   // 最新タグを使用
)

// releaseNotesCmd は release-notes コマンドの定義です。
var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "既存のタグからGitHubリリースノートを自動生成",
	Long: `既存のタグからGitHubのリリースノートを自動生成します。
GitHub CLIのgh release createコマンドを使用して、タグ間の変更内容を
自動的に解析し、リリースノートを作成します。

注意: このコマンドは既存のタグに対してリリースを作成します。
      新しいタグを作成する場合は、事前に git new-tag コマンドを使用してください。`,
	Example: `  git-plus release-notes                  # 対話的にタグを選択
  git-plus release-notes --tag v1.2.3     # 指定したタグからリリース作成
  git-plus release-notes --latest         # 最新タグからリリース作成
  git-plus release-notes --draft          # ドラフトとして作成
  git-plus release-notes --prerelease     # プレリリースとして作成
  git-plus release-notes --tag v1.2.3 --draft --prerelease`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		var selectedTag string

		// タグの選択
		if releaseTag != "" {
			// --tag オプションで指定されたタグを使用
			selectedTag = releaseTag
			// タグの存在確認
			if err := verifyTagExists(selectedTag); err != nil {
				return fmt.Errorf("タグ '%s' が存在しません", selectedTag)
			}
		} else if releaseLatest {
			// --latest オプションで最新タグを使用
			latestTag, err := getLatestTag()
			if err != nil {
				return fmt.Errorf("最新タグの取得に失敗しました: %w\nタグが存在しない可能性があります", err)
			}
			selectedTag = latestTag
			fmt.Printf("最新タグ: %s\n", selectedTag)
		} else {
			// 対話的にタグを選択
			tag, err := selectTagInteractively()
			if err != nil {
				return err
			}
			selectedTag = tag
		}

		fmt.Printf("\nタグ: %s\n", selectedTag)
		if releaseDraft {
			fmt.Println("モード: ドラフト")
		}
		if releasePrerelease {
			fmt.Println("モード: プレリリース")
		}

		// 確認プロンプト
		if !ui.Confirm("\nリリースノートを作成しますか？", true) {
			fmt.Println("キャンセルしました")
			return nil
		}

		// リリースノートを作成
		if err := createReleaseNotes(selectedTag, releaseDraft, releasePrerelease); err != nil {
			return fmt.Errorf("リリースノートの作成に失敗しました: %w", err)
		}

		fmt.Printf("\n✓ リリースノートを作成しました\n")
		fmt.Printf("詳細を確認するには: gh release view %s --web\n", selectedTag)

		return nil
	},
}

// selectTagInteractively は対話的にタグを選択します。
//
// 戻り値:
//   - string: 選択されたタグ名
//   - error: エラーが発生した場合のエラー情報
func selectTagInteractively() (string, error) {
	// タグ一覧を取得（最新10個）
	tags, err := getRecentTags(10)
	if err != nil {
		return "", fmt.Errorf("タグの取得に失敗しました: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("タグが存在しません\n最初のタグを作成するには: git new-tag コマンドを使用してください")
	}

	// タグ一覧を表示
	fmt.Printf("\n最近のタグ一覧 (%d 個):\n\n", len(tags))
	for i, tag := range tags {
		fmt.Printf("%d. %s\n", i+1, tag)
	}
	fmt.Println()

	// ユーザー入力を取得
	fmt.Print("リリースノートを作成するタグを選択してください (番号を入力、Enterでキャンセル): ")
	var input string
	fmt.Scanln(&input)

	input = ui.NormalizeNumberInput(input)
	if input == "" {
		return "", fmt.Errorf("キャンセルされました")
	}

	// 番号を解析
	index, err := strconv.Atoi(input)
	if err != nil || index < 1 || index > len(tags) {
		return "", fmt.Errorf("無効な選択です: %s", input)
	}

	selectedTag := tags[index-1]
	fmt.Printf("\n選択されたタグ: %s\n", selectedTag)

	return selectedTag, nil
}

// getRecentTags は最新のタグを取得します。
//
// パラメータ:
//   - limit: 取得するタグの最大数
//
// 戻り値:
//   - []string: タグのスライス（新しい順）
//   - error: エラーが発生した場合のエラー情報
func getRecentTags(limit int) ([]string, error) {
	// セマンティックバージョン順でソートして全タグを取得
	output, err := gitcmd.Run("tag", "--sort=-v:refname")
	if err != nil {
		return nil, err
	}

	tagsStr := strings.TrimSpace(string(output))
	if tagsStr == "" {
		return []string{}, nil
	}

	tags := strings.Split(tagsStr, "\n")

	// limit個まで制限
	if len(tags) > limit {
		tags = tags[:limit]
	}

	return tags, nil
}

// createReleaseNotes はGitHubリリースノートを作成します。
//
// パラメータ:
//   - tag: リリースを作成するタグ
//   - draft: ドラフトとして作成するかどうか
//   - prerelease: プレリリースとして作成するかどうか
//
// 戻り値:
//   - error: エラーが発生した場合のエラー情報
func createReleaseNotes(tag string, draft, prerelease bool) error {
	args := []string{"release", "create", tag, "--generate-notes"}

	if draft {
		args = append(args, "--draft")
	}
	if prerelease {
		args = append(args, "--prerelease")
	}

	cmd := exec.Command("gh", args...)

	// 出力をキャプチャして表示
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// エラーメッセージを表示
		if stderr.Len() > 0 {
			return fmt.Errorf("%w\n%s", err, stderr.String())
		}
		return err
	}

	// 成功メッセージを表示
	if stdout.Len() > 0 {
		fmt.Println(stdout.String())
	}

	return nil
}

// init は release-notes コマンドを root コマンドに登録し、フラグを設定します。
func init() {
	releaseNotesCmd.Flags().StringVar(&releaseTag, "tag", "", "リリースを作成するタグを指定")
	releaseNotesCmd.Flags().BoolVar(&releaseDraft, "draft", false, "ドラフトとして作成")
	releaseNotesCmd.Flags().BoolVar(&releasePrerelease, "prerelease", false, "プレリリースとして作成")
	releaseNotesCmd.Flags().BoolVar(&releaseLatest, "latest", false, "最新タグからリリースを作成")
	rootCmd.AddCommand(releaseNotesCmd)
}
