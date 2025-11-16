/*
Package repo は git の拡張コマンドのうち、リポジトリ関連のコマンドを定義します。

このファイル (batch_clone.go) は、ファイルに記載されたリポジトリURLを一括クローンするコマンドを提供します。
テキストファイルに記載されたリポジトリURLを読み込み、順次クローンします。

主な機能:
  - ファイルからリポジトリURL一覧を読み込み
  - 空行やコメント行（#で始まる行）のスキップ
  - クローン先ディレクトリの自動作成またはカスタマイズ
  - shallow クローンのサポート
  - すでに存在するリポジトリのスキップ

使用例:
  git batch-clone repos.txt                     # repos.txt のリポジトリを "repos" フォルダにクローン
  git batch-clone repos.txt --dir myprojects    # "myprojects" フォルダにクローン
  git batch-clone repos.txt -d myprojects       # 短縮オプション
  git batch-clone repos.txt --shallow           # shallow クローンを使用
  git batch-clone repos.txt -d proj -s          # カスタムフォルダ + shallow クローン

ファイルフォーマット例 (repos.txt):
  # マイプロジェクト
  https://github.com/user/repo1
  https://github.com/user/repo2

  # アーカイブ
  https://github.com/user/old-repo
*/
package repo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	// batchCloneDir はクローン先のディレクトリ名を指定します。
	// 指定がない場合は、入力ファイル名（拡張子なし）が使用されます。
	batchCloneDir string
	// batchCloneShallow は shallow クローン（--depth=1）を使用するかどうかのフラグです。
	batchCloneShallow bool
)

// batchCloneCmd はファイルに記載されたリポジトリを一括クローンするコマンドです。
var batchCloneCmd = &cobra.Command{
	Use:   "batch-clone <file>",
	Short: "ファイルに記載されたリポジトリをまとめてクローン",
	Long: `テキストファイルに記載されたリポジトリURLを読み込み、一括でクローンします。

ファイルフォーマット:
  - 1行に1つのリポジトリURL（HTTPSまたはSSH形式）
  - 空行は無視されます
  - # で始まる行はコメントとして無視されます

クローン先ディレクトリ:
  - デフォルト: ファイル名（拡張子なし）
    例: repos.txt → repos/ ディレクトリ
  - --dir オプションで任意のディレクトリ名を指定可能

既に同じ名前のリポジトリが存在する場合はスキップします。`,
	Example: `  git batch-clone repos.txt                     # "repos" フォルダにクローン
  git batch-clone repos.txt --dir myprojects    # "myprojects" フォルダにクローン
  git batch-clone repos.txt -d myprojects       # 短縮オプション
  git batch-clone repos.txt --shallow           # shallow クローンを使用
  git batch-clone repos.txt -d proj -s          # カスタムフォルダ + shallow クローン`,
	Args: cobra.ExactArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		filePath := args[0]

		// ファイルの存在確認
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("ファイルが見つかりません: %s", filePath)
		}

		// クローン先ディレクトリの決定
		targetDir := batchCloneDir
		if targetDir == "" {
			// ファイル名から拡張子を除いた名前をディレクトリ名として使用
			// 例: repos.txt → repos
			baseName := filepath.Base(filePath)
			targetDir = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		}

		fmt.Printf("入力ファイル: %s\n", filePath)
		fmt.Printf("クローン先ディレクトリ: %s\n", targetDir)
		if batchCloneShallow {
			fmt.Println("オプション: shallow クローン (--depth=1)")
		}

		// リポジトリURLを読み込み
		fmt.Println("\n[1/3] リポジトリURLを読み込んでいます...")
		urls, err := readRepositoryURLs(filePath)
		if err != nil {
			return fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
		}

		if len(urls) == 0 {
			fmt.Println("\nクローンするリポジトリがありません。")
			fmt.Println("ファイルにリポジトリURLを記載してください。")
			fmt.Println("\nファイルフォーマット例:")
			fmt.Println("  # コメント")
			fmt.Println("  https://github.com/user/repo1")
			fmt.Println("  https://github.com/user/repo2")
			return nil
		}
		fmt.Printf("✓ %d個のリポジトリURLを読み込みました\n", len(urls))

		// リポジトリURLリストを表示
		fmt.Println("\nクローン対象リポジトリ:")
		for i, url := range urls {
			repoName := extractRepoName(url)
			fmt.Printf("  %d. %s (%s)\n", i+1, repoName, url)
		}

		// 確認プロンプト
		fmt.Printf("\n%d個のリポジトリをクローンしますか？\n", len(urls))
		if !ui.Confirm("続行しますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// クローン先ディレクトリを作成
		fmt.Println("\n[2/3] クローン先ディレクトリを作成しています...")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("ディレクトリ作成に失敗しました: %w", err)
		}
		fmt.Printf("✓ ディレクトリを作成しました: %s\n", targetDir)

		// リポジトリをクローン
		fmt.Println("\n[3/3] リポジトリをクローンしています...")
		cloned, skipped, failed := cloneRepositories(urls, targetDir, batchCloneShallow)

		// 結果を表示
		fmt.Printf("\n✓ すべての処理が完了しました！\n")
		fmt.Printf("📊 結果: %d個クローン, %d個スキップ, %d個失敗\n", cloned, skipped, failed)
		return nil
	},
}

