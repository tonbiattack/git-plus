package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// AuthorStats はユーザーごとの統計情報を表す構造体
type AuthorStats struct {
	Name          string
	Added         int
	Deleted       int
	Net           int
	Modified      int
	CurrentCode   int
	Commits       int
	AvgCommitSize float64
}

var (
	stepSince          string
	stepUntil          string
	stepWeeks          int
	stepMonths         int
	stepYears          int
	stepIncludeInitial bool
)

var stepCmd = &cobra.Command{
	Use:   "step",
	Short: "リポジトリのステップ数とユーザーごとの貢献度を表示",
	Long: `リポジトリ全体のステップ数（行数）とユーザーごとの貢献度を集計します。
デフォルトで初回コミットは除外されます（大量の行数が追加されることが多いため）。
結果はステップ数が多い順に表示され、自動的にテキストファイルとCSVファイルに保存されます。`,
	Example: `  git-plus step                    # 全期間のステップ数を表示
  git-plus step -w 1               # 過去1週間
  git-plus step -m 1               # 過去1ヶ月
  git-plus step -y 1               # 過去1年
  git-plus step --since 2024-01-01 # 指定日以降
  git-plus step --include-initial  # 初回コミットを含める`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 期間指定の優先順位: -w/-m/-y > --since
		sinceArg := stepSince
		if stepWeeks > 0 || stepMonths > 0 || stepYears > 0 {
			now := time.Now()
			if stepYears > 0 {
				sinceArg = now.AddDate(-stepYears, 0, 0).Format("2006-01-02")
			} else if stepMonths > 0 {
				sinceArg = now.AddDate(0, -stepMonths, 0).Format("2006-01-02")
			} else if stepWeeks > 0 {
				sinceArg = now.AddDate(0, 0, -stepWeeks*7).Format("2006-01-02")
			}
		}

		// 作成者ごとの統計を取得
		authorStats := collectAuthorStats(sinceArg, stepUntil, !stepIncludeInitial)

		if len(authorStats) == 0 {
			fmt.Println("コミットが見つかりませんでした。")
			return nil
		}

		// 全体の統計を計算
		totalAdded, totalDeleted, totalNet, totalModified, totalCommits := 0, 0, 0, 0, 0
		for _, stat := range authorStats {
			totalAdded += stat.Added
			totalDeleted += stat.Deleted
			totalNet += stat.Net
			totalModified += stat.Modified
			totalCommits += stat.Commits
		}

		// 現在のリポジトリの総行数を取得
		currentLines := getTotalLines()

		// 現在のコードベースでの各ユーザーの行数を取得
		currentCodeStats := getCodeByAuthor(sinceArg, stepUntil)

		// 統計を更新
		for i := range authorStats {
			if lines, exists := currentCodeStats[authorStats[i].Name]; exists {
				authorStats[i].CurrentCode = lines
			}

			if authorStats[i].Commits > 0 {
				authorStats[i].AvgCommitSize = float64(authorStats[i].Modified) / float64(authorStats[i].Commits)
			}
		}

		// コード割合が多い順にソート
		sort.Slice(authorStats, func(i, j int) bool {
			return authorStats[i].CurrentCode > authorStats[j].CurrentCode
		})

		// 結果を表示
		showStats(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, stepUntil)

		// ファイルに保存
		saveStatsToFile(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, stepUntil)

		// CSVファイルに保存
		saveStatsToCSV(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, stepUntil)

		return nil
	},
}

