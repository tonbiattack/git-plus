/*
Package repo は git の拡張コマンドのうち、リポジトリ関連のコマンドを定義します。

このファイル (create_repository.go) は、GitHub リポジトリの作成から
クローン、VSCode 起動までを一括で行うコマンドを提供します。

主な機能:
  - GitHub リポジトリの作成（public/private 選択可能）
  - リポジトリの説明文の設定
  - 作成したリポジトリのクローン
  - VSCode の自動起動

使用例:
  git create-repository my-new-project  # 対話的にリポジトリを作成
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

// createRepositoryCmd は GitHub リポジトリの作成から VSCode 起動までを
// 一括で実行するコマンドです。
var createRepositoryCmd = &cobra.Command{
	Use:   "create-repository <リポジトリ名>",
	Short: "GitHubリポジトリの作成からクローン、VSCode起動まで",
	Long: `以下の処理を自動的に実行します:
  1. GitHubにリモートリポジトリを作成（public/private選択可能、Description指定可能）
  2. 作成したリポジトリをクローン
  3. クローンしたディレクトリに移動
  4. VSCodeでプロジェクトを開く`,
	Example: `  git create-repository my-new-project`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]
		fmt.Printf("リポジトリ名: %s\n", repoName)

		// 公開設定の確認（public/private）
		visibility := promptForVisibility()
		fmt.Printf("公開設定: %s\n", visibility)

		// 説明の確認
		description := promptForDescription()
		if description != "" {
			fmt.Printf("説明: %s\n", description)
		}

		// 確認プロンプト
		if !ui.Confirm("\nGitHubにリポジトリを作成しますか？", true) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// Step 1: GitHubリポジトリを作成
		fmt.Println("\n[1/4] GitHubにリポジトリを作成しています...")
		repoURL, err := createGitHubRepository(repoName, visibility, description)
		if err != nil {
			return fmt.Errorf("リポジトリの作成に失敗しました: %w", err)
		}
		fmt.Printf("✓ リポジトリを作成しました: %s\n", repoURL)

		// Step 2: リポジトリをクローン
		fmt.Println("\n[2/4] リポジトリをクローンしています...")
		if err := cloneRepo(repoURL); err != nil {
			return fmt.Errorf("クローンに失敗しました: %w", err)
		}
		fmt.Println("✓ リポジトリをクローンしました")

		// Step 3: クローンしたディレクトリに移動
		fmt.Println("\n[3/4] ディレクトリに移動しています...")
		cloneDir := filepath.Join(".", repoName)
		if err := os.Chdir(cloneDir); err != nil {
			return fmt.Errorf("ディレクトリの移動に失敗しました: %w", err)
		}
		currentDir, _ := os.Getwd()
		fmt.Printf("✓ ディレクトリに移動しました: %s\n", currentDir)

		// Step 4: VSCodeを開く
		fmt.Println("\n[4/4] VSCodeを開いています...")
		if err := launchVSCode(); err != nil {
			fmt.Printf("警告: VSCodeの起動に失敗しました: %v\n", err)
			fmt.Println("手動で 'code .' を実行してください。")
		} else {
			fmt.Println("✓ VSCodeを開きました")
		}

		fmt.Println("\n✓ すべての処理が完了しました！")
		return nil
	},
}

// promptForVisibility はユーザーにリポジトリの公開設定を尋ねます。
//
// 戻り値:
//   - string: "public" または "private"
func promptForVisibility() string {
	fmt.Print("公開設定を選択してください [public/private] (デフォルト: private): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "private"
	}

	visibility := strings.TrimSpace(strings.ToLower(input))
	if visibility == "public" {
		return "public"
	}
	return "private"
}

// promptForDescription はユーザーにリポジトリの説明文を尋ねます。
//
// 戻り値:
//   - string: 入力された説明文（空の場合もあり）
func promptForDescription() string {
	fmt.Print("リポジトリの説明を入力してください (省略可): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(input)
}

// createGitHubRepository は GitHub CLI (gh) を使用してリポジトリを作成します。
//
// パラメータ:
//   name: リポジトリ名
//   visibility: "public" または "private"
//   description: リポジトリの説明文
//
// 戻り値:
//   - string: 作成されたリポジトリのURL
//   - error: エラーが発生した場合はエラーオブジェクト
func createGitHubRepository(name, visibility, description string) (string, error) {
	args := []string{"repo", "create", name}

	// 公開設定を追加
	if visibility == "public" {
		args = append(args, "--public")
	} else {
		args = append(args, "--private")
	}

	// 説明文が指定されている場合は追加
	if description != "" {
		args = append(args, "--description", description)
	}

	// gh repo create コマンドを実行
	ghCmd := exec.Command("gh", args...)
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(output))
	}

	// 出力からリポジトリ URL を抽出
	repoURL := strings.TrimSpace(string(output))
	lines := strings.Split(repoURL, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "https://github.com/") {
			return strings.TrimSpace(line), nil
		}
	}

	// URL が見つからない場合はデフォルトの URL を返す
	return fmt.Sprintf("https://github.com/%s", name), nil
}

// cloneRepo は指定された URL のリポジトリをクローンします。
//
// パラメータ:
//   repoURL: クローンするリポジトリの URL
//
// 戻り値:
//   error: エラーが発生した場合はエラーオブジェクト
func cloneRepo(repoURL string) error {
	gitCmd := exec.Command("git", "clone", repoURL)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin
	return gitCmd.Run()
}

// launchVSCode は現在のディレクトリで VSCode を起動します。
//
// 戻り値:
//   error: エラーが発生した場合はエラーオブジェクト
func launchVSCode() error {
	codeCmd := exec.Command("code", ".")
	codeCmd.Stdout = os.Stdout
	codeCmd.Stderr = os.Stderr
	return codeCmd.Run()
}

// init はコマンドの初期化を行います。
// createRepositoryCmd を RootCmd に登録することで、CLI から実行可能にします。
func init() {
	cmd.RootCmd.AddCommand(createRepositoryCmd)
}
