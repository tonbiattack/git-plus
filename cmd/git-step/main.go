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

// AuthorStats はユーザーごとの統計情報を表す構造体
//
// フィールド:
//  - Name: ユーザー名（git の作成者名）。例: "Taro Yamada"
//  - Added: 追加した総行数
//  - Deleted: 削除した総行数
//  - Net: 純増行数（Added - Deleted）
//  - Modified: 更新行数（Added + Deleted）
//  - CurrentCode: 現在のコードベースに残っている行数（git blame ベース）
//  - Commits: コミット数
//  - AvgCommitSize: 平均コミットサイズ（Modified / Commits）
//
// 使用箇所:
//  - getAuthorStats: git log --numstat から統計を集計
//  - displayStats: 統計情報の表示
//  - saveToFile, saveToCSV: ファイル出力
type AuthorStats struct {
	Name          string
	Added         int
	Deleted       int
	Net           int     // 純増行数 (Added - Deleted)
	Modified      int     // 更新行数 (Added + Deleted)
	CurrentCode   int     // 現在のコードベースに残っている行数
	Commits       int     // コミット数
	AvgCommitSize float64 // 平均コミットサイズ (Modified / Commits)
}

// main はリポジトリのステップ数とユーザーごとの貢献度を表示するメイン処理
//
// 処理フロー:
//  1. コマンドライン引数を解析（期間、スコープなど）
//  2. 期間の計算（-w/-m/-y オプションから --since を計算）
//  3. 全ブランチのコミット履歴から作成者ごとの統計を取得
//  4. 現在のコードベースでの各ユーザーの行数を取得（git blame）
//  5. 統計情報を計算（追加/削除/更新行数、コミット数、平均コミットサイズなど）
//  6. 結果を表示（コンソール、テキストファイル、CSVファイル）
//
// 使用するgitコマンド:
//  - git log --all --pretty=format:%H%x09%an --numstat: コミット履歴と変更行数を取得
//  - git rev-list --max-parents=0 HEAD: 初回コミットのハッシュを取得
//  - git ls-files: 現在のファイル一覧を取得
//  - git blame --line-porcelain <file>: 各行の作成者を取得
//
// 実装の詳細:
//  - デフォルトで初回コミットを除外（--include-initial で含める）
//  - 期間指定がある場合、その期間内のコミットのみを対象とする
//  - 現在のコードベースでの行数は git blame で正確に計算
//  - 結果はテキストファイルとCSVファイルの両方に保存
//
// 終了コード:
//  - 0: 正常終了
//  - 1: エラー発生（コミット履歴取得失敗など）
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
	totalModified := 0
	totalCommits := 0
	for _, stat := range authorStats {
		totalAdded += stat.Added
		totalDeleted += stat.Deleted
		totalNet += stat.Net
		totalModified += stat.Modified
		totalCommits += stat.Commits
	}

	// 現在のリポジトリの総行数を取得
	currentLines := getCurrentTotalLines()

	// 現在のコードベースでの各ユーザーの行数を取得（git blameベース）
	// 期間指定時は、その期間内のコミットのみを対象とする
	currentCodeStats := getCurrentCodeByAuthor(sinceArg, untilArg)

	// authorStatsに現在のコード行数を追加し、派生指標を計算
	for i := range authorStats {
		if lines, exists := currentCodeStats[authorStats[i].Name]; exists {
			authorStats[i].CurrentCode = lines
		}

		// 平均コミットサイズを計算
		if authorStats[i].Commits > 0 {
			authorStats[i].AvgCommitSize = float64(authorStats[i].Modified) / float64(authorStats[i].Commits)
		}
	}

	// コード割合が多い順にソート（現在のコードベース）
	sort.Slice(authorStats, func(i, j int) bool {
		return authorStats[i].CurrentCode > authorStats[j].CurrentCode
	})

	// 結果を表示
	displayStats(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, untilArg)

	// ファイルに保存
	saveToFile(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, untilArg)

	// CSVファイルに保存
	saveToCSV(authorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines, sinceArg, untilArg)
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - 全てのコマンドラインオプションの説明
//  - 期間指定の方法（-w/-m/-y、--since/--until）
//  - 統計情報の内容と計算方法
//  - 出力ファイルの形式（テキスト、CSV）
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
  結果はステップ数が多い順に表示され、自動的にテキストファイルとCSVファイルに保存されます。