func collectAuthorStats(since, until string, excludeInitial bool) []AuthorStats {
	var initialCommitHash string
	if excludeInitial {
		output, err := gitcmd.Run("rev-list", "--max-parents=0", "HEAD")
		if err == nil {
			initialCommitHash = strings.TrimSpace(string(output))
		}
	}

	authorMap := make(map[string]*AuthorStats)
	commitCountMap := make(map[string]int)

	cmdArgs := []string{"log", "--all", "--pretty=format:%H%x09%an", "--numstat"}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since="+since)
	}
	if until != "" {
		cmdArgs = append(cmdArgs, "--until="+until)
	}

	output, err := gitcmd.Run(cmdArgs...)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	currentAuthor := ""
	currentHash := ""
	processedCommits := make(map[string]bool)

	for _, line := range lines {
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, " ") && strings.Count(line, "\t") == 1 {
			parts := strings.Split(line, "\t")
			if len(parts) == 2 {
				currentHash = parts[0]
				currentAuthor = parts[1]

				if excludeInitial && currentHash == initialCommitHash {
					currentAuthor = ""
					continue
				}

				if _, exists := authorMap[currentAuthor]; !exists {
					authorMap[currentAuthor] = &AuthorStats{Name: currentAuthor}
				}

				commitKey := currentAuthor + ":" + currentHash
				if !processedCommits[commitKey] {
					commitCountMap[currentAuthor]++
					processedCommits[commitKey] = true
				}
			}
			continue
		}

		if currentAuthor != "" && strings.Contains(line, "\t") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				added, err1 := strconv.Atoi(parts[0])
				deleted, err2 := strconv.Atoi(parts[1])

				if err1 == nil && err2 == nil {
					authorMap[currentAuthor].Added += added
					authorMap[currentAuthor].Deleted += deleted
					authorMap[currentAuthor].Net += (added - deleted)
					authorMap[currentAuthor].Modified += (added + deleted)
				}
			}
		}
	}

	stats := make([]AuthorStats, 0, len(authorMap))
	for _, stat := range authorMap {
		stat.Commits = commitCountMap[stat.Name]
		stats = append(stats, *stat)
	}

	return stats
}

