package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tonbiattack/git-plus/internal/ui"
)

// Repository は GitHub リポジトリの情報を表す構造体
type Repository struct {
	Name       string    `json:"name"`
	IsArchived bool      `json:"isArchived"`
	Url        string    `json:"url"` // HTTPS URL
	PushedAt   time.Time `json:"pushedAt"`
}

// main は GitHub 組織の全リポジトリをクローンするメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. 組織名の引数チェック
//  3. オプション（--archived, --shallow）の解析
//  4. GitHub CLI でリポジトリ一覧を取得
//  5. 組織名のディレクトリを作成
//  6. 各リポジトリをクローン（既存の場合はスキップ）
//  7. 結果を表示
//
// 使用する主なコマンド:
//  - gh repo list: リポジトリ一覧の取得
//  - git clone: リポジトリのクローン
//
// 終了コード:
//   - 0: 正常終了
//   - 1: エラー発生
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
	}

	// 組織名の引数チェック
	if len(os.Args) < 2 {
		fmt.Println("組織名を指定してください。")
		fmt.Println("使い方: git clone-org <organization> [--archived] [--shallow] [--limit N]")
		os.Exit(1)
	}

	org := os.Args[1]
	includeArchived := false
	shallow := false
	limit := 0

	// オプションの解析
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "--archived":
			includeArchived = true
		case "--shallow":
			shallow = true
		case "--limit", "-n":
			if i+1 < len(os.Args) {
				if n, err := strconv.Atoi(os.Args[i+1]); err == nil && n > 0 {
					limit = n
					i++ // 次の引数をスキップ
				} else {
					fmt.Printf("エラー: --limit の値が無効です: %s\n", os.Args[i+1])
					os.Exit(1)
				}
			} else {
				fmt.Println("エラー: --limit にはリポジトリ数を指定してください")
				os.Exit(1)
			}
		}
	}

	fmt.Printf("組織名: %s\n", org)
	if includeArchived {
		fmt.Println("オプション: アーカイブされたリポジトリを含める")
	}
	if shallow {
		fmt.Println("オプション: shallow クローン (--depth=1)")
	}
	if limit > 0 {
		fmt.Printf("オプション: 最新 %d 個のリポジトリのみをクローン\n", limit)
	}

	// リポジトリ一覧を取得
	fmt.Println("\n[1/3] リポジトリ一覧を取得しています...")
	repos, err := fetchRepositories(org)
	if err != nil {
		fmt.Printf("エラー: リポジトリ一覧の取得に失敗しました: %v\n", err)
		fmt.Println("\n注意事項:")
		fmt.Println("  - GitHub CLI (gh) がインストールされている必要があります")
		fmt.Println("  - gh auth login でログイン済みである必要があります")
		fmt.Println("  - 組織名が正しいか確認してください")
		os.Exit(1)
	}
	fmt.Printf("✓ %d個のリポジトリを取得しました\n", len(repos))

	// 最終更新日時でソート（最新順）
	sortRepositoriesByPushedAt(repos)

	// アーカイブされたリポジトリをフィルタリング
	filteredRepos := filterRepositories(repos, includeArchived)
	if len(filteredRepos) == 0 {
		fmt.Println("\nクローンするリポジトリがありません。")
		return
	}

	archivedCount := len(repos) - len(filteredRepos)
	if archivedCount > 0 && !includeArchived {
		fmt.Printf("\n注意: %d個のアーカイブされたリポジトリをスキップします。\n", archivedCount)
		fmt.Println("アーカイブされたリポジトリも含める場合は --archived オプションを使用してください。")
	}

	// limit オプションが指定されている場合は上位N個のみに制限
	if limit > 0 && len(filteredRepos) > limit {
		fmt.Printf("\n最新 %d 個のリポジトリに制限します。\n", limit)
		filteredRepos = filteredRepos[:limit]
	}

	// リポジトリ数が多い場合に警告を表示
	if limit == 0 && len(filteredRepos) > 50 {
		fmt.Printf("\n⚠️  警告: %d個のリポジトリをクローンします。\n", len(filteredRepos))
		fmt.Println("   多数のリポジトリをクローンする場合は時間がかかります。")
		fmt.Printf("   最新のリポジトリのみが必要な場合は --limit オプションを検討してください。\n")
		fmt.Printf("   例: git clone-org %s --limit 10\n", org)
	}

	// 確認プロンプト
	fmt.Printf("\n%d個のリポジトリをクローンしますか？\n", len(filteredRepos))
	if !ui.Confirm("続行しますか？", true) {
		fmt.Println("キャンセルしました。")
		os.Exit(0)
	}

	// クローン先ディレクトリを作成
	fmt.Println("\n[2/3] クローン先ディレクトリを作成しています...")
	baseDir := filepath.Join(".", org)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("エラー: ディレクトリ作成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ ディレクトリを作成しました: %s\n", baseDir)

	// リポジトリをクローン
	fmt.Println("\n[3/3] リポジトリをクローンしています...")
	cloned, skipped := cloneRepositories(filteredRepos, baseDir, shallow)

	// 結果を表示
	fmt.Printf("\n✓ すべての処理が完了しました！\n")
	fmt.Printf("📊 結果: %d個クローン, %d個スキップ\n", cloned, skipped)
}

