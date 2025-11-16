/*
Package repo は git の拡張コマンドのうち、リポジトリ関連のコマンドを定義します。

このファイル (clone_org.go) は、GitHub 組織のリポジトリを一括クローンするコマンドを提供します。
GitHub CLI (gh) を使用して、指定された組織のすべてのリポジトリを取得し、
最終更新日時順にクローンします。

主な機能:
  - GitHub 組織のリポジトリ一覧取得
  - 最終更新日時（pushedAt）でのソート
  - アーカイブされたリポジトリのフィルタリング
  - shallow クローンのサポート
  - クローン数の制限オプション
  - すでに存在するリポジトリのスキップ

使用例:
  git clone-org myorg                    # myorg 組織の全リポジトリをクローン
  git clone-org myorg --limit 5          # 最新5個のリポジトリのみをクローン
  git clone-org myorg --archived         # アーカイブも含める
  git clone-org myorg --shallow          # shallow クローンを使用
*/
package repo

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
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// Repository は GitHub リポジトリの情報を表す構造体です。
// GitHub CLI (gh) の JSON 出力から取得したリポジトリ情報を保持します。
type Repository struct {
	Name       string    `json:"name"`       // リポジトリ名
	IsArchived bool      `json:"isArchived"` // アーカイブ済みフラグ
	Url        string    `json:"url"`        // リポジトリURL
	PushedAt   time.Time `json:"pushedAt"`   // 最終プッシュ日時
}

var (
	// cloneOrgArchived はアーカイブされたリポジトリを含めるかどうかのフラグです。
	cloneOrgArchived bool
	// cloneOrgShallow は shallow クローン（--depth=1）を使用するかどうかのフラグです。
	cloneOrgShallow bool
	// cloneOrgLimit はクローンするリポジトリの最大数を指定します（0の場合は無制限）。
	cloneOrgLimit int
)

// cloneOrgCmd は GitHub 組織のリポジトリを一括クローンするコマンドです。
// gh コマンドを使用してリポジトリ一覧を取得し、最新順にクローンします。
var cloneOrgCmd = &cobra.Command{
	Use:   "clone-org <organization>",
	Short: "組織のリポジトリをクローン",
	Long: `指定した GitHub 組織のリポジトリを一括クローンします。
リポジトリは最終更新日時（pushedAt）でソートされ、最新順にクローンされます。
すでに同じフォルダに同じ名前のリポジトリがある場合はスキップします。
リポジトリは組織名のディレクトリ配下にクローンされます。`,
	Example: `  git clone-org myorg                    # myorg 組織の全リポジトリをクローン
  git clone-org myorg --limit 5          # 最新5個のリポジトリのみをクローン
  git clone-org myorg -n 10              # 最新10個のリポジトリのみをクローン
  git clone-org myorg --archived         # アーカイブも含める
  git clone-org myorg --shallow          # shallow クローンを使用
  git clone-org myorg --limit 3 --shallow  # 最新3個をshallowクローン`,
	Args: cobra.ExactArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
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
			fmt.Printf("   例: git clone-org %s --limit 10\n", org)
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

// getRepositories は指定された組織のリポジトリ一覧を GitHub CLI (gh) を使用して取得します。
//
// パラメータ:
//   org: 組織名
//
// 戻り値:
//   - []Repository: リポジトリ情報のスライス
//   - error: エラーが発生した場合はエラーオブジェクト、成功時は nil
func getRepositories(org string) ([]Repository, error) {
	// gh repo list コマンドで組織のリポジトリ一覧を JSON 形式で取得
	ghCmd := exec.Command("gh", "repo", "list", org, "--limit", "1000", "--json", "name,isArchived,url,pushedAt")
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, string(output))
	}

	// JSON をパースして Repository 構造体のスライスに変換
	var repos []Repository
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %v", err)
	}

	return repos, nil
}

// sortReposByPushedAt はリポジトリを最終プッシュ日時（PushedAt）でソートします。
// 最新のリポジトリが先頭に来るように降順でソートします。
//
// パラメータ:
//   repos: ソート対象のリポジトリスライス（この関数はスライスを直接変更します）
func sortReposByPushedAt(repos []Repository) {
	sort.Slice(repos, func(i, j int) bool {
		// 最新のリポジトリを前に配置（降順）
		return repos[i].PushedAt.After(repos[j].PushedAt)
	})
}

// filterRepos はアーカイブ状態に基づいてリポジトリをフィルタリングします。
//
// パラメータ:
//   repos: フィルタリング対象のリポジトリスライス
//   includeArchived: アーカイブされたリポジトリを含める場合は true
//
// 戻り値:
//   フィルタリング後のリポジトリスライス
func filterRepos(repos []Repository, includeArchived bool) []Repository {
	// アーカイブを含める場合は、そのまま全リポジトリを返す
	if includeArchived {
		return repos
	}

	// アーカイブされていないリポジトリのみを抽出
	var filtered []Repository
	for _, repo := range repos {
		if !repo.IsArchived {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

// cloneRepos はリポジトリのスライスをクローンします。
//
// パラメータ:
//   repos: クローン対象のリポジトリスライス
//   baseDir: クローン先のベースディレクトリ
//   shallow: shallow クローン（--depth=1）を使用する場合は true
//
// 戻り値:
//   - int: 成功したクローン数
//   - int: スキップしたリポジトリ数
func cloneRepos(repos []Repository, baseDir string, shallow bool) (int, int) {
	cloned := 0
	skipped := 0

	// 各リポジトリを順番にクローン
	for i, repo := range repos {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(repos), repo.Name)

		// アーカイブ状態を表示するためのラベル
		archiveStatus := ""
		if repo.IsArchived {
			archiveStatus = " (アーカイブ済み)"
		}

		repoPath := filepath.Join(baseDir, repo.Name)

		// 既存のリポジトリをチェック
		// 既に同じ名前のディレクトリが存在する場合はスキップ
		if _, err := os.Stat(repoPath); err == nil {
			fmt.Printf("  ⏩ スキップ: すでに存在します%s\n", archiveStatus)
			skipped++
			continue
		}

		// クローン引数を構築
		args := []string{"clone", repo.Url, repoPath}
		if shallow {
			// shallow クローンの場合は --depth 1 を追加
			args = append(args, "--depth", "1")
		}

		// クローン実行
		fmt.Printf("  📥 クローン中...%s\n", archiveStatus)
		gitCmd := exec.Command("git", args...)
		if output, err := gitCmd.CombinedOutput(); err != nil {
			// エラー発生時はエラーメッセージを表示してスキップ
			fmt.Printf("  ❌ 失敗: %v\n", err)
			errMsg := strings.TrimSpace(string(output))
			// エラーメッセージが長すぎる場合は200文字で切り詰める
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

// init はコマンドの初期化を行います。
// フラグの定義と cloneOrgCmd を RootCmd に登録します。
func init() {
	cloneOrgCmd.Flags().BoolVarP(&cloneOrgArchived, "archived", "a", false, "アーカイブされたリポジトリも含める")
	cloneOrgCmd.Flags().BoolVarP(&cloneOrgShallow, "shallow", "s", false, "shallow クローンを使用（--depth=1）")
	cloneOrgCmd.Flags().IntVarP(&cloneOrgLimit, "limit", "n", 0, "最新N個のリポジトリのみをクローン")
	cmd.RootCmd.AddCommand(cloneOrgCmd)
}
