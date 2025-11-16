// ================================================================================
// repo_others.go
// ================================================================================
// このファイルは git の拡張コマンド repo-others コマンドを実装しています。
//
// 【概要】
// repo-others コマンドは、ローカルにクローン済みの他人のGitHubリポジトリを
// 一覧表示し、番号選択でGitHubのWebページを開く機能を提供します。
//
// 【主な機能】
// - ローカルディレクトリを走査してGitリポジトリを検出
// - 他人のリポジトリとフォークしたリポジトリを抽出
// - 最終コミット日時の降順（最新順）で表示
// - ページング機能（10件ごと）
// - 番号選択でブラウザで開く
//
// 【使用例】
//   git repo-others              # 現在のディレクトリ配下を検索
//   git repo-others --path ~/dev # 指定ディレクトリを検索
//   git repo-others --all        # 自分のリポジトリも含める
//
// 【必要な外部ツール】
// - GitHub CLI (gh): https://cli.github.com/
// ================================================================================

package repo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
)

// RepoInfo はリポジトリの情報を保持する構造体です。
type RepoInfo struct {
	Owner          string    // リポジトリのオーナー
	Name           string    // リポジトリ名
	LocalPath      string    // ローカルパス
	IsFork         bool      // フォークかどうか
	LastCommitTime time.Time // 最終コミット日時
}

var (
	repoOthersPath string
	repoOthersAll  bool
)

// repoOthersCmd は repo-others コマンドの定義です。
var repoOthersCmd = &cobra.Command{
	Use:   "repo-others",
	Short: "他人のGitHubリポジトリ一覧を表示",
	Long: `ローカルにクローン済みの他人のGitHubリポジトリを一覧表示します。
番号を入力することで、選択したリポジトリをブラウザで開くことができます。

自分のリポジトリは除外されますが、フォークしたリポジトリは含まれます。
リポジトリは最終コミット日時の降順（最新順）で表示されます。`,
	Example: `  git repo-others              # 現在のディレクトリ配下を検索
  git repo-others --path ~/dev # 指定ディレクトリを検索
  git repo-others --all        # 自分のリポジトリも含める`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// GitHub CLI の確認
		if !checkGitHubCLIInstalled() {
			return fmt.Errorf("GitHub CLI (gh) がインストールされていません\nインストール方法: https://cli.github.com/")
		}

		// 検索パスの決定
		searchPath := repoOthersPath
		if searchPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("現在のディレクトリの取得に失敗: %w", err)
			}
			searchPath = cwd
		}

		// 絶対パスに変換
		absPath, err := filepath.Abs(searchPath)
		if err != nil {
			return fmt.Errorf("パスの変換に失敗: %w", err)
		}

		fmt.Printf("検索中: %s\n", absPath)

		// リポジトリを検出
		repos, err := findGitRepositories(absPath)
		if err != nil {
			return fmt.Errorf("リポジトリの検出に失敗: %w", err)
		}

		if len(repos) == 0 {
			fmt.Println("リポジトリが見つかりませんでした。")
			return nil
		}

		fmt.Printf("見つかったリポジトリ: %d 件\n", len(repos))

		// 自分のユーザー名を取得
		myUsername, err := getMyGitHubUsername()
		if err != nil {
			return fmt.Errorf("GitHubユーザー名の取得に失敗: %w", err)
		}

		// フィルタリング
		fmt.Println("リポジトリ情報を取得中...")
		filteredRepos := []RepoInfo{}
		for _, repo := range repos {
			info, err := getRepoInfo(repo, myUsername)
			if err != nil {
				// エラーがあってもスキップして続行
				continue
			}

			// フィルタリング条件
			if repoOthersAll {
				// --all オプションが指定されている場合はすべて含める
				filteredRepos = append(filteredRepos, info)
			} else {
				// 他人のリポジトリまたはフォークのみ含める
				if info.Owner != myUsername || info.IsFork {
					filteredRepos = append(filteredRepos, info)
				}
			}
		}

		if len(filteredRepos) == 0 {
			if repoOthersAll {
				fmt.Println("リポジトリが見つかりませんでした。")
			} else {
				fmt.Println("他人のリポジトリが見つかりませんでした。")
			}
			return nil
		}

		// 最終コミット日時の降順でソート
		sort.Slice(filteredRepos, func(i, j int) bool {
			return filteredRepos[i].LastCommitTime.After(filteredRepos[j].LastCommitTime)
		})

		// ページング表示
		return displayReposWithPaging(filteredRepos)
	},
}

