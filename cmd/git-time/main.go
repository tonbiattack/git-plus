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

// CommitInfo はコミット情報を表す構造体
// git log から取得した各コミットのメタデータを保持
type CommitInfo struct {
	Hash      string    // コミットハッシュ（SHA-1）
	Branch    string    // コミットが所属するブランチ名
	Author    string    // コミット作成者名
	Timestamp time.Time // コミット日時
	Message   string    // コミットメッセージ
	Files     []string  // 変更されたファイルのリスト（git diff-tree から取得）
}

// BranchStat はブランチごとの集計情報を表す構造体
type BranchStat struct {
	Name         string         // ブランチ名
	CommitCount  int            // ブランチに含まれるコミット数
	TotalHours   float64        // ブランチ全体の総作業時間（時間単位）
	AuthorCounts map[string]int // 作成者ごとのコミット数（割合算出用）
}

// CommitStat はコミットごとの集計情報を表す構造体
type CommitStat struct {
	Message   string    // コミットメッセージ
	Branch    string    // コミットが所属するブランチ名
	Author    string    // コミット作成者名
	Hours     float64   // 推定作業時間（時間単位）
	Timestamp time.Time // コミット日時
	Files     []string  // 変更されたファイルのリスト
}

func main() {
	// コマンドライン引数を解析
	// サポートするオプション: --since/-s, --until/-u, --commits/-c, -w, -m, -y, --scope, --limit, --help/-h
	sinceArg := ""       // 集計開始日時（例: "2024-01-01", "2 weeks ago"）
	untilArg := ""       // 集計終了日時（デフォルト: 現在）
	showCommits := false // true: コミット別表示、false: ブランチ別表示
	weeks := 0           // -w オプションで指定された週数
	months := 0          // -m オプションで指定された月数
	years := 0           // -y オプションで指定された年数
	scope := "remotes"   // 対象スコープ: "remotes", "all", "local"
	commitLimit := 20    // コミット表示件数の上限（-c使用時）

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
		case "--scope":
			if i+1 < len(args) {
				s := args[i+1]
				if s == "remotes" || s == "all" || s == "local" {
					scope = s
				}
				i++
			}
		case "--limit", "-l":
			if i+1 < len(args) {
				l, err := strconv.Atoi(args[i+1])
				if err == nil && l > 0 {
					commitLimit = l
				}
				i++
			}
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		}
	}

	// -w, -m, -y オプションから期間を計算
	// 優先順位: weeks > months > years（最初に設定されたものが優先）
	if weeks > 0 {
		sinceArg = fmt.Sprintf("%d weeks ago", weeks)
	} else if months > 0 {
		sinceArg = fmt.Sprintf("%d months ago", months)
	} else if years > 0 {
		sinceArg = fmt.Sprintf("%d years ago", years)
	}

	// デフォルトは1週間前から（オプション未指定時）
	if sinceArg == "" {
		sinceArg = "1 week ago"
	}

	fmt.Printf("コミット履歴を分析しています（%s から", sinceArg)
	if untilArg != "" {
		fmt.Printf(" %s まで", untilArg)
	}
	fmt.Println("）...")

	// コミット履歴を取得
	commits, err := getCommits(sinceArg, untilArg, scope)
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
		displayCommitStats(commits, outputFile, commitLimit)
	} else {
		// ブランチごとの集計
		displayBranchStats(commits, outputFile)
	}
}

// generateOutputFileName は出力ファイル名を生成する
// フォーマット: git_time_{期間}_{実行日}.txt
// 例: git_time_1_weeks_ago_2025-11-08.txt
func generateOutputFileName(since, until string) string {
	now := time.Now()
	dateStr := now.Format("2006-01-02") // YYYY-MM-DD形式

	// 期間をファイル名に含める（スペースとコロンを変換）
	periodStr := strings.ReplaceAll(since, " ", "_")
	periodStr = strings.ReplaceAll(periodStr, ":", "-")

	return fmt.Sprintf("git_time_%s_%s.txt", periodStr, dateStr)
}

