/*
Package tag は git の拡張コマンドのうち、タグ関連コマンドを定義します。

このファイル (tag_diff_all.go) は、全てのタグ間の差分を一括取得するコマンドを提供します。
リポジトリ内の全てのタグを時系列順にソートし、連続するタグ間の差分を一括で出力します。

主な機能:
  - 全タグの自動取得と時系列ソート
  - 連続するタグ間の差分を一括出力
  - マージコミットの自動除外
  - 単一ファイルまたは複数ファイルへの出力
  - タグプレフィックスによるフィルタリング
  - 詳細なサマリー情報

使用例:

	git tag-diff-all                    # 全タグ間の差分を取得
	git tag-diff-all --prefix=V4        # V4で始まるタグのみ
	git tag-diff-all --split            # タグペアごとにファイル分割
	git tag-diff-all --output=diff.txt  # 出力ファイル名を指定
*/
package tag

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// タグ情報を保持する構造体
type tagInfo struct {
	Name string
	Date time.Time
}

// tagDiffAllCmd は全てのタグ間の差分を一括取得するコマンドです。
var tagDiffAllCmd = &cobra.Command{
	Use:   "tag-diff-all",
	Short: "全タグ間の差分を一括取得",
	Long: `リポジトリ内の全てのタグを時系列順にソートし、
連続するタグ間のコミット差分を一括でファイルに出力します。

Mergeコミットは自動的に除外されます。
出力形式: - コミットメッセージ (作成者名, 日付)

各セクションにはタグ名、コミット数、統計情報が含まれます。`,
	Example: `  git tag-diff-all                    # 全タグ間の差分を取得
  git tag-diff-all --prefix=V4        # V4で始まるタグのみ
  git tag-diff-all --split            # タグペアごとにファイル分割
  git tag-diff-all --output=diff.txt  # 出力ファイル名を指定
  git tag-diff-all --limit=10         # 最新10タグ間の差分のみ`,
	RunE: runTagDiffAll,
}

var (
	tagDiffPrefix  string
	tagDiffOutput  string
	tagDiffSplit   bool
	tagDiffLimit   int
	tagDiffReverse bool
)

func init() {
	cmd.RootCmd.AddCommand(tagDiffAllCmd)

	tagDiffAllCmd.Flags().StringVarP(&tagDiffPrefix, "prefix", "p", "", "タグ名のプレフィックスでフィルタリング")
	tagDiffAllCmd.Flags().StringVarP(&tagDiffOutput, "output", "o", "tag_diff_all.txt", "出力ファイル名")
	tagDiffAllCmd.Flags().BoolVarP(&tagDiffSplit, "split", "s", false, "タグペアごとにファイルを分割")
	tagDiffAllCmd.Flags().IntVarP(&tagDiffLimit, "limit", "l", 0, "処理するタグ数の上限（0=無制限）")
	tagDiffAllCmd.Flags().BoolVarP(&tagDiffReverse, "reverse", "r", false, "新しいタグから古いタグの順で出力")
}

func runTagDiffAll(c *cobra.Command, args []string) error {
	// 全タグを取得
	tags, err := getAllTags()
	if err != nil {
		return fmt.Errorf("タグの取得に失敗しました: %w", err)
	}

	if len(tags) < 2 {
		return fmt.Errorf("差分を取得するには少なくとも2つのタグが必要です（現在: %d個）", len(tags))
	}

	// プレフィックスでフィルタリング
	if tagDiffPrefix != "" {
		tags = filterTagsByPrefix(tags, tagDiffPrefix)
		if len(tags) < 2 {
			return fmt.Errorf("プレフィックス '%s' に一致するタグが2つ未満です", tagDiffPrefix)
		}
	}

	// タグ数の制限
	if tagDiffLimit > 0 && len(tags) > tagDiffLimit {
		tags = tags[len(tags)-tagDiffLimit:]
	}

	fmt.Printf("🔍 %d個のタグを検出しました\n", len(tags))
	fmt.Printf("📊 %d個のタグペアの差分を取得します\n\n", len(tags)-1)

	// 差分を取得
	if tagDiffSplit {
		return generateSplitFiles(tags)
	}
	return generateSingleFile(tags)
}

