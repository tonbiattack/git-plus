package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tonbiattack/git-plus/internal/ui"
)

// main はGitHubリポジトリの作成からクローン、VSCode起動までを実行するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. リポジトリ名の引数チェック
//  3. リポジトリの公開設定（public/private）を確認
//  4. リポジトリの説明（Description）を確認
//  5. GitHubにリモートリポジトリを作成
//  6. 作成したリポジトリをクローン
//  7. クローンしたディレクトリに移動
//  8. VSCodeでプロジェクトを開く
//
// 終了コード:
//   - 0: 正常終了
//   - 1: エラー発生
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// リポジトリ名の引数チェック
	if len(os.Args) < 2 {
		fmt.Println("リポジトリ名を指定してください。")
		fmt.Println("使い方: git create-repository <リポジトリ名>")
		os.Exit(1)
	}
	repoName := os.Args[1]

	fmt.Printf("リポジトリ名: %s\n", repoName)

	// 公開設定の確認
	visibility := askForVisibility()
	fmt.Printf("公開設定: %s\n", visibility)

	// 説明の確認
	description := askForDescription()
	if description != "" {
		fmt.Printf("説明: %s\n", description)
	}

	// 確認
	if !ui.Confirm("\nGitHubにリポジトリを作成しますか？", true) {
		fmt.Println("キャンセルしました。")
		os.Exit(0)
	}

	// Step 1: GitHubリポジトリを作成
	fmt.Println("\n[1/4] GitHubにリポジトリを作成しています...")
	repoURL, err := createRepository(repoName, visibility, description)
	if err != nil {
		fmt.Printf("エラー: リポジトリの作成に失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ リポジトリを作成しました: %s\n", repoURL)

	// Step 2: リポジトリをクローン
	fmt.Println("\n[2/4] リポジトリをクローンしています...")
	if err := cloneRepository(repoURL); err != nil {
		fmt.Printf("エラー: クローンに失敗しました: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ リポジトリをクローンしました")

	// Step 3: クローンしたディレクトリに移動
	fmt.Println("\n[3/4] ディレクトリに移動しています...")
	cloneDir := filepath.Join(".", repoName)
	if err := os.Chdir(cloneDir); err != nil {
		fmt.Printf("エラー: ディレクトリの移動に失敗しました: %v\n", err)
		os.Exit(1)
	}
	currentDir, _ := os.Getwd()
	fmt.Printf("✓ ディレクトリに移動しました: %s\n", currentDir)

	// Step 4: VSCodeを開く
	fmt.Println("\n[4/4] VSCodeを開いています...")
	if err := openVSCode(); err != nil {
		fmt.Printf("警告: VSCodeの起動に失敗しました: %v\n", err)
		fmt.Println("手動で 'code .' を実行してください。")
	} else {
		fmt.Println("✓ VSCodeを開きました")
	}

	fmt.Println("\n✓ すべての処理が完了しました！")
}

// askForVisibility はリポジトリの公開設定をユーザーに確認する
//
// 戻り値:
//   - string: "public" または "private"
func askForVisibility() string {
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

// askForDescription はリポジトリの説明をユーザーに確認する
//
// 戻り値:
//   - string: リポジトリの説明（空の場合もある）
func askForDescription() string {
	fmt.Print("リポジトリの説明を入力してください (省略可): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}

	return strings.TrimSpace(input)
}

// createRepository はGitHubにリポジトリを作成する
//
// パラメータ:
//   - name: リポジトリ名
//   - visibility: "public" または "private"
//   - description: リポジトリの説明
//
// 戻り値:
//   - string: 作成されたリポジトリのURL
//   - error: gh コマンドの実行エラー
func createRepository(name, visibility, description string) (string, error) {
	args := []string{"repo", "create", name}

	// 公開設定を追加
	if visibility == "public" {
		args = append(args, "--public")
	} else {
		args = append(args, "--private")
	}

	// 説明を追加
	if description != "" {
		args = append(args, "--description", description)
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, string(output))
	}

	// リポジトリURLを取得
	repoURL := strings.TrimSpace(string(output))
	// gh repo create の出力から URL を抽出
	lines := strings.Split(repoURL, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "https://github.com/") {
			return strings.TrimSpace(line), nil
		}
	}

	// URLが見つからない場合は、リポジトリ名から構築
	return fmt.Sprintf("https://github.com/%s", name), nil
}

// cloneRepository はリポジトリをクローンする
//
// パラメータ:
//   - repoURL: クローンするリポジトリのURL
//
// 戻り値:
//   - error: git コマンドの実行エラー
func cloneRepository(repoURL string) error {
	cmd := exec.Command("git", "clone", repoURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// openVSCode はVSCodeを起動する
//
// 戻り値:
//   - error: code コマンドの実行エラー
func openVSCode() error {
	cmd := exec.Command("code", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// printHelp はコマンドのヘルプメッセージを表示する
func printHelp() {
	help := `git create-repository - GitHubリポジトリの作成からクローン、VSCode起動まで

使い方:
  git create-repository <リポジトリ名>

引数:
  リポジトリ名          作成するGitHubリポジトリの名前

説明:
  以下の処理を自動的に実行します:
  1. GitHubにリモートリポジトリを作成（public/private選択可能、Description指定可能）
  2. 作成したリポジトリをクローン
  3. クローンしたディレクトリに移動
  4. VSCodeでプロジェクトを開く

オプション:
  -h                    このヘルプを表示

使用例:
  git create-repository my-new-project       # my-new-projectリポジトリを作成

使用方法:
  1. コマンドを実行してリポジトリ名を指定
  2. 公開設定（public/private）を選択
  3. 説明を入力（省略可）
  4. 確認メッセージで y を入力
  5. 自動的にリポジトリ作成→クローン→移動→VSCode起動を実行

注意事項:
  - GitHub CLI (gh) がインストールされている必要があります
  - gh auth login でログイン済みである必要があります
  - VSCode (code コマンド) がパスに含まれている必要があります

使用する主なコマンド:
  - gh repo create: GitHubリポジトリの作成
  - git clone: リポジトリのクローン
  - code .: VSCodeの起動
`
	fmt.Print(help)
}