// readRepositoryURLs はファイルからリポジトリURLを読み込みます。
// 空行と#で始まるコメント行は無視されます。
//
// パラメータ:
//   filePath: 読み込むファイルのパス
//
// 戻り値:
//   - []string: リポジトリURLのスライス
//   - error: エラーが発生した場合はエラーオブジェクト、成功時は nil
func readRepositoryURLs(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var urls []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// 空行をスキップ
		if line == "" {
			continue
		}

		// コメント行をスキップ
		if strings.HasPrefix(line, "#") {
			continue
		}

		// URLの簡易的な検証
		if !strings.HasPrefix(line, "http://") &&
			!strings.HasPrefix(line, "https://") &&
			!strings.HasPrefix(line, "git@") {
			fmt.Printf("⚠️  警告 (行 %d): URLとして認識できません: %s\n", lineNumber, line)
			continue
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// extractRepoName はリポジトリURLからリポジトリ名を抽出します。
//
// パラメータ:
//   url: リポジトリURL
//
// 戻り値:
//   リポジトリ名（例: "user/repo" または "repo"）
func extractRepoName(url string) string {
	// URLから .git サフィックスを削除
	url = strings.TrimSuffix(url, ".git")

	// パスの最後の部分を取得
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		// user/repo の形式で返す
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	} else if len(parts) >= 1 {
		return parts[len(parts)-1]
	}

	return url
}

// cloneRepositories はリポジトリのURLスライスをクローンします。
//
// パラメータ:
//   urls: クローン対象のリポジトリURLスライス
//   baseDir: クローン先のベースディレクトリ
//   shallow: shallow クローン（--depth=1）を使用する場合は true
//
// 戻り値:
//   - int: 成功したクローン数
//   - int: スキップしたリポジトリ数
//   - int: 失敗したリポジトリ数
func cloneRepositories(urls []string, baseDir string, shallow bool) (int, int, int) {
	cloned := 0
	skipped := 0
	failed := 0

	for i, url := range urls {
		repoName := extractRepoName(url)
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(urls), repoName)

		// リポジトリ名（最後のパス部分）を取得してディレクトリ名として使用
		// 例: https://github.com/user/repo → repo
		urlParts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
		dirName := urlParts[len(urlParts)-1]
		repoPath := filepath.Join(baseDir, dirName)

		// 既存のリポジトリをチェック
		if _, err := os.Stat(repoPath); err == nil {
			fmt.Printf("  ⏩ スキップ: すでに存在します\n")
			skipped++
			continue
		}

		// クローン引数を構築
		args := []string{"clone", url, repoPath}
		if shallow {
			args = append(args, "--depth", "1")
		}

		// クローン実行
		fmt.Printf("  📥 クローン中...\n")
		gitCmd := exec.Command("git", args...)
		if output, err := gitCmd.CombinedOutput(); err != nil {
			// エラー発生時はエラーメッセージを表示
			fmt.Printf("  ❌ 失敗: %v\n", err)
			errMsg := strings.TrimSpace(string(output))
			// エラーメッセージが長すぎる場合は200文字で切り詰める
			if len(errMsg) > 200 {
				errMsg = errMsg[:200] + "..."
			}
			if errMsg != "" {
				fmt.Printf("     %s\n", errMsg)
			}
			failed++
			continue
		}

		fmt.Println("  ✅ 完了")
		cloned++
	}

	return cloned, skipped, failed
}

// init はコマンドの初期化を行います。
// フラグの定義と batchCloneCmd を RootCmd に登録します。
func init() {
	batchCloneCmd.Flags().StringVarP(&batchCloneDir, "dir", "d", "", "クローン先ディレクトリ名（省略時はファイル名が使用されます）")
	batchCloneCmd.Flags().BoolVarP(&batchCloneShallow, "shallow", "s", false, "shallow クローンを使用（--depth=1）")
	cmd.RootCmd.AddCommand(batchCloneCmd)
}