// getAllTags はリポジトリ内の全タグを時系列順で取得します
func getAllTags() ([]tagInfo, error) {
	// タグ一覧を日付付きで取得
	output, err := gitcmd.Run("tag", "-l", "--format=%(refname:short)|%(creatordate:iso)")
	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, fmt.Errorf("タグが見つかりません")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	tags := make([]tagInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		tagName := parts[0]

		var tagDate time.Time
		if len(parts) == 2 && parts[1] != "" {
			// 日付をパース
			parsedDate, err := parseGitDate(parts[1])
			if err == nil {
				tagDate = parsedDate
			} else {
				// パースに失敗した場合はコミット日を使用
				tagDate = getTagCommitDate(tagName)
			}
		} else {
			tagDate = getTagCommitDate(tagName)
		}

		tags = append(tags, tagInfo{
			Name: tagName,
			Date: tagDate,
		})
	}

	// 日付でソート（古い順）
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Date.Before(tags[j].Date)
	})

	return tags, nil
}

// parseGitDate はgitの日付文字列をパースします
func parseGitDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	layouts := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 MST",
		time.RFC3339,
		"Mon Jan 2 15:04:05 2006 -0700",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("日付のパースに失敗: %s", dateStr)
}

// getTagCommitDate はタグが指すコミットの日付を取得します
func getTagCommitDate(tag string) time.Time {
	output, err := gitcmd.Run("log", "-1", "--format=%ci", tag)
	if err != nil {
		return time.Time{}
	}

	dateStr := strings.TrimSpace(string(output))
	t, _ := parseGitDate(dateStr)
	return t
}

// filterTagsByPrefix はプレフィックスでタグをフィルタリングします
func filterTagsByPrefix(tags []tagInfo, prefix string) []tagInfo {
	filtered := make([]tagInfo, 0)
	for _, tag := range tags {
		if strings.HasPrefix(tag.Name, prefix) {
			filtered = append(filtered, tag)
		}
	}
	return filtered
}

