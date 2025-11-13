package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// Repository は GitHub リポジトリの情報を表す構造体
type Repository struct {
	Name       string    `json:"name"`
	IsArchived bool      `json:"isArchived"`
	Url        string    `json:"url"`
	PushedAt   time.Time `json:"pushedAt"`
}

var (
	cloneOrgArchived bool
	cloneOrgShallow  bool
	cloneOrgLimit    int
)

var cloneOrgCmd = &cobra.Command{
	Use:   "clone-org <organization>",
	Short: "組織のリポジトリをクローン",
	Long: `指定した GitHub 組織のリポジトリを一括クローンします。
リポジトリは最終更新日時（pushedAt）でソートされ、最新順にクローンされます。
すでに同じフォルダに同じ名前のリポジトリがある場合はスキップします。
リポジトリは組織名のディレクトリ配下にクローンされます。`,
	Example: `  git-plus clone-org myorg                    # myorg 組織の全リポジトリをクローン
  git-plus clone-org myorg --limit 5          # 最新5個のリポジトリのみをクローン
  git-plus clone-org myorg -n 10              # 最新10個のリポジトリのみをクローン
  git-plus clone-org myorg --archived         # アーカイブも含める
  git-plus clone-org myorg --shallow          # shallow クローンを使用
  git-plus clone-org myorg --limit 3 --shallow  # 最新3個をshallowクローン`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		org := args[0]

		fmt.Printf("組織名: %s\n", org)
		if cloneOrgArchived {
			fmt.Println("オプション: アーカイブされたリポジトリを含める")
		}
		if cloneOrgShallow {
			fmt.Println("オプション: shallow クローン (--depth=1)")
		}
		if cloneOrgLimit > 0 {
			fmt.Printf("オプション: 最新 %d 個のリポジトリのみをクローン\n", cloneOrgLimit)
		}

		// リポジトリ一覧を取得
		fmt.Println("\n[1/3] リポジトリ一覧を取得しています...")
		repos, err := getRepositories(org)
		if err != nil {
			fmt.Println("\n注意事項:")
			fmt.Println("  - GitHub CLI (gh) がインストールされている必要があります")
			fmt.Println("  - gh auth login でログイン済みである必要があります")
			fmt.Println("  - 組織名が正しいか確認してください")
			return fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
		}
		fmt.Printf("✓ %d個のリポジトリを取得しました\n", len(repos))

		// 最終更新日時でソート（最新順）
		sortReposByPushedAt(repos)

		// アーカイブされたリポジトリをフィルタリング
		filteredRepos := filterRepos(repos, cloneOrgArchived)
		if len(filteredRepos) == 0 {
			fmt.Println("\nクローンするリポジトリがありません。")
			return nil
		}

		archivedCount := len(repos) - len(filteredRepos)
		if archivedCount > 0 && !cloneOrgArchived {
			fmt.Printf("\n注意: %d個のアーカイブされたリポジトリをスキップします。\n", archivedCount)
			fmt.Println("アーカイブされたリポジトリも含める場合は --archived オプションを使用してください。")
		}

		// limit オプションが指定されている場合は上位N個のみに制限
		if cloneOrgLimit > 0 && len(filteredRepos) > cloneOrgLimit {
			fmt.Printf("\n最新 %d 個のリポジトリに制限します。\n", cloneOrgLimit)
			filteredRepos = filteredRepos[:cloneOrgLimit]
		}

		// リポジトリ数が多い場合に警告を表示
		if cloneOrgLimit == 0 && len(filteredRepos) > 50 {
			fmt.Printf("\n⚠️  警告: %d個のリポジトリをクローンします。\n", len(filteredRepos))
			fmt.Println("   多数のリポジトリをクローンする場合は時間がかかります。")
			fmt.Printf("   最新のリポジトリのみが必要な場合は --limit オプションを検討してください。\n")
			fmt.Printf("   例: git-plus clone-org %s --limit 10\n", org)
		}

		// 確認プロンプト
		fmt.Printf("\n%d個のリポジトリをクローンしますか？\n", len(filteredRepos))
		if !ui.Confirm("続行しますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// クローン先ディレクトリを作成
		fmt.Println("\n[2/3] クローン先ディレクトリを作成しています...")
		baseDir := filepath.Join(".", org)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("ディレクトリ作成に失敗しました: %w", err)
		}
		fmt.Printf("✓ ディレクトリを作成しました: %s\n", baseDir)

		// リポジトリをクローン
		fmt.Println("\n[3/3] リポジトリをクローンしています...")
		cloned, skipped := cloneRepos(filteredRepos, baseDir, cloneOrgShallow)

		// 結果を表示
		fmt.Printf("\n✓ すべての処理が完了しました！\n")
		fmt.Printf("📊 結果: %d個クローン, %d個スキップ\n", cloned, skipped)
		return nil
	},
}

func getRepositories(org string) ([]Repository, error) {
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

func sortReposByPushedAt(repos []Repository) {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].PushedAt.After(repos[j].PushedAt)
	})
}

func filterRepos(repos []Repository, includeArchived bool) []Repository {
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

func cloneRepos(repos []Repository, baseDir string, shallow bool) (int, int) {
	cloned := 0
	skipped := 0

	for i, repo := range repos {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(repos), repo.Name)

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
			errMsg := strings.TrimSpace(string(output))
			if len(errMsg) > 200 {
				errMsg = errMsg[:200] + "..."
			}
			if errMsg != "" {
				fmt.Printf("     %s\n", errMsg)
			}
			continue
		}

		fmt.Println("  ✅ 完了")
		cloned++
	}

	return cloned, skipped
}

func init() {
	cloneOrgCmd.Flags().BoolVar(&cloneOrgArchived, "archived", false, "アーカイブされたリポジトリも含める")
	cloneOrgCmd.Flags().BoolVar(&cloneOrgShallow, "shallow", false, "shallow クローンを使用（--depth=1）")
	cloneOrgCmd.Flags().IntVarP(&cloneOrgLimit, "limit", "n", 0, "最新N個のリポジトリのみをクローン")
	rootCmd.AddCommand(cloneOrgCmd)
}