`
	fmt.Print(help)
}

// getAuthorStats は作成者ごとの統計情報を取得する
//
// パラメータ:
//  - since: 集計開始日時。例: "2024-01-01", "1 week ago"
//  - until: 集計終了日時。空文字列の場合は現在まで
//  - excludeInitial: true の場合、初回コミットを除外
//
// 戻り値:
//  - []AuthorStats: 作成者ごとの統計情報のスライス
//
// 使用するgitコマンド:
//  - git rev-list --max-parents=0 HEAD: 初回コミットのハッシュを取得
//  - git log --all --pretty=format:%H%x09%an --numstat: コミット履歴と変更行数を取得
//
// 実装の詳細:
//  - git log --numstat の出力形式:
//    - コミット行: <ハッシュ>\t<作成者>
//    - numstat行: <追加>\t<削除>\t<ファイル名>
//  - バイナリファイルは "-" と表示されるのでスキップ
//  - 同じコミットを重複カウントしないように管理
//  - 初回コミットはデフォルトで除外（大量の行数が追加されることが多いため）
func getAuthorStats(since, until string, excludeInitial bool) []AuthorStats {
	// 初回コミットのハッシュを取得（除外する場合）
	var initialCommitHash string
	if excludeInitial {
		output, err := gitcmd.Run("rev-list", "--max-parents=0", "HEAD")
		if err == nil {
			initialCommitHash = strings.TrimSpace(string(output))
		}
	}

	// 作成者ごとに集計
	authorMap := make(map[string]*AuthorStats)
	commitCountMap := make(map[string]int) // コミット数を別途カウント

	// git log --numstat の出力を解析
	// コミットごとに作成者を特定する必要がある
	cmdArgs := []string{"log", "--all", "--pretty=format:%H%x09%an", "--numstat"}
	if since != "" {
		cmdArgs = append(cmdArgs, "--since="+since)
	}
	if until != "" {
		cmdArgs = append(cmdArgs, "--until="+until)
	}

	output, err := gitcmd.Run(cmdArgs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git log の実行に失敗しました: %v\n", err)
		return nil
	}

	lines := strings.Split(string(output), "\n")
	currentAuthor := ""
	currentHash := ""
	processedCommits := make(map[string]bool) // 同じコミットを重複カウントしないため

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

				// コミット数をカウント（重複を避ける）
				commitKey := currentAuthor + ":" + currentHash
				if !processedCommits[commitKey] {
					commitCountMap[currentAuthor]++
					processedCommits[commitKey] = true
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
					authorMap[currentAuthor].Modified += (added + deleted)
				}
			}
		}
	}

	// マップをスライスに変換し、コミット数を設定
	stats := make([]AuthorStats, 0, len(authorMap))
	for _, stat := range authorMap {
		stat.Commits = commitCountMap[stat.Name]
		stats = append(stats, *stat)
	}

	return stats
}

// getCurrentTotalLines は現在のリポジトリの総行数を取得する
//
// 戻り値:
//  - int: 総行数（すべてのトラッキング対象ファイルの合計）
//
// 使用するgitコマンド:
//  - git ls-files: Gitで管理されているファイルの一覧を取得
//
// 実装の詳細:
//  - git ls-files で取得したすべてのファイルを読み込み
//  - 各ファイルの行数をカウント
//  - 末尾が改行で終わるファイルは空行を除外
//  - バイナリファイルや読み込みエラーは無視
func getCurrentTotalLines() int {
	// git ls-files | xargs cat | wc -l
	// Windows環境を考慮して、Goで実装
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
		// 末尾が改行で終わっているファイルは空行が追加されるので除外
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		totalLines += len(lines)
	}

	return totalLines
}

// getCurrentCodeByAuthor は現在のコードベースでの各ユーザーの行数を取得する
//
// パラメータ:
//  - since: 集計開始日時。空文字列の場合は全期間
//  - until: 集計終了日時。空文字列の場合は現在まで
//
// 戻り値:
//  - map[string]int: ユーザー名をキー、行数を値とするマップ
//
// 使用するgitコマンド:
//  - git ls-files: Gitで管理されているファイルの一覧を取得
//  - git blame --line-porcelain <file>: 各行の作成者情報を取得
//  - git log --all --pretty=format:%H: 期間内のコミットハッシュを取得
//
// 実装の詳細:
//  - git blame を使用して各行の最終更新者を特定
//  - 期間指定がある場合、その期間内のコミットのみを対象とする
//  - --line-porcelain フォーマット:
//    - <ハッシュ> <元行番号> <最終行番号> <行数>
//    - author <作成者名>
//  - 期間指定がない場合は全ての行をカウント
func getCurrentCodeByAuthor(since, until string) map[string]int {
	// git ls-filesで全ファイルを取得
	output, err := gitcmd.Run("ls-files")
	if err != nil {
		return make(map[string]int)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	authorLines := make(map[string]int)

	// 期間指定がある場合、その期間内のコミットハッシュを取得
	var validCommits map[string]bool
	if since != "" || until != "" {
		validCommits = getCommitsInPeriod(since, until)
	}

	for _, file := range files {
		if file == "" {
			continue
		}

		// git blameで各行の作成者を取得
		blameOutput, err := gitcmd.Run("blame", "--line-porcelain", file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(blameOutput), "\n")
		currentCommit := ""
		for _, line := range lines {
			// コミットハッシュ行（40文字の16進数で始まる行）
			// git blame --line-porcelainの最初の行は "<hash> <orig-line> <final-line> <num-lines>" の形式
			if len(line) >= 40 {
				fields := strings.Fields(line)
				if len(fields) > 0 && len(fields[0]) == 40 {
					// 最初のフィールドがフルハッシュ(40文字)の場合
					currentCommit = fields[0]
				}
			}

			// "author " で始まる行から作成者名を取得
			if strings.HasPrefix(line, "author ") {
				author := strings.TrimPrefix(line, "author ")
				
				// 期間指定がある場合は、その期間内のコミットのみカウント
				if validCommits != nil {
					if validCommits[currentCommit] {
						authorLines[author]++
					}
				} else {
					// 期間指定がない場合は全てカウント
					authorLines[author]++
				}
			}
		}
	}

	return authorLines
}

// getCommitsInPeriod は指定期間内のコミットハッシュのセットを取得する
//
// パラメータ:
//  - since: 集計開始日時。空文字列の場合は制限なし
//  - until: 集計終了日時。空文字列の場合は現在まで
//
// 戻り値:
//  - map[string]bool: コミットハッシュ（40文字）をキーとするセット
//
// 使用するgitコマンド:
//  - git log --all --pretty=format:%H: 全コミットのハッシュを取得
//
// 実装の詳細:
//  - since/until オプションで期間を絞り込み
//  - フルハッシュ（40文字）のみを有効とする
//  - map[string]bool をセットとして使用
func getCommitsInPeriod(since, until string) map[string]bool {
	// 期間内の全コミットハッシュを取得
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

// displayStats は統計情報をコンソールに表示する
//
// パラメータ:
//  - stats: ユーザーごとの統計情報のスライス（ソート済み）
//  - totalAdded: 全体の追加行数
//  - totalDeleted: 全体の削除行数
//  - totalNet: 全体の純増行数
//  - totalModified: 全体の更新行数
//  - totalCommits: 全体のコミット数
//  - currentLines: 現在のリポジトリ総行数
//  - since: 集計開始日時（空文字列の場合は全期間）
//  - until: 集計終了日時（空文字列の場合は現在まで）
//
// 実装の詳細:
//  - 全体統計とユーザー別統計を表示
//  - 各種比率（追加比、削除比、更新比、コード割合）を計算
//  - 数値は3桁区切りで表示（formatNumber関数を使用）
//  - コード割合順（降順）で表示
func displayStats(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
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
	fmt.Printf("  更新行数: %s\n", formatNumber(totalModified))
	fmt.Printf("  総コミット数: %s\n", formatNumber(totalCommits))
	fmt.Println()

	// ユーザー別統計
	fmt.Println("【ユーザー別統計】（コード割合が多い順）")
	fmt.Println()
	fmt.Printf("%-30s %10s %10s %10s %10s %8s %10s %8s %8s %8s %10s\n",
		"作成者", "追加", "削除", "更新", "現在", "コミ数", "平均", "追加比", "削除比", "更新比", "コード割合")
	fmt.Println(strings.Repeat("-", 138))

	// 期間指定がある場合は、期間内のコード行数の合計を使用
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
			formatNumber(stat.Added),
			formatNumber(stat.Deleted),
			formatNumber(stat.Modified),
			formatNumber(stat.CurrentCode),
			formatNumber(stat.Commits),
			stat.AvgCommitSize,
			addedRatio,
			deletedRatio,
			modifiedRatio,
			codeRatio,
		)
	}

	fmt.Println()
}

// saveToFile は統計情報をテキストファイルに保存する
//
// パラメータ:
//  - stats: ユーザーごとの統計情報のスライス
//  - totalAdded: 全体の追加行数
//  - totalDeleted: 全体の削除行数
//  - totalNet: 全体の純増行数
//  - totalModified: 全体の更新行数
//  - totalCommits: 全体のコミット数
//  - currentLines: 現在のリポジトリ総行数
//  - since: 集計開始日時
//  - until: 集計終了日時
//
// 実装の詳細:
//  - ファイル名: "git_step_<期間>_<日付>.txt"
//  - displayStats と同じ内容を出力
//  - ファイル作成エラーは標準エラー出力に表示
func saveToFile(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
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
	fmt.Fprintf(file, "  更新行数: %s\n", formatNumber(totalModified))
	fmt.Fprintf(file, "  総コミット数: %s\n", formatNumber(totalCommits))
	fmt.Fprintln(file)

	fmt.Fprintln(file, "【ユーザー別統計】（コード割合が多い順）")
	fmt.Fprintln(file)
	fmt.Fprintf(file, "%-30s %10s %10s %10s %10s %8s %10s %8s %8s %8s %10s\n",
		"作成者", "追加", "削除", "更新", "現在", "コミ数", "平均", "追加比", "削除比", "更新比", "コード割合")
	fmt.Fprintln(file, strings.Repeat("-", 138))

	// 期間指定がある場合は、期間内のコード行数の合計を使用
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
			formatNumber(stat.Added),
			formatNumber(stat.Deleted),
			formatNumber(stat.Modified),
			formatNumber(stat.CurrentCode),
			formatNumber(stat.Commits),
			stat.AvgCommitSize,
			addedRatio,
			deletedRatio,
			modifiedRatio,
			codeRatio,
		)
	}

	fmt.Printf("\n結果を %s に保存しました。\n", filename)
}

// saveToCSV は統計情報をCSVファイルに保存する
//
// パラメータ:
//  - stats: ユーザーごとの統計情報のスライス
//  - totalAdded: 全体の追加行数
//  - totalDeleted: 全体の削除行数
//  - totalNet: 全体の純増行数
//  - totalModified: 全体の更新行数
//  - totalCommits: 全体のコミット数
//  - currentLines: 現在のリポジトリ総行数
//  - since: 集計開始日時
//  - until: 集計終了日時
//
// 実装の詳細:
//  - ファイル名: "git_step_<期間>_<日付>.csv"
//  - ヘッダー行付きのCSV形式
//  - 数値はカンマ区切りなし（Excel等で読み込みやすくするため）
//  - 小数点は2桁まで表示
func saveToCSV(stats []AuthorStats, totalAdded, totalDeleted, totalNet, totalModified, totalCommits, currentLines int, since, until string) {
	// ファイル名を生成
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

	// CSVヘッダー
	fmt.Fprintln(file, "作成者,追加行数,削除行数,純増行数,更新行数,現在行数,コミット数,平均コミットサイズ,追加比率(%),削除比率(%),更新比率(%),コード割合(%)")

	// 期間指定がある場合は、期間内のコード行数の合計を使用
	totalCurrentCode := currentLines
	if since != "" || until != "" {
		totalCurrentCode = 0
		for _, stat := range stats {
			totalCurrentCode += stat.CurrentCode
		}
	}

	// データ行
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

// formatNumber は整数を3桁区切りの文字列に変換する
//
// パラメータ:
//  - n: 変換する整数（負数も対応）
//
// 戻り値:
//  - string: 3桁区切りの文字列。例: 1234567 -> "1,234,567"
//
// 実装の詳細:
//  - 負数の場合は符号を保持
//  - 右から3桁ごとにカンマを挿入
//  - Go標準ライブラリを使わず手動実装
func formatNumber(n int) string {
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
