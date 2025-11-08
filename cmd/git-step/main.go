package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AuthorStats はユーザーごとの統計情報
type AuthorStats struct {
	Name    string
	Added   int
	Deleted int
	Net     int // 純増行数 (Added - Deleted)
}

func main() {
	// コマンドライン引数を解析
	sinceArg := ""
	untilArg := ""
	weeks := 0
	months := 0
	years := 0
	excludeInitial := true // デフォルトで初回コミットを除外

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--since", "-s":
			if i+1 < len(args) {
				sinceArg = args[i+1]
				i++
			}
		case "--until", "-u":
			if i+1 < len(args) {
				untilArg = args[i+1]
				i++
			}
		case "-w", "--weeks":
			if i+1 < len(args) {
				weeks, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "-m", "--months":
			if i+1 < len(args) {
				months, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "-y", "--years":
			if i+1 < len(args) {
				years, _ = strconv.Atoi(args[i+1])
				i++
			}
		case "--include-initial":
			excludeInitial = false
		case "-h":
			printHelp()
			return
		}
	}

	// デフォルトは全期間
	if sinceArg == "" && weeks == 0 && months == 0 && years == 0 {
		// 全期間（引数なし）
	}

	// 期間指定の優先順位: -w/-m/-y > --since
	if weeks > 0 || months > 0 || years > 0 {
		now := time.Now()
		if years > 0 {
			sinceArg = now.AddDate(-years, 0, 0).Format("2006-01-02")
		} else if months > 0 {
			sinceArg = now.AddDate(0, -months, 0).Format("2006-01-02")
		} else if weeks > 0 {
			sinceArg = now.AddDate(0, 0, -weeks*7).Format("2006-01-02")
		}
	}

	// 作成者ごとの統計を取得
	authorStats := getAuthorStats(sinceArg, untilArg, excludeInitial)

	if len(authorStats) == 0 {
		fmt.Println("コミットが見つかりませんでした。")
		return
	}

	// 全体の統計を計算
	totalAdded := 0
	totalDeleted := 0
	totalNet := 0
	for _, stat := range authorStats {
		totalAdded += stat.Added
		totalDeleted += stat.Deleted
		totalNet += stat.Net
	}

	// 現在のリポジトリの総行数を取得
	currentLines := getCurrentTotalLines()

	// コード割合が多い順にソート
	sort.Slice(authorStats, func(i, j int) bool {
		ratioI := 0.0
		ratioJ := 0.0
		if currentLines > 0 {
			ratioI = float64(authorStats[i].Added) / float64(currentLines)
			ratioJ = float64(authorStats[j].Added) / float64(currentLines)
		}
		return ratioI > ratioJ
	})

	// 結果を表示
	displayStats(authorStats, totalAdded, totalDeleted, totalNet, currentLines, sinceArg, untilArg)

	// ファイルに保存
	saveToFile(authorStats, totalAdded, totalDeleted, totalNet, currentLines, sinceArg, untilArg)
}

func printHelp() {
	help := `git step - リポジトリのステップ数とユーザーごとの貢献度を表示

使い方:
  git step                    # 全期間のステップ数を表示
  git step -w 1               # 過去1週間
  git step -m 1               # 過去1ヶ月
  git step -y 1               # 過去1年
  git step --since 2024-01-01 # 指定日以降
  git step --include-initial  # 初回コミットを含める

オプション:
  -w, --weeks <数>        過去N週間を集計
  -m, --months <数>       過去Nヶ月を集計
  -y, --years <数>        過去N年を集計
  --since, -s <日時>      集計開始日時
  --until, -u <日時>      集計終了日時
  --include-initial       初回コミットを含める（デフォルトは除外）
  -h                      このヘルプを表示

説明:
  リポジトリ全体のステップ数（行数）とユーザーごとの貢献度を集計します。
  デフォルトで初回コミットは除外されます（大量の行数が追加されることが多いため）。
  結果はステップ数が多い順に表示され、自動的にファイルに保存されます。
`
	fmt.Print(help)
}