// checkGitHubCLIInstalled は GitHub CLI がインストールされているか確認します。
func checkGitHubCLIInstalled() bool {
	ghCmd := exec.Command("gh", "--version")
	return ghCmd.Run() == nil
}

// findGitRepositories は指定されたディレクトリ配下の Git リポジトリを検出します。
func findGitRepositories(rootPath string) ([]string, error) {
	repos := []string{}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // エラーがあってもスキップして続行
		}

		// .git ディレクトリを見つけたらその親ディレクトリをリポジトリとして記録
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repos = append(repos, repoPath)
			return filepath.SkipDir // .git 配下は探索しない
		}

		// node_modules や vendor などのディレクトリはスキップ
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		return nil
	})

	return repos, err
}

// getMyGitHubUsername は自分の GitHub ユーザー名を取得します。
func getMyGitHubUsername() (string, error) {
	ghCmd := exec.Command("gh", "api", "user", "-q", ".login")
	output, err := ghCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getRepoInfo は指定されたリポジトリの情報を取得します。
func getRepoInfo(repoPath string, myUsername string) (RepoInfo, error) {
	info := RepoInfo{
		LocalPath: repoPath,
	}

	// origin URL を取得
	gitCmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := gitCmd.Output()
	if err != nil {
		return info, fmt.Errorf("origin URL の取得に失敗: %w", err)
	}

	originURL := strings.TrimSpace(string(output))

	// owner と repo 名を抽出
	owner, name, err := parseGitHubURL(originURL)
	if err != nil {
		return info, err
	}

	info.Owner = owner
	info.Name = name

	// フォーク判定（owner が自分の場合のみ）
	if owner == myUsername {
		isFork, err := checkIfFork(owner, name)
		if err != nil {
			// エラーがあってもスキップして続行
			info.IsFork = false
		} else {
			info.IsFork = isFork
		}
	}

	// 最終コミット日時を取得
	lastCommitTime, err := getLastCommitTime(repoPath)
	if err != nil {
		// エラーがあってもスキップして続行（デフォルト値はゼロ値）
		info.LastCommitTime = time.Time{}
	} else {
		info.LastCommitTime = lastCommitTime
	}

	return info, nil
}

// parseGitHubURL は GitHub の URL から owner と repo 名を抽出します。
func parseGitHubURL(url string) (string, string, error) {
	// HTTPS の場合: https://github.com/owner/repo.git
	// SSH の場合: git@github.com:owner/repo.git

	url = strings.TrimSpace(url)

	// .git を削除
	url = strings.TrimSuffix(url, ".git")

	var ownerRepo string
	if strings.HasPrefix(url, "https://github.com/") {
		ownerRepo = strings.TrimPrefix(url, "https://github.com/")
	} else if strings.HasPrefix(url, "git@github.com:") {
		ownerRepo = strings.TrimPrefix(url, "git@github.com:")
	} else {
		return "", "", fmt.Errorf("GitHub URL ではありません: %s", url)
	}

	parts := strings.Split(ownerRepo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("無効な GitHub URL: %s", url)
	}

	return parts[0], parts[1], nil
}

// checkIfFork はリポジトリがフォークかどうかを GitHub API で確認します。
func checkIfFork(owner, repo string) (bool, error) {
	ghCmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/%s", owner, repo), "-q", ".fork")
	output, err := ghCmd.Output()
	if err != nil {
		return false, err
	}

	result := strings.TrimSpace(string(output))
	return result == "true", nil
}

// getLastCommitTime はリポジトリの最終コミット日時を取得します。
func getLastCommitTime(repoPath string) (time.Time, error) {
	gitCmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%ct")
	output, err := gitCmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timestampStr := strings.TrimSpace(string(output))
	if timestampStr == "" {
		return time.Time{}, fmt.Errorf("コミットが見つかりません")
	}

	// Unix タイムスタンプをパース
	var timestamp int64
	_, err = fmt.Sscanf(timestampStr, "%d", &timestamp)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// displayReposWithPaging はリポジトリ一覧をページング表示します。
func displayReposWithPaging(repos []RepoInfo) error {
	pageSize := 10
	currentPage := 0
	totalPages := (len(repos) + pageSize - 1) / pageSize

	for {
		// 現在のページの範囲を計算
		start := currentPage * pageSize
		end := start + pageSize
		if end > len(repos) {
			end = len(repos)
		}

		// ページを表示
		fmt.Printf("\n=== ページ %d/%d ===\n\n", currentPage+1, totalPages)
		for i := start; i < end; i++ {
			repo := repos[i]
			index := i + 1
			repoType := "clone"
			if repo.IsFork {
				repoType = "fork"
			}

			// 相対的な日時を表示
			relativeTime := formatRelativeTime(repo.LastCommitTime)
			fmt.Printf("%d) %s/%s [%s] - %s\n", index, repo.Owner, repo.Name, repoType, relativeTime)
			fmt.Printf("   %s\n\n", repo.LocalPath)
		}

		// 入力プロンプト
		fmt.Println("番号を入力してブラウザで開く | n: 次ページ | p: 前ページ | q: 終了")
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("入力の読み込みに失敗: %w", err)
		}

		input = strings.TrimSpace(input)

		// コマンド処理
		if input == "q" {
			fmt.Println("終了します。")
			return nil
		} else if input == "n" {
			if currentPage < totalPages-1 {
				currentPage++
			} else {
				fmt.Println("最後のページです。")
			}
			continue
		} else if input == "p" {
			if currentPage > 0 {
				currentPage--
			} else {
				fmt.Println("最初のページです。")
			}
			continue
		}

		// 番号選択
		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(repos) {
			fmt.Printf("無効な番号です。1から%dの範囲で入力してください。\n", len(repos))
			continue
		}

		// 選択されたリポジトリを開く
		selectedRepo := repos[selection-1]
		if err := openRepoInBrowser(selectedRepo); err != nil {
			fmt.Printf("エラー: %v\n", err)
			continue
		}

		fmt.Printf("✓ %s/%s をブラウザで開きました。\n", selectedRepo.Owner, selectedRepo.Name)
	}
}

// formatRelativeTime は時刻を相対的な表現に変換します。
func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "不明"
	}

	duration := time.Since(t)
	hours := int(duration.Hours())
	days := hours / 24

	if days > 365 {
		years := days / 365
		if years == 1 {
			return "1年前"
		}
		return fmt.Sprintf("%d年前", years)
	}
	if days > 30 {
		months := days / 30
		if months == 1 {
			return "1ヶ月前"
		}
		return fmt.Sprintf("%dヶ月前", months)
	}
	if days > 0 {
		if days == 1 {
			return "1日前"
		}
		return fmt.Sprintf("%d日前", days)
	}
	if hours > 0 {
		if hours == 1 {
			return "1時間前"
		}
		return fmt.Sprintf("%d時間前", hours)
	}
	minutes := int(duration.Minutes())
	if minutes > 0 {
		if minutes == 1 {
			return "1分前"
		}
		return fmt.Sprintf("%d分前", minutes)
	}
	return "たった今"
}

// openRepoInBrowser は指定されたリポジトリをブラウザで開きます。
func openRepoInBrowser(repo RepoInfo) error {
	repoFullName := fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
	ghCmd := exec.Command("gh", "repo", "view", repoFullName, "--web")
	ghCmd.Stdout = os.Stdout
	ghCmd.Stderr = os.Stderr
	return ghCmd.Run()
}

// init は repo-others コマンドを RootCmd に登録します。
func init() {
	cmd.RootCmd.AddCommand(repoOthersCmd)
	repoOthersCmd.Flags().StringVarP(&repoOthersPath, "path", "p", "", "検索するディレクトリ（デフォルト: カレントディレクトリ）")
	repoOthersCmd.Flags().BoolVarP(&repoOthersAll, "all", "a", false, "自分のリポジトリも含めてすべて表示")
}