func getTotalLines() int {
	output, err := gitcmd.Run("ls-files")
	if err != nil {
		return 0
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	totalLines := 0

	for _, file := range files {
		if file == "" {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		totalLines += len(lines)
	}

	return totalLines
}

func getCodeByAuthor(since, until string) map[string]int {
	output, err := gitcmd.Run("ls-files")
	if err != nil {
		return make(map[string]int)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	authorLines := make(map[string]int)

	var validCommits map[string]bool
	if since != "" || until != "" {
		validCommits = getValidCommits(since, until)
	}

	for _, file := range files {
		if file == "" {
			continue
		}

		blameOutput, err := gitcmd.Run("blame", "--line-porcelain", file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(blameOutput), "\n")
		currentCommit := ""
		for _, line := range lines {
			if len(line) >= 40 {
				fields := strings.Fields(line)
				if len(fields) > 0 && len(fields[0]) == 40 {
					currentCommit = fields[0]
				}
			}

			if strings.HasPrefix(line, "author ") {
				author := strings.TrimPrefix(line, "author ")

				if validCommits != nil {
					if validCommits[currentCommit] {
						authorLines[author]++
					}
				} else {
					authorLines[author]++
				}
			}
		}
	}

	return authorLines
}

func getValidCommits(since, until string) map[string]bool {
	cmdArgs := []string{"log", "--all", "--pretty=format:%H"}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since="+since)
	}
	if until != "" {
		cmdArgs = append(cmdArgs, "--until="+until)
	}

	output, err := gitcmd.Run(cmdArgs...)
	if err != nil {
		return make(map[string]bool)
	}

	commits := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		hash := strings.TrimSpace(line)
		if hash != "" && len(hash) == 40 {
			commits[hash] = true
		}
	}

	return commits
}

func showStats(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
	fmt.Println("=== リポジトリステップ数統計 ===")
	fmt.Println()

	if since != "" || until != "" {
		fmt.Print("期間: ")
		if since != "" {
			fmt.Print(since + " から ")
		}
		if until != "" {
			fmt.Print(until + " まで")
		} else {
			fmt.Print("現在まで")
		}
		fmt.Println()
	} else {
		fmt.Println("期間: 全期間")
	}
	fmt.Println()

	fmt.Printf("現在のリポジトリ総行数: %s 行\n", formatNum(currentLines))
	fmt.Println()

	fmt.Println("【全体統計】")
	fmt.Printf("  追加行数: %s\n", formatNum(totalAdded))
	fmt.Printf("  削除行数: %s\n", formatNum(totalDeleted))
	fmt.Printf("  純増行数: %s\n", formatNum(totalNet))
	fmt.Printf("  更新行数: %s\n", formatNum(totalModified))
	fmt.Printf("  総コミット数: %s\n", formatNum(totalCommits))
	fmt.Println()

	fmt.Println("【ユーザー別統計】（コード割合が多い順）")
	fmt.Println()
	fmt.Printf("%-30s %10s %10s %10s %10s %8s %10s %8s %8s %8s %10s\n",
		"作成者", "追加", "削除", "更新", "現在", "コミ数", "平均", "追加比", "削除比", "更新比", "コード割合")
	fmt.Println(strings.Repeat("-", 138))

	totalCurrentCode := currentLines
	if since != "" || until != "" {
		totalCurrentCode = 0
		for _, stat := range stats {
			totalCurrentCode += stat.CurrentCode
		}
	}

	for _, stat := range stats {
		codeRatio := 0.0
		if totalCurrentCode > 0 {
			codeRatio = float64(stat.CurrentCode) / float64(totalCurrentCode) * 100
		}

		addedRatio := 0.0
		if totalAdded > 0 {
			addedRatio = float64(stat.Added) / float64(totalAdded) * 100
		}

		deletedRatio := 0.0
		if totalDeleted > 0 {
			deletedRatio = float64(stat.Deleted) / float64(totalDeleted) * 100
		}

		modifiedRatio := 0.0
		if totalModified > 0 {
			modifiedRatio = float64(stat.Modified) / float64(totalModified) * 100
		}

		fmt.Printf("%-30s %10s %10s %10s %10s %8s %10.0f %7.1f%% %7.1f%% %7.1f%% %9.1f%%\n",
			stat.Name,
			formatNum(stat.Added),
			formatNum(stat.Deleted),
			formatNum(stat.Modified),
			formatNum(stat.CurrentCode),
			formatNum(stat.Commits),
			stat.AvgCommitSize,
			addedRatio,
			deletedRatio,
			modifiedRatio,
			codeRatio,
		)
	}

	fmt.Println()
}

func saveStatsToFile(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
	filename := "git_step"
	if since != "" {
		filename += "_" + strings.ReplaceAll(since, "-", "")
	} else {
		filename += "_all"
	}
	filename += "_" + time.Now().Format("2006-01-02") + ".txt"

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ファイルの作成に失敗しました: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintln(file, "=== リポジトリステップ数統計 ===")
	fmt.Fprintln(file)

	if since != "" || until != "" {
		fmt.Fprint(file, "期間: ")
		if since != "" {
			fmt.Fprint(file, since+" から ")
		}
		if until != "" {
			fmt.Fprint(file, until+" まで")
		} else {
			fmt.Fprint(file, "現在まで")
		}
		fmt.Fprintln(file)
	} else {
		fmt.Fprintln(file, "期間: 全期間")
	}
	fmt.Fprintln(file)

	fmt.Fprintf(file, "現在のリポジトリ総行数: %s 行\n", formatNum(currentLines))
	fmt.Fprintln(file)

	fmt.Fprintln(file, "【全体統計】")
	fmt.Fprintf(file, "  追加行数: %s\n", formatNum(totalAdded))
	fmt.Fprintf(file, "  削除行数: %s\n", formatNum(totalDeleted))
	fmt.Fprintf(file, "  純増行数: %s\n", formatNum(totalNet))
	fmt.Fprintf(file, "  更新行数: %s\n", formatNum(totalModified))
	fmt.Fprintf(file, "  総コミット数: %s\n", formatNum(totalCommits))
	fmt.Fprintln(file)

	fmt.Fprintln(file, "【ユーザー別統計】（コード割合が多い順）")
	fmt.Fprintln(file)
	fmt.Fprintf(file, "%-30s %10s %10s %10s %10s %8s %10s %8s %8s %8s %10s\n",
		"作成者", "追加", "削除", "更新", "現在", "コミ数", "平均", "追加比", "削除比", "更新比", "コード割合")
	fmt.Fprintln(file, strings.Repeat("-", 138))

	totalCurrentCode := currentLines
	if since != "" || until != "" {
		totalCurrentCode = 0
		for _, stat := range stats {
			totalCurrentCode += stat.CurrentCode
		}
	}

	for _, stat := range stats {
		codeRatio := 0.0
		if totalCurrentCode > 0 {
			codeRatio = float64(stat.CurrentCode) / float64(totalCurrentCode) * 100
		}

		addedRatio := 0.0
		if totalAdded > 0 {
			addedRatio = float64(stat.Added) / float64(totalAdded) * 100
		}

		deletedRatio := 0.0
		if totalDeleted > 0 {
			deletedRatio = float64(stat.Deleted) / float64(totalDeleted) * 100
		}

		modifiedRatio := 0.0
		if totalModified > 0 {
			modifiedRatio = float64(stat.Modified) / float64(totalModified) * 100
		}

		fmt.Fprintf(file, "%-30s %10s %10s %10s %10s %8s %10.0f %7.1f%% %7.1f%% %7.1f%% %9.1f%%\n",
			stat.Name,
			formatNum(stat.Added),
			formatNum(stat.Deleted),
			formatNum(stat.Modified),
			formatNum(stat.CurrentCode),
			formatNum(stat.Commits),
			stat.AvgCommitSize,
			addedRatio,
			deletedRatio,
			modifiedRatio,
			codeRatio,
		)
	}

	fmt.Printf("\n結果を %s に保存しました。\n", filename)
}

func saveStatsToCSV(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
	filename := "git_step"
	if since != "" {
		filename += "_" + strings.ReplaceAll(since, "-", "")
	} else {
		filename += "_all"
	}
	filename += "_" + time.Now().Format("2006-01-02") + ".csv"

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CSVファイルの作成に失敗しました: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintln(file, "作成者,追加行数,削除行数,純増行数,更新行数,現在行数,コミット数,平均コミットサイズ,追加比率(%),削除比率(%),更新比率(%),コード割合(%)")

	totalCurrentCode := currentLines
	if since != "" || until != "" {
		totalCurrentCode = 0
		for _, stat := range stats {
			totalCurrentCode += stat.CurrentCode
		}
	}

	for _, stat := range stats {
		codeRatio := 0.0
		if totalCurrentCode > 0 {
			codeRatio = float64(stat.CurrentCode) / float64(totalCurrentCode) * 100
		}

		addedRatio := 0.0
		if totalAdded > 0 {
			addedRatio = float64(stat.Added) / float64(totalAdded) * 100
		}

		deletedRatio := 0.0
		if totalDeleted > 0 {
			deletedRatio = float64(stat.Deleted) / float64(totalDeleted) * 100
		}

		modifiedRatio := 0.0
		if totalModified > 0 {
			modifiedRatio = float64(stat.Modified) / float64(totalModified) * 100
		}

		fmt.Fprintf(file, "%s,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			stat.Name,
			stat.Added,
			stat.Deleted,
			stat.Net,
			stat.Modified,
			stat.CurrentCode,
			stat.Commits,
			stat.AvgCommitSize,
			addedRatio,
			deletedRatio,
			modifiedRatio,
			codeRatio,
		)
	}

	fmt.Printf("CSVファイルを %s に保存しました。\n", filename)
}

func formatNum(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	s := strconv.Itoa(n)
	result := ""
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}

	return sign + result
}

func init() {
	stepCmd.Flags().StringVarP(&stepSince, "since", "s", "", "集計開始日時")
	stepCmd.Flags().StringVarP(&stepUntil, "until", "u", "", "集計終了日時")
	stepCmd.Flags().IntVarP(&stepWeeks, "weeks", "w", 0, "過去N週間を集計")
	stepCmd.Flags().IntVarP(&stepMonths, "months", "m", 0, "過去Nヶ月を集計")
	stepCmd.Flags().IntVarP(&stepYears, "years", "y", 0, "過去N年を集計")
	stepCmd.Flags().BoolVar(&stepIncludeInitial, "include-initial", false, "初回コミットを含める")
	rootCmd.AddCommand(stepCmd)
}