// generateSingleFile は全差分を1つのファイルに出力します
func generateSingleFile(tags []tagInfo) error {
	var builder strings.Builder
	totalCommits := 0
	processedPairs := 0

	// ヘッダー
	builder.WriteString("================================================================================\n")
	builder.WriteString("タグ間差分レポート\n")
	builder.WriteString(fmt.Sprintf("生成日時: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("タグ数: %d\n", len(tags)))
	builder.WriteString(fmt.Sprintf("タグペア数: %d\n", len(tags)-1))
	builder.WriteString("================================================================================\n\n")

	// 各タグペアの差分を取得
	pairs := generateTagPairs(tags)
	for i, pair := range pairs {
		oldTag := pair[0]
		newTag := pair[1]

		fmt.Printf("  [%d/%d] %s → %s ... ", i+1, len(pairs), oldTag.Name, newTag.Name)

		diff, commitCount, err := getTagDiff(oldTag.Name, newTag.Name)
		if err != nil {
			fmt.Printf("エラー\n")
			builder.WriteString(fmt.Sprintf("## %s → %s\n", oldTag.Name, newTag.Name))
			builder.WriteString(fmt.Sprintf("エラー: %s\n\n", err.Error()))
			continue
		}

		fmt.Printf("%d コミット\n", commitCount)

		// セクションヘッダー
		builder.WriteString("--------------------------------------------------------------------------------\n")
		builder.WriteString(fmt.Sprintf("## %s → %s\n", oldTag.Name, newTag.Name))
		builder.WriteString(fmt.Sprintf("   期間: %s → %s\n", oldTag.Date.Format("2006-01-02"), newTag.Date.Format("2006-01-02")))
		builder.WriteString(fmt.Sprintf("   コミット数: %d\n", commitCount))
		builder.WriteString("--------------------------------------------------------------------------------\n")

		if diff == "" {
			builder.WriteString("(差分なし)\n")
		} else {
			builder.WriteString(diff)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")

		totalCommits += commitCount
		processedPairs++
	}

	// サマリー
	builder.WriteString("================================================================================\n")
	builder.WriteString("サマリー\n")
	builder.WriteString("================================================================================\n")
	builder.WriteString(fmt.Sprintf("処理済みタグペア: %d\n", processedPairs))
	builder.WriteString(fmt.Sprintf("総コミット数: %d\n", totalCommits))
	builder.WriteString(fmt.Sprintf("平均コミット数/ペア: %.1f\n", float64(totalCommits)/float64(processedPairs)))

	// ファイルに書き込み
	absPath, err := filepath.Abs(tagDiffOutput)
	if err != nil {
		return fmt.Errorf("ファイルパスの取得に失敗しました: %w", err)
	}

	if err := os.WriteFile(tagDiffOutput, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("ファイルへの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("\n✓ 全差分を %s に出力しました。\n", absPath)
	fmt.Printf("  総コミット数: %d\n", totalCommits)

	return nil
}

// generateSplitFiles はタグペアごとに別ファイルに出力します
func generateSplitFiles(tags []tagInfo) error {
	totalCommits := 0
	outputDir := "tag_diffs"

	// 出力ディレクトリを作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	pairs := generateTagPairs(tags)
	for i, pair := range pairs {
		oldTag := pair[0]
		newTag := pair[1]

		fmt.Printf("  [%d/%d] %s → %s ... ", i+1, len(pairs), oldTag.Name, newTag.Name)

		diff, commitCount, err := getTagDiff(oldTag.Name, newTag.Name)
		if err != nil {
			fmt.Printf("エラー\n")
			continue
		}

		fmt.Printf("%d コミット\n", commitCount)

		// ファイル名の生成（タグ名の/を_に置換）
		safeName := fmt.Sprintf("diff_%s_to_%s.txt",
			strings.ReplaceAll(oldTag.Name, "/", "_"),
			strings.ReplaceAll(newTag.Name, "/", "_"))
		filePath := filepath.Join(outputDir, safeName)

		// ファイル内容
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("タグ差分: %s → %s\n", oldTag.Name, newTag.Name))
		builder.WriteString(fmt.Sprintf("期間: %s → %s\n", oldTag.Date.Format("2006-01-02"), newTag.Date.Format("2006-01-02")))
		builder.WriteString(fmt.Sprintf("コミット数: %d\n", commitCount))
		builder.WriteString("--------------------------------------------------------------------------------\n")
		if diff == "" {
			builder.WriteString("(差分なし)\n")
		} else {
			builder.WriteString(diff)
		}

		if err := os.WriteFile(filePath, []byte(builder.String()), 0644); err != nil {
			return fmt.Errorf("ファイル %s への書き込みに失敗しました: %w", filePath, err)
		}

		totalCommits += commitCount
	}

	absPath, _ := filepath.Abs(outputDir)
	fmt.Printf("\n✓ 全差分を %s/ に出力しました。\n", absPath)
	fmt.Printf("  ファイル数: %d\n", len(pairs))
	fmt.Printf("  総コミット数: %d\n", totalCommits)

	return nil
}

// generateTagPairs はタグのペアを生成します
func generateTagPairs(tags []tagInfo) [][2]tagInfo {
	pairs := make([][2]tagInfo, 0, len(tags)-1)

	if tagDiffReverse {
		// 新しい→古い順
		for i := len(tags) - 1; i > 0; i-- {
			pairs = append(pairs, [2]tagInfo{tags[i-1], tags[i]})
		}
	} else {
		// 古い→新しい順
		for i := 0; i < len(tags)-1; i++ {
			pairs = append(pairs, [2]tagInfo{tags[i], tags[i+1]})
		}
	}

	return pairs
}

// getTagDiff は2つのタグ間の差分を取得します
func getTagDiff(oldTag, newTag string) (string, int, error) {
	tagRange := fmt.Sprintf("%s..%s", oldTag, newTag)
	output, err := gitcmd.Run("log", tagRange, "--no-merges", "--pretty=format:- %s (%an, %ad)", "--date=short")
	if err != nil {
		return "", 0, err
	}

	if len(output) == 0 {
		return "", 0, nil
	}

	diff := strings.TrimSpace(string(output))
	commitCount := len(strings.Split(diff, "\n"))

	return diff, commitCount, nil
}
