package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// CommitInfo はコミット情報を表す構造体
type CommitInfo struct {
	Hash      string
	Branch    string
	Author    string
	Timestamp time.Time
	Message   string
}

// AuthorSummary はユーザーごとのサマリ統計
type AuthorSummary struct {
	Name         string
	Commits      int
	WorkHours    float64
	LinesAdded   int
	LinesDeleted int
	Branches     map[string]bool
}

func main() {
	// コマンドライン引数を解析
	sinceArg := ""
	untilArg := ""
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
		case "-h", "--help":
			printHelp()
			return
		}
	}

	// 期間指定の優先順位: -w/-m/-y > --since
	periodStr := "全期間"
	if weeks > 0 || months > 0 || years > 0 {
		now := time.Now()
		if years > 0 {
			sinceArg = now.AddDate(-years, 0, 0).Format("2006-01-02")
			periodStr = fmt.Sprintf("過去%d年", years)
		} else if months > 0 {
			sinceArg = now.AddDate(0, -months, 0).Format("2006-01-02")
			periodStr = fmt.Sprintf("過去%dヶ月", months)
		} else if weeks > 0 {
			sinceArg = now.AddDate(0, 0, -weeks*7).Format("2006-01-02")
			if weeks == 1 {
				periodStr = "過去1週間"
			} else {
				periodStr = fmt.Sprintf("過去%d週間", weeks)
			}
		}
	} else if sinceArg != "" {
		periodStr = fmt.Sprintf("%s以降", sinceArg)
	}

	// コミット情報を取得
	commits := getCommits(sinceArg, untilArg)
	if len(commits) == 0 {
		fmt.Println("コミットが見つかりませんでした。")
		return
	}

	// 作成者ごとの統計を集計
	authorStats := make(map[string]*AuthorSummary)
	totalBranches := make(map[string]bool)
	totalCommits := len(commits)
	totalHours := 0.0

	// コミットごとの作業時間と統計を計算
	commitsByAuthorBranch := groupCommitsByAuthorBranch(commits)
	for _, commitGroup := range commitsByAuthorBranch {
		// ソート: 古い順
		sort.Slice(commitGroup, func(i, j int) bool {
			return commitGroup[i].Timestamp.Before(commitGroup[j].Timestamp)
		})

		// 作業時間を計算
		for i := 0; i < len(commitGroup); i++ {
			commit := commitGroup[i]
			author := commit.Author
			branch := commit.Branch

			if _, exists := authorStats[author]; !exists {
				authorStats[author] = &AuthorSummary{
					Name:     author,
					Branches: make(map[string]bool),
				}
			}

			authorStats[author].Commits++
			authorStats[author].Branches[branch] = true
			totalBranches[branch] = true

			// 作業時間の計算
			hours := 0.5 // デフォルト
			if i < len(commitGroup)-1 {
				nextCommit := commitGroup[i+1]
				timeDiff := nextCommit.Timestamp.Sub(commit.Timestamp).Hours()
				if timeDiff <= 2.0 {
					hours = timeDiff
				}
			}
			authorStats[author].WorkHours += hours
			totalHours += hours
		}
	}

	// 行数の統計を取得
	linesStats := getLinesStats(sinceArg, untilArg)
	totalAdded := 0
	totalDeleted := 0
	for author, lines := range linesStats {
		if stat, exists := authorStats[author]; exists {
			stat.LinesAdded = lines.Added
			stat.LinesDeleted = lines.Deleted
			totalAdded += lines.Added
			totalDeleted += lines.Deleted
		}
	}

	// 1行サマリを出力
	fmt.Printf("\n%s: %dコミット / %dブランチ / 作業時間%.1fh / 変更+%d -%d行\n\n",
		periodStr, totalCommits, len(totalBranches), totalHours, totalAdded, totalDeleted)

	// ユーザーごとの統計を出力
	fmt.Println("【ユーザー別統計】")
	fmt.Println(strings.Repeat("=", 100))

	// ユーザーをコミット数でソート
	var authors []*AuthorSummary
	for _, stat := range authorStats {
		authors = append(authors, stat)
	}
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Commits > authors[j].Commits
	})

	// ヘッダー出力
	fmt.Printf("%-20s %8s %12s %10s %10s %10s\n",
		"作成者", "コミット", "作業時間(h)", "ブランチ数", "追加行", "削除行")
	fmt.Println(strings.Repeat("-", 100))

	// 各ユーザーの統計を出力
	for _, stat := range authors {
		fmt.Printf("%-20s %8d %12.1f %10d %10d %10d\n",
			stat.Name,
			stat.Commits,
			stat.WorkHours,
			len(stat.Branches),
			stat.LinesAdded,
			stat.LinesDeleted)
	}

	fmt.Println()
}

