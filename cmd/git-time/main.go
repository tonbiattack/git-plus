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

type CommitInfo struct {
	Hash      string
	Branch    string
	Author    string
	Timestamp time.Time
	Message   string
	Files     []string // 変更されたファイルのリスト
}

type BranchStat struct {
	Name        string
	CommitCount int
	TotalHours  float64
}

type CommitStat struct {
	Message   string
	Branch    string
	Hours     float64
	Timestamp time.Time
	Files     []string // 変更されたファイルのリスト
}

func main() {
	// コマンドライン引数を解析
	sinceArg := ""
	untilArg := ""
	showCommits := false
	weeks := 0
	months := 0
	years := 0

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
		case "--commits", "-c":
			showCommits = true
		case "-w", "--weeks":
			if i+1 < len(args) {
				w, err := strconv.Atoi(args[i+1])
				if err == nil {
					weeks = w
				}
				i++
			}
		case "-m", "--months":
			if i+1 < len(args) {
				m, err := strconv.Atoi(args[i+1])
				if err == nil {
					months = m
				}
				i++
			}
		case "-y", "--years":
			if i+1 < len(args) {
				y, err := strconv.Atoi(args[i+1])
				if err == nil {
					years = y
				}
				i++
			}
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// -w, -m, -y オプションから期間を計算
	if weeks > 0 {
		sinceArg = fmt.Sprintf("%d weeks ago", weeks)
	} else if months > 0 {
		sinceArg = fmt.Sprintf("%d months ago", months)
	} else if years > 0 {
		sinceArg = fmt.Sprintf("%d years ago", years)
	}

	// デフォルトは1週間前から
	if sinceArg == "" {
		sinceArg = "1 week ago"
	}

	fmt.Printf("コミット履歴を分析しています（%s から", sinceArg)
	if untilArg != "" {
		fmt.Printf(" %s まで", untilArg)
	}
	fmt.Println("）...")

	// コミット履歴を取得
	commits, err := getCommits(sinceArg, untilArg)
	if err != nil {
		fmt.Printf("エラー: コミット履歴の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("指定された期間にコミットが見つかりませんでした。")
		os.Exit(0)
	}

	fmt.Printf("合計 %d 件のコミットが見つかりました。\n\n", len(commits))

	// 出力ファイル名を生成
	outputFile := generateOutputFileName(sinceArg, untilArg)

	if showCommits {
		// コミットごとの集計
		displayCommitStats(commits, outputFile)
	} else {
		// ブランチごとの集計
		displayBranchStats(commits, outputFile)
	}
}

// generateOutputFileName は出力ファイル名を生成する
func generateOutputFileName(since, until string) string {
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 期間をファイル名に含める
	periodStr := strings.ReplaceAll(since, " ", "_")
	periodStr = strings.ReplaceAll(periodStr, ":", "-")

	return fmt.Sprintf("git_time_%s_%s.txt", periodStr, dateStr)
}

// getCommits は指定された期間のコミット履歴を取得する
func getCommits(since, until string) ([]CommitInfo, error) {
	args := []string{
		"log",
		"--all",
		"--format=%H|%D|%an|%at|%s",
	}

	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		hash := parts[0]
		refs := parts[1]
		author := parts[2]
		timestampStr := parts[3]
		message := parts[4]

		// タイムスタンプを解析
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}

		// ブランチ名を抽出
		branch := extractBranch(refs, hash)

		// stashとdetachedは除外
		if branch == "stash" || branch == "detached" {
			continue
		}

		// 変更されたファイルを取得
		files := getChangedFiles(hash)

		commits = append(commits, CommitInfo{
			Hash:      hash,
			Branch:    branch,
			Author:    author,
			Timestamp: time.Unix(timestamp, 0),
			Message:   message,
			Files:     files,
		})
	}

	// 時系列順にソート（古い順）
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Timestamp.Before(commits[j].Timestamp)
	})

	return commits, nil
}

// getChangedFiles は指定されたコミットで変更されたファイルのリストを取得する
func getChangedFiles(hash string) []string {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", hash)
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := make([]string, 0, len(files))
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file != "" {
			result = append(result, file)
		}
	}
	return result
}

