package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var createRepositoryCmd = &cobra.Command{
	Use:   "create-repository <リポジトリ名>",
	Short: "GitHubリポジトリの作成からクローン、VSCode起動まで",
	Long: `以下の処理を自動的に実行します:
  1. GitHubにリモートリポジトリを作成（public/private選択可能、Description指定可能）
  2. 作成したリポジトリをクローン
  3. クローンしたディレクトリに移動
  4. VSCodeでプロジェクトを開く`,
	Example: `  git-plus create-repository my-new-project`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]
		fmt.Printf("リポジトリ名: %s\n", repoName)

		// 公開設定の確認
		visibility := promptForVisibility()
		fmt.Printf("公開設定: %s\n", visibility)

		// 説明の確認
		description := promptForDescription()
		if description != "" {
			fmt.Printf("説明: %s\n", description)
		}

		// 確認
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

func promptForDescription() string {
	fmt.Print("リポジトリの説明を入力してください (省略可): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(input)
}

func createGitHubRepository(name, visibility, description string) (string, error) {
	args := []string{"repo", "create", name}

	if visibility == "public" {
		args = append(args, "--public")
	} else {
		args = append(args, "--private")
	}

	if description != "" {
		args = append(args, "--description", description)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(output))
	}

	repoURL := strings.TrimSpace(string(output))
	lines := strings.Split(repoURL, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "https://github.com/") {
			return strings.TrimSpace(line), nil
		}
	}

	return fmt.Sprintf("https://github.com/%s", name), nil
}

func cloneRepo(repoURL string) error {
	cmd := exec.Command("git", "clone", repoURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func launchVSCode() error {
	cmd := exec.Command("code", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(createRepositoryCmd)
}