// getCommits は指定された期間のコミット履歴を取得する
// scope: "remotes" (リモート追跡ブランチのみ), "all" (すべてのブランチ), "local" (ローカルブランチのみ)
// 注意: scope="remotes"の場合、最新の結果を得るには事前に git fetch を実行してください。
func getCommits(since, until, scope string) ([]CommitInfo, error) {
	args := []string{"log", "--format=%H|%D|%an|%at|%s"}

	// スコープに応じてオプションを追加
	switch scope {
	case "remotes":
		args = append(args, "--remotes") // リモート追跡ブランチのみ
	case "all":
		args = append(args, "--all") // すべてのブランチ（ローカル+リモート）
	case "local":
		args = append(args, "--branches") // ローカルブランチのみ
	default:
		args = append(args, "--remotes") // デフォルトはremotes
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

		// stashとdetachedは除外（develop、main、その他すべてのブランチは表示される）
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
// git diff-tree を使用してファイル名のみを取得（差分内容は取得しない）
// マージコミットの場合も含め、すべての変更ファイルを返す
func getChangedFiles(hash string) []string {
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-only", "-r", hash)
	output, err := cmd.Output()
	if err != nil {
		return []string{} // エラー時は空配列を返す
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
// 優先順位: 1) HEAD -> branch (現在チェックアウト中)
//  2. ローカルブランチ
//  3. リモートブランチ（origin/xxx -> xxx に変換）
//  4. "unknown"
//
// 特殊な値: "stash" (stashエントリ), "detached" (detached HEAD)
func extractBranch(refs, hash string) string {
	if refs == "" {
		// refsがない場合は、コミットが属するブランチを取得
		branch := getBranchForCommit(hash)
		if branch != "" {
			return branch
		}
		return "detached" // どのブランチにも属さない場合
	}

	// "HEAD -> main, origin/main" のような形式から解析
	parts := strings.Split(refs, ",")

	// stashを検出（完全一致）
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "refs/stash" || strings.HasPrefix(part, "refs/stash@") {
			return "stash"
		}
	}

	// フェーズ1: "HEAD -> branch" 形式を優先
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

	// フェーズ2: ローカルブランチが見つからない場合、最初のリモートブランチを使用
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "/") && !strings.HasPrefix(part, "tag:") {
			// "origin/main" -> "main" に変換
			parts := strings.SplitN(part, "/", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}

	return "unknown"
}

// getBranchForCommit はコミットハッシュから所属ブランチを取得
// git branch --contains を使用して、コミットを含むブランチの最初のものを返す
// 複数のブランチに含まれる場合は、git が返す最初のブランチを使用
func getBranchForCommit(hash string) string {
	cmd := exec.Command("git", "branch", "--contains", hash, "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return "" // エラー時は空文字列を返す
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 && lines[0] != "" {
		return lines[0] // 最初のブランチを返す
	}

	return ""
}

// displayBranchStats はブランチごとの統計を表示する
// 表示内容:
//   - ブランチ名
//   - コミット数と総作業時間(時間単位、小数点第1位まで)
//   - 貢献者一覧(名前: コミット数 (割合%))
//
// 表示順: 最新コミットの新しい順(降順) - branchLastCommitマップで判定
// 除外ルール: "stash"と"detached"を完全一致で除外
// 表示件数: 無制限(全ブランチ表示)
// 割合計算: (作成者のコミット数 ÷ ブランチ全体のコミット数) × 100
// 作業時間計算: ブランチ内の同一作成者による連続コミットの時間差が2時間以内なら実時間、超えたら30分と見積もる
func displayBranchStats(commits []CommitInfo, outputFile string) {
	// ブランチごと・作成者ごとにコミットをグループ化
	type branchAuthorKey struct {
		branch string
		author string
	}
	branchAuthorCommits := make(map[branchAuthorKey][]CommitInfo)
	branchMap := make(map[string]*BranchStat)

	// まず、ブランチ×作成者の組み合わせでコミットをグループ化
	for _, commit := range commits {
		key := branchAuthorKey{branch: commit.Branch, author: commit.Author}
		branchAuthorCommits[key] = append(branchAuthorCommits[key], commit)

		// BranchStatの初期化
		if _, exists := branchMap[commit.Branch]; !exists {
			branchMap[commit.Branch] = &BranchStat{
				Name:         commit.Branch,
				CommitCount:  0,
				TotalHours:   0,
				AuthorCounts: make(map[string]int),
			}
		}
		branchMap[commit.Branch].CommitCount++
		branchMap[commit.Branch].AuthorCounts[commit.Author]++
	}

	// ブランチ×作成者ごとに作業時間を計算
	for key, authorCommits := range branchAuthorCommits {
		// 同一ブランチ・同一作成者のコミットを時系列順にソート
		sort.Slice(authorCommits, func(i, j int) bool {
			return authorCommits[i].Timestamp.Before(authorCommits[j].Timestamp)
		})

		// この作成者のこのブランチでの作業時間を計算
		for i := 0; i < len(authorCommits); i++ {
			commit := authorCommits[i]
			hours := 0.5 // デフォルト30分

			// 次のコミット（同じブランチ・同じ作成者）との時間差を計算
			if i+1 < len(authorCommits) {
				nextCommit := authorCommits[i+1]
				duration := nextCommit.Timestamp.Sub(commit.Timestamp)
				h := duration.Hours()

				// 2時間以内の差であれば作業時間とみなす
				if h <= 2.0 && h >= 0 {
					hours = h
				}
			}

			branchMap[key.branch].TotalHours += hours
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
	output.WriteString(strings.Repeat("-", 80) + "\n")

	totalCommits := 0
	totalHours := 0.0

	// すべてのブランチを表示（表示件数制限なし）
	for _, stat := range stats {
		// "stash"と"detached"を完全一致で除外
		if stat.Name == "stash" || stat.Name == "detached" {
			continue
		}

		// ブランチ名、コミット数、作業時間
		line := fmt.Sprintf("%-30s %3d commits (約%.1fh)\n",
			stat.Name, stat.CommitCount, stat.TotalHours)
		output.WriteString(line)

		// 作成者ごとのコミット数と割合を表示
		if len(stat.AuthorCounts) > 0 {
			// 作成者をコミット数でソート（多い順）
			type authorStat struct {
				name  string
				count int
			}
			authors := make([]authorStat, 0, len(stat.AuthorCounts))
			for author, count := range stat.AuthorCounts {
				authors = append(authors, authorStat{name: author, count: count})
			}
			sort.Slice(authors, func(i, j int) bool {
				return authors[i].count > authors[j].count
			})

			// 作成者情報を表示
			for _, author := range authors {
				// 割合の算出: (作成者のコミット数 ÷ ブランチ全体のコミット数) × 100
				// 例: ブランチに10コミットあり、Aさんが7コミット → 7 ÷ 10 × 100 = 70.0%
				percentage := float64(author.count) / float64(stat.CommitCount) * 100
				output.WriteString(fmt.Sprintf("  └─ %-25s %3d commits (%.1f%%)\n",
					author.name, author.count, percentage))
			}
		}

		totalCommits += stat.CommitCount
		totalHours += stat.TotalHours
	}

	output.WriteString(strings.Repeat("-", 80) + "\n")
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
// 表示内容:
//   - 作業時間(時間単位、小数点第1位まで)
//   - ブランチ名
//   - コミットメッセージ
//   - コミット日時(YYYY-MM-DD HH:MM形式)
//   - 作成者名
//   - 変更ファイル名(最大3件、超える場合は件数表示)
//
// 表示順: コミット日時の新しい順(降順)
// 表示件数: limit パラメータで指定（デフォルト20件）
// 作業時間計算: 同じブランチ・同じ作成者の次のコミットとの時間差を使用
// ファイル出力: git_time_*_commits.txt に保存
func displayCommitStats(commits []CommitInfo, outputFile string, limit int) {
	// ブランチ×作成者ごとにコミットをグループ化
	type branchAuthorKey struct {
		branch string
		author string
	}
	branchAuthorCommits := make(map[branchAuthorKey][]CommitInfo)

	for _, commit := range commits {
		key := branchAuthorKey{branch: commit.Branch, author: commit.Author}
		branchAuthorCommits[key] = append(branchAuthorCommits[key], commit)
	}

	// 各グループ内でコミットを時系列順にソート
	for key := range branchAuthorCommits {
		sort.Slice(branchAuthorCommits[key], func(i, j int) bool {
			return branchAuthorCommits[key][i].Timestamp.Before(branchAuthorCommits[key][j].Timestamp)
		})
	}

	// コミットごとに作業時間を計算
	commitStats := make([]CommitStat, 0, len(commits))
	commitHoursMap := make(map[string]float64) // Hash -> Hours のマップ

	for _, authorCommits := range branchAuthorCommits {
		for i := 0; i < len(authorCommits); i++ {
			commit := authorCommits[i]
			hours := 0.5 // デフォルト30分

			// 次のコミット（同じブランチ・同じ作成者）との時間差を計算
			if i+1 < len(authorCommits) {
				nextCommit := authorCommits[i+1]
				duration := nextCommit.Timestamp.Sub(commit.Timestamp)
				h := duration.Hours()

				// 2時間以内の差であれば作業時間とみなす
				if h <= 2.0 && h >= 0 {
					hours = h
				}
			}

			commitHoursMap[commit.Hash] = hours
		}
	}

	// 元のコミットリストをもとにCommitStatを作成
	for _, commit := range commits {
		hours, ok := commitHoursMap[commit.Hash]
		if !ok {
			hours = 0.5 // デフォルト
		}

		commitStats = append(commitStats, CommitStat{
			Message:   commit.Message,
			Branch:    commit.Branch,
			Author:    commit.Author,
			Hours:     hours,
			Timestamp: commit.Timestamp,
			Files:     commit.Files,
		})
	}

	// コミット時刻順にソート（新しい順）
	sort.Slice(commitStats, func(i, j int) bool {
		return commitStats[i].Timestamp.After(commitStats[j].Timestamp)
	})

	// 出力ファイル名を生成（commits用 - 元のファイル名の拡張子前に _commits を挿入）
	commitsOutputFile := strings.Replace(outputFile, ".txt", "_commits.txt", 1)

	// 出力内容を構築
	var output strings.Builder
	output.WriteString("コミットごとの作業時間（直近順）:\n")
	output.WriteString(strings.Repeat("-", 80) + "\n")

	// 表示件数の上限を設定（0の場合は全件表示）
	displayCount := limit
	if limit == 0 || len(commitStats) < displayCount {
		displayCount = len(commitStats)
	}

	// 上位displayCount件のみ表示
	for i := 0; i < displayCount; i++ {
		stat := commitStats[i]
		dateStr := stat.Timestamp.Format("2006-01-02 15:04")

		// ファイル名のリストを整形（最大3件、超える場合は件数表示）
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

		// ブランチ、作成者、ファイル名を分けて表示
		line := fmt.Sprintf("[%.1fh] %-20s %s (%s)\n       作成者: %s\n       ファイル: %s\n",
			stat.Hours, stat.Branch, stat.Message, dateStr, stat.Author, filesStr)
		output.WriteString(line)
	}

	// 表示しきれなかったコミット数を表示
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
// 全てのコマンドラインオプションの使い方を説明
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
                        注: -w/-m/-y を指定すると --since は上書きされます
  
  --until, -u <日時>    集計終了日時 (デフォルト: 現在)
                        例: "yesterday", "2024-12-31"
  
  --scope <範囲>        対象範囲を指定 (デフォルト: remotes)
                        remotes: リモート追跡ブランチのみ (再現性高)
                        all: すべてのブランチ (ローカル+リモート)
                        local: ローカルブランチのみ
  
  --commits, -c         ブランチ別ではなくコミット別に表示
  
  --limit, -l <数>      コミット別表示時の表示件数 (デフォルト: 20)
                        例: -l 50 (50件表示), -l 0 (全件表示)
  
  --help, -h            このヘルプを表示

使用例:
  git time                              # 過去1週間の作業時間をブランチ別に表示
  git time -w 1                         # 過去1週間の作業時間
  git time -m 1                         # 過去1ヶ月の作業時間
  git time -y 1                         # 過去1年の作業時間
  git time -w 2 -c                      # 過去2週間をコミット別に表示 (20件)
  git time -w 2 -c -l 50                # 過去2週間をコミット別に表示 (50件)
  git time -w 2 -c -l 0                 # 過去2週間をコミット別に全件表示
  git time --scope all                  # すべてのブランチ (ローカル作業含む)
  git time --scope local                # ローカルブランチのみ
  git time --since "2024-01-01" --until "2024-12-31"  # 期間指定

作業時間の計算方法:
  - 同じブランチ・同じ作成者の連続するコミット間の時間差が2時間以内: その時間を作業時間とする
  - 2時間を超える場合: デフォルトで30分と見積もる
  - 最後のコミット: デフォルトで30分
  ※ ブランチや作成者が異なるコミット間の時間は計算に含まれません

出力:
  - コンソールに結果を表示
  - 自動的にファイル(git_time_*.txt)に結果を保存

注意:
  - --scope remotes の場合、事前に git fetch を実行して最新情報を取得してください
  - stash と detached HEAD は除外されます
`)
}