// fetchRepositories は GitHub 組織のリポジトリ一覧を取得する
//
// パラメータ:
//   - org: 組織名
//
// 戻り値:
//   - []Repository: リポジトリ情報の配列
//   - error: gh コマンドの実行エラー
func fetchRepositories(org string) ([]Repository, error) {
	cmd := exec.Command("gh", "repo", "list", org, "--limit", "1000", "--json", "name,isArchived,url,pushedAt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, string(output))
	}

	var repos []Repository
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %v", err)
	}

	return repos, nil
}

// sortRepositoriesByPushedAt はリポジトリを最終更新日時でソートする（最新順）
//
// パラメータ:
//   - repos: リポジトリ情報の配列
func sortRepositoriesByPushedAt(repos []Repository) {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].PushedAt.After(repos[j].PushedAt)
	})
}

// filterRepositories はリポジトリをフィルタリングする
//
// パラメータ:
//   - repos: リポジトリ情報の配列
//   - includeArchived: アーカイブされたリポジトリを含めるかどうか
//
// 戻り値:
//   - []Repository: フィルタリングされたリポジトリ情報の配列
func filterRepositories(repos []Repository, includeArchived bool) []Repository {
	if includeArchived {
		return repos
	}

	var filtered []Repository
	for _, repo := range repos {
		if !repo.IsArchived {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

// cloneRepositories はリポジトリをクローンする
//
// パラメータ:
//   - repos: リポジトリ情報の配列
//   - baseDir: クローン先のベースディレクトリ
//   - shallow: shallow クローン（--depth=1）を使用するかどうか
//
// 戻り値:
//   - int: クローンしたリポジトリ数
//   - int: スキップしたリポジトリ数
func cloneRepositories(repos []Repository, baseDir string, shallow bool) (int, int) {
	cloned := 0
	skipped := 0

	for i, repo := range repos {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(repos), repo.Name)

		// アーカイブされたリポジトリの場合は表示
		archiveStatus := ""
		if repo.IsArchived {
			archiveStatus = " (アーカイブ済み)"
		}

		repoPath := filepath.Join(baseDir, repo.Name)

		// 既存のリポジトリをチェック
		if _, err := os.Stat(repoPath); err == nil {
			fmt.Printf("  ⏩ スキップ: すでに存在します%s\n", archiveStatus)
			skipped++
			continue
		}

		// クローン引数を構築
		args := []string{"clone", repo.Url, repoPath}
		if shallow {
			args = append(args, "--depth", "1")
		}

		// クローン実行
		fmt.Printf("  📥 クローン中...%s\n", archiveStatus)
		cmd := exec.Command("git", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("  ❌ 失敗: %v\n", err)
			// エラー出力が長い場合は最初の200文字のみ表示
			errMsg := strings.TrimSpace(string(output))
			if len(errMsg) > 200 {
				errMsg = errMsg[:200] + "..."
			}
			if errMsg != "" {
				fmt.Printf("     %s\n", errMsg)
			}
			continue
		}

		fmt.Printf("  ✅ 完了\n")
		cloned++
	}

	return cloned, skipped
}

// printHelp はコマンドのヘルプメッセージを表示する
func printHelp() {
	help := `git clone-org - 組織のリポジトリをクローン

使い方:
  git clone-org <organization> [オプション]

引数:
  organization          GitHub組織名

オプション:
  --archived            アーカイブされたリポジトリも含める（デフォルト: 除外）
  --shallow             shallow クローンを使用（--depth=1）
  --limit N, -n N       最新N個のリポジトリのみをクローン（デフォルト: すべて）
  -h, --help            このヘルプを表示

説明:
  指定した GitHub 組織のリポジトリを一括クローンします。
  リポジトリは最終更新日時（pushedAt）でソートされ、最新順にクローンされます。
  すでに同じフォルダに同じ名前のリポジトリがある場合はスキップします。
  リポジトリは組織名のディレクトリ配下にクローンされます。

使用例:
  git clone-org myorg                    # myorg 組織の全リポジトリをクローン
  git clone-org myorg --limit 5          # 最新5個のリポジトリのみをクローン
  git clone-org myorg -n 10              # 最新10個のリポジトリのみをクローン
  git clone-org myorg --archived         # アーカイブも含める
  git clone-org myorg --shallow          # shallow クローンを使用
  git clone-org myorg --limit 3 --shallow  # 最新3個をshallowクローン

処理フロー:
  1. GitHub CLI を使用してリポジトリ一覧を取得
  2. 最終更新日時でソート（最新順）
  3. 組織名のディレクトリを作成
  4. 各リポジトリを順次クローン
     - 既存のリポジトリはスキップ
     - アーカイブされたリポジトリは --archived オプションがない限りスキップ
     - --limit N が指定されている場合は上位N個のみをクローン
  5. 結果を表示

注意事項:
  - GitHub CLI (gh) がインストールされている必要があります
  - gh auth login でログイン済みである必要があります
  - HTTPS URLを使用するため、SSH認証の設定は不要です
  - リポジトリ数が多い場合は時間がかかることがあります

GitHub CLI のインストール:
  Windows: winget install --id GitHub.cli
  macOS:   brew install gh
  Linux:   sudo apt install gh

認証方法:
  gh auth login

  認証後、HTTPS経由でリポジトリをクローンします。
  SSH認証の設定は不要です。
`
	fmt.Print(help)
}