func getAuthorStats(since, until string, excludeInitial bool) []AuthorStats {
	// 初回コミットのハッシュを取得（除外する場合）
	var initialCommitHash string
	if excludeInitial {
		cmd := exec.Command("git", "rev-list", "--max-parents=0", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			initialCommitHash = strings.TrimSpace(string(output))
		}
	}

	// 作成者ごとに集計
	authorMap := make(map[string]*AuthorStats)

	// git log --numstat の出力を解析
	// コミットごとに作成者を特定する必要がある
	cmdArgs := []string{"log", "--all", "--pretty=format:%H%x09%an", "--numstat"}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since="+since)
	}
	if until != "" {
		cmdArgs = append(cmdArgs, "--until="+until)
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git log の実行に失敗しました: %v\n", err)
		return nil
	}

	lines := strings.Split(string(output), "\n")
	currentAuthor := ""
	currentHash := ""

	for _, line := range lines {
		if line == "" {
			continue
		}

		// コミット行（ハッシュ\t作成者）を検出
		// タブが1つだけ含まれ、スペースで始まらない行
		if !strings.HasPrefix(line, " ") && strings.Count(line, "\t") == 1 {
			parts := strings.Split(line, "\t")
			if len(parts) == 2 {
				currentHash = parts[0]
				currentAuthor = parts[1]

				// 初回コミットをスキップ
				if excludeInitial && currentHash == initialCommitHash {
					currentAuthor = ""
					continue
				}

				if _, exists := authorMap[currentAuthor]; !exists {
					authorMap[currentAuthor] = &AuthorStats{Name: currentAuthor}
				}
			}
			continue
		}

		// numstat 行（追加\t削除\tファイル名）
		if currentAuthor != "" && strings.Contains(line, "\t") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				added, err1 := strconv.Atoi(parts[0])
				deleted, err2 := strconv.Atoi(parts[1])

				// バイナリファイルは "-" と表示されるのでスキップ
				if err1 == nil && err2 == nil {
					authorMap[currentAuthor].Added += added
					authorMap[currentAuthor].Deleted += deleted
					authorMap[currentAuthor].Net += (added - deleted)
				}
			}
		}
	}

	// マップをスライスに変換
	stats := make([]AuthorStats, 0, len(authorMap))
	for _, stat := range authorMap {
		stats = append(stats, *stat)
	}

	return stats
}

func getCurrentTotalLines() int {
	// git ls-files | xargs cat | wc -l
	// Windows環境を考慮して、Goで実装
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
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
		totalLines += len(lines)
	}

	return totalLines
}

func displayStats(stats []AuthorStats, totalAdded, totalDeleted, totalNet, currentLines int, since, until string) {
	fmt.Println("=== リポジトリステップ数統計 ===")
	fmt.Println()

	// 期間情報
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

	// 現在の総行数
	fmt.Printf("現在のリポジトリ総行数: %s 行\n", formatNumber(currentLines))
	fmt.Println()

	// 全体統計
	fmt.Println("【全体統計】")
	fmt.Printf("  追加行数: %s\n", formatNumber(totalAdded))
	fmt.Printf("  削除行数: %s\n", formatNumber(totalDeleted))
	fmt.Printf("  純増行数: %s\n", formatNumber(totalNet))
	fmt.Println()

	// ユーザー別統計
	fmt.Println("【ユーザー別統計】（コード割合が多い順）")
	fmt.Println()
	fmt.Printf("%-30s %12s %12s %10s\n", "作成者", "追加", "削除", "コード割合")
	fmt.Println(strings.Repeat("-", 70))

	for _, stat := range stats {
		codeRatio := 0.0
		if currentLines > 0 {
			codeRatio = float64(stat.Added) / float64(currentLines) * 100
		}

		fmt.Printf("%-30s %12s %12s %9.1f%%\n",
			stat.Name,
			formatNumber(stat.Added),
			formatNumber(stat.Deleted),
			codeRatio,
		)
	}

	fmt.Println()
}

func saveToFile(stats []AuthorStats, totalAdded, totalDeleted, totalNet, currentLines int, since, until string) {
	// ファイル名を生成
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

	// ファイルに書き込み
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

	fmt.Fprintf(file, "現在のリポジトリ総行数: %s 行\n", formatNumber(currentLines))
	fmt.Fprintln(file)

	fmt.Fprintln(file, "【全体統計】")
	fmt.Fprintf(file, "  追加行数: %s\n", formatNumber(totalAdded))
	fmt.Fprintf(file, "  削除行数: %s\n", formatNumber(totalDeleted))
	fmt.Fprintf(file, "  純増行数: %s\n", formatNumber(totalNet))
	fmt.Fprintln(file)

	fmt.Fprintln(file, "【ユーザー別統計】（コード割合が多い順）")
	fmt.Fprintln(file)
	fmt.Fprintf(file, "%-30s %12s %12s %10s\n", "作成者", "追加", "削除", "コード割合")
	fmt.Fprintln(file, strings.Repeat("-", 70))

	for _, stat := range stats {
		codeRatio := 0.0
		if currentLines > 0 {
			codeRatio = float64(stat.Added) / float64(currentLines) * 100
		}

		fmt.Fprintf(file, "%-30s %12s %12s %9.1f%%\n",
			stat.Name,
			formatNumber(stat.Added),
			formatNumber(stat.Deleted),
			codeRatio,
		)
	}

	fmt.Printf("\n結果を %s に保存しました。\n", filename)
}

func formatNumber(n int) string {
	s := strconv.Itoa(n)
	if n < 0 {
		s = s[1:]
		defer func() { s = "-" + s }()
	}

	result := ""
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}

	return result
}