// extractBranch は refs から ブランチ名を抽出する
func extractBranch(refs, hash string) string {
	if refs == "" {
		// refsがない場合は、コミットが属するブランチを取得
		branch := getBranchForCommit(hash)
		if branch != "" {
			return branch
		}
		return "detached"
	}

	// "HEAD -> main, origin/main" のような形式から解析
	parts := strings.Split(refs, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// "HEAD -> branch" の形式
		if strings.Contains(part, "HEAD -> ") {
			branch := strings.TrimPrefix(part, "HEAD -> ")
			return strings.TrimSpace(branch)
		}

		// "origin/branch" の形式をスキップして、ローカルブランチを優先
		if !strings.Contains(part, "/") && !strings.HasPrefix(part, "tag:") {
			return part
		}
	}

	// ローカルブランチが見つからない場合、最初のリモートブランチを使用
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "/") && !strings.HasPrefix(part, "tag:") {
			// "origin/main" -> "main"
			parts := strings.SplitN(part, "/", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	return "unknown"
}

// getBranchForCommit はコミットハッシュから所属ブランチを取得
func getBranchForCommit(hash string) string {
	cmd := exec.Command("git", "branch", "--contains", hash, "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		return lines[0]
	}

	return ""
}

// displayBranchStats はブランチごとの統計を表示する
func displayBranchStats(commits []CommitInfo, outputFile string) {
	branchMap := make(map[string]*BranchStat)

	// ブランチごとにコミットをグループ化
	for i, commit := range commits {
		if _, exists := branchMap[commit.Branch]; !exists {
			branchMap[commit.Branch] = &BranchStat{
				Name:        commit.Branch,
				CommitCount: 0,
				TotalHours:  0,
			}
		}

		branchMap[commit.Branch].CommitCount++

		// 次のコミットとの時間差を計算（同じ作業セッションとみなす上限は2時間）
		if i+1 < len(commits) {
			nextCommit := commits[i+1]
			duration := nextCommit.Timestamp.Sub(commit.Timestamp)
			hours := duration.Hours()

			// 2時間以内の差であれば作業時間とみなす
			if hours <= 2.0 && hours >= 0 {
				branchMap[commit.Branch].TotalHours += hours
			} else {
				// 2時間を超える場合は、デフォルトで30分と見積もる
				branchMap[commit.Branch].TotalHours += 0.5
			}
		} else {
			// 最後のコミットはデフォルトで30分
			branchMap[commit.Branch].TotalHours += 0.5
		}
	}

	// 統計を配列に変換してソート（最新のコミット順）
	stats := make([]BranchStat, 0, len(branchMap))

	// 各ブランチの最新コミット時刻を記録
	branchLastCommit := make(map[string]time.Time)
	for i := len(commits) - 1; i >= 0; i-- {
		commit := commits[i]
		if _, exists := branchLastCommit[commit.Branch]; !exists {
			branchLastCommit[commit.Branch] = commit.Timestamp
		}
	}

	for _, stat := range branchMap {
		stats = append(stats, *stat)
	}

	// 最新のコミット時刻順にソート（新しい順）
	sort.Slice(stats, func(i, j int) bool {
		return branchLastCommit[stats[i].Name].After(branchLastCommit[stats[j].Name])
	})

	// 出力内容を構築
	var output strings.Builder
	output.WriteString("ブランチごとの作業時間:\n")
	output.WriteString(strings.Repeat("-", 60) + "\n")

	totalCommits := 0
	totalHours := 0.0

	for _, stat := range stats {
		line := fmt.Sprintf("%-30s %3d commits (約%.1fh)\n",
			stat.Name, stat.CommitCount, stat.TotalHours)
		output.WriteString(line)
		totalCommits += stat.CommitCount
		totalHours += stat.TotalHours
	}

	output.WriteString(strings.Repeat("-", 60) + "\n")
	output.WriteString(fmt.Sprintf("合計: %d commits (約%.1fh)\n", totalCommits, totalHours))

	// コンソールに表示
	fmt.Print(output.String())

	// ファイルに出力
	if err := os.WriteFile(outputFile, []byte(output.String()), 0644); err != nil {
		fmt.Printf("\n警告: ファイルへの書き込みに失敗しました: %v\n", err)
	} else {
		fmt.Printf("\n✓ 結果を %s に出力しました。\n", outputFile)
	}
}

// displayCommitStats はコミットごとの統計を表示する
func displayCommitStats(commits []CommitInfo, outputFile string) {
	commitStats := make([]CommitStat, 0, len(commits))

	for i, commit := range commits {
		hours := 0.5 // デフォルト30分

		// 次のコミットとの時間差を計算
		if i+1 < len(commits) {
			nextCommit := commits[i+1]
			duration := nextCommit.Timestamp.Sub(commit.Timestamp)
			h := duration.Hours()

			// 2時間以内の差であれば作業時間とみなす
			if h <= 2.0 && h >= 0 {
				hours = h
			}
		}

		commitStats = append(commitStats, CommitStat{
			Message:   commit.Message,
			Branch:    commit.Branch,
			Hours:     hours,
			Timestamp: commit.Timestamp,
			Files:     commit.Files,
		})
	}

	// コミット時刻順にソート（新しい順）
	sort.Slice(commitStats, func(i, j int) bool {
		return commitStats[i].Timestamp.After(commitStats[j].Timestamp)
	})

	// 出力ファイル名を生成（commits用）
	commitsOutputFile := strings.Replace(outputFile, ".txt", "_commits.txt", 1)

	// 出力内容を構築
	var output strings.Builder
	output.WriteString("コミットごとの作業時間（直近順）:\n")
	output.WriteString(strings.Repeat("-", 80) + "\n")

	displayCount := 20
	if len(commitStats) < displayCount {
		displayCount = len(commitStats)
	}

	for i := 0; i < displayCount; i++ {
		stat := commitStats[i]
		dateStr := stat.Timestamp.Format("2006-01-02 15:04")

		// ファイル名のリストを整形
		filesStr := ""
		if len(stat.Files) > 0 {
			// ファイルが多い場合は最初の3つのみ表示
			displayFiles := stat.Files
			hasMore := false
			if len(displayFiles) > 3 {
				displayFiles = stat.Files[:3]
				hasMore = true
			}
			filesStr = strings.Join(displayFiles, ", ")
			if hasMore {
				filesStr += fmt.Sprintf(", ... (+%d)", len(stat.Files)-3)
			}
		}

		// ブランチとファイル名を分けて表示
		line := fmt.Sprintf("[%.1fh] %-20s %s (%s)\n       ファイル: %s\n",
			stat.Hours, stat.Branch, stat.Message, dateStr, filesStr)
		output.WriteString(line)
	}

	if len(commitStats) > displayCount {
		output.WriteString(fmt.Sprintf("\n... 他 %d 件\n", len(commitStats)-displayCount))
	}

	// コンソールに表示
	fmt.Print(output.String())

	// ファイルに出力
	if err := os.WriteFile(commitsOutputFile, []byte(output.String()), 0644); err != nil {
		fmt.Printf("\n警告: ファイルへの書き込みに失敗しました: %v\n", err)
	} else {
		fmt.Printf("\n✓ 結果を %s に出力しました。\n", commitsOutputFile)
	}
}

// printHelp はヘルプメッセージを表示する
func printHelp() {
	fmt.Print(`git time - コミット履歴から作業時間を自動集計

使用方法:
  git time [オプション]

オプション:
  -w, --weeks <数>      過去N週間の作業時間を集計
                        例: -w 1 (1週間), -w 2 (2週間)
  
  -m, --months <数>     過去Nヶ月間の作業時間を集計
                        例: -m 1 (1ヶ月), -m 3 (3ヶ月)
  
  -y, --years <数>      過去N年間の作業時間を集計
                        例: -y 1 (1年)
  
  --since, -s <日時>    集計開始日時
                        例: "2 weeks ago", "2024-01-01", "yesterday"
  
  --until, -u <日時>    集計終了日時 (デフォルト: 現在)
                        例: "yesterday", "2024-12-31"
  
  --commits, -c         ブランチ別ではなくコミット別に表示
  
  --help, -h            このヘルプを表示

使用例:
  git time                # 過去1週間の作業時間をブランチ別に表示
  git time -w 1           # 過去1週間の作業時間
  git time -m 1           # 過去1ヶ月の作業時間
  git time -y 1           # 過去1年の作業時間
  git time -w 2 -c        # 過去2週間をコミット別に表示
  git time --since "2024-01-01" --until "2024-12-31"  # 期間指定

作業時間の計算方法:
  - 連続するコミット間の時間差が2時間以内の場合: その時間を作業時間とする
  - 2時間を超える場合: デフォルトで30分と見積もる
  - 最後のコミット: デフォルトで30分

出力:
  - コンソールに結果を表示
  - 自動的にファイル(git_time_*.txt)に結果を保存
`)
}