// getCommits はgit logからコミット情報を取得
func getCommits(since, until string) []CommitInfo {
	args := []string{"log", "--all", "--pretty=format:%H|%D|%an|%at|%s"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}

	output, err := gitcmd.Run(args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git logの実行に失敗: %v\n", err)
		return nil
	}

	lines := strings.Split(string(output), "\n")
	var commits []CommitInfo

	for _, line := range lines {
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

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}

		branch := extractBranch(refs, hash)
		if branch == "stash" || branch == "detached" {
			continue
		}

		commits = append(commits, CommitInfo{
			Hash:      hash,
			Branch:    branch,
			Author:    author,
			Timestamp: time.Unix(timestamp, 0),
			Message:   message,
		})
	}

	return commits
}

// extractBranch はrefsからブランチ名を抽出
func extractBranch(refs, hash string) string {
	if refs == "" {
		// git branch --contains で検索
		output, err := gitcmd.Run("branch", "--contains", hash)
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "* ") {
					return strings.TrimPrefix(line, "* ")
				}
				if line != "" {
					return line
				}
			}
		}
		return "unknown"
	}

	parts := strings.Split(refs, ",")
	var localBranches []string
	var remoteBranches []string

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "HEAD -> ") {
			return strings.TrimPrefix(part, "HEAD -> ")
		}
		if strings.Contains(part, "tag:") || strings.Contains(part, "grafted") {
			continue
		}

		if strings.HasPrefix(part, "origin/") {
			remoteBranches = append(remoteBranches, strings.TrimPrefix(part, "origin/"))
		} else if !strings.Contains(part, "/") {
			localBranches = append(localBranches, part)
		}
	}

	if len(localBranches) > 0 {
		return localBranches[0]
	}
	if len(remoteBranches) > 0 {
		return remoteBranches[0]
	}

	return "unknown"
}

// groupCommitsByAuthorBranch はコミットを作成者×ブランチでグループ化
func groupCommitsByAuthorBranch(commits []CommitInfo) map[string][]CommitInfo {
	groups := make(map[string][]CommitInfo)
	for _, commit := range commits {
		key := commit.Author + "|" + commit.Branch
		groups[key] = append(groups[key], commit)
	}
	return groups
}

// LinesStats は行数の統計
type LinesStats struct {
	Added   int
	Deleted int
}

// getLinesStats はgit logから行数の統計を取得
func getLinesStats(since, until string) map[string]LinesStats {
	args := []string{"log", "--all", "--pretty=format:%H%x09%an", "--numstat", "-m"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}

	output, err := gitcmd.Run(args...)
	if err != nil {
		return make(map[string]LinesStats)
	}

	stats := make(map[string]LinesStats)
	lines := strings.Split(string(output), "\n")
	currentAuthor := ""

	for _, line := range lines {
		if line == "" {
			continue
		}

		// コミット行: hash\tauthor（タブが1つだけ、スペースで始まらない）
		if !strings.HasPrefix(line, " ") && strings.Count(line, "\t") == 1 {
			parts := strings.Split(line, "\t")
			if len(parts) == 2 {
				currentAuthor = parts[1]
			}
			continue
		}

		// 統計行: added\tdeleted\tfilename
		if currentAuthor != "" {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				added, errA := strconv.Atoi(parts[0])
				deleted, errD := strconv.Atoi(parts[1])
				if errA == nil && errD == nil {
					stat := stats[currentAuthor]
					stat.Added += added
					stat.Deleted += deleted
					stats[currentAuthor] = stat
				}
			}
		}
	}

	return stats
}

func printHelp() {
	help := `git summary - git time + git step の結果を1行サマリで表示

使用方法:
  git summary [オプション]

オプション:
  -w, --weeks <N>    過去N週間のデータを表示
  -m, --months <N>   過去Nヶ月のデータを表示
  -y, --years <N>    過去N年のデータを表示
  -s, --since <日付> 開始日を指定（例: 2024-01-01）
  -u, --until <日付> 終了日を指定（デフォルト: 現在）
  -h, --help         ヘルプを表示

例:
  git summary -w 1              # 過去1週間の統計
  git summary -m 1              # 過去1ヶ月の統計
  git summary -s 2024-01-01     # 2024年1月1日以降の統計

出力例:
  過去1週間: 26コミット / 3ブランチ / 作業時間14.3h / 変更+1200 -800行

  【ユーザー別統計】
  作成者              コミット   作業時間(h)  ブランチ数    追加行    削除行
  ---------------------------------------------------------------------------------
  Daichi Toyooka           20         10.5          2       800       400
  ...
`
	fmt.Print(help)
}
