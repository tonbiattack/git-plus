package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/tonbiattack/git-plus/internal/ui"
)

func main() {
	args := os.Args[1:]

	// ヘルプオプションのチェック
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		printHelp()
		os.Exit(0)
	}

	// 最新タグを取得
	currentTag, err := getCurrentTag()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 最新タグの取得に失敗: %v\n", err)
		fmt.Println("タグが存在しない可能性があります。最初のタグを手動で作成してください。")
		os.Exit(1)
	}

	// バージョンを解析
	major, minor, patch, err := parseVersion(currentTag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: バージョンの解析に失敗: %v\n", err)
		fmt.Printf("現在のタグ: %s\n", currentTag)
		os.Exit(1)
	}

	fmt.Printf("現在のタグ: %s\n", currentTag)

	var versionType string
	var message string
	var push bool
	var dryRun bool

	// 引数の解析
	if len(args) == 0 {
		// 対話的モード
		versionType = interactiveMode(major, minor, patch)
	} else {
		// コマンドライン引数からタイプを取得
		versionType = normalizeVersionType(args[0])
		if versionType == "" {
			fmt.Fprintf(os.Stderr, "エラー: 無効なバージョンタイプ: %s\n", args[0])
			fmt.Println("使用可能なタイプ: major, minor, patch, feature, bug, fix, m, n, p, f, b")
			os.Exit(1)
		}

		// オプションの解析
		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "-m", "--message":
				if i+1 < len(args) {
					message = args[i+1]
					i++
				} else {
					fmt.Fprintf(os.Stderr, "エラー: -m オプションにはメッセージが必要です\n")
					os.Exit(1)
				}
			case "--push":
				push = true
			case "--dry-run":
				dryRun = true
			default:
				fmt.Fprintf(os.Stderr, "警告: 不明なオプション: %s\n", args[i])
			}
		}
	}

	// 新しいバージョンを計算
	newMajor, newMinor, newPatch := calculateNewVersion(major, minor, patch, versionType)
	newTag := fmt.Sprintf("v%d.%d.%d", newMajor, newMinor, newPatch)
	versionTypeDisplay := strings.ToUpper(versionType)

	fmt.Printf("新しいタグ: %s (%s)\n", newTag, versionTypeDisplay)
	if message != "" {
		fmt.Printf("メッセージ: %s\n", message)
	}

	// --dry-run の場合はここで終了
	if dryRun {
		fmt.Println("(--dry-run のため、タグは作成されません)")
		return
	}

	// 確認プロンプト
	if !ui.Confirm("タグを作成しますか？", true) {
		fmt.Println("キャンセルしました")
		return
	}

	// タグを作成
	if err := createTag(newTag, message); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: タグの作成に失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ タグを作成しました: %s\n", newTag)

	// --push オプションが指定されている場合、または対話モードでプッシュ確認がYesの場合
	shouldPush := push
	if !push && len(args) == 0 {
		// 対話モードの場合はプッシュするか確認
		shouldPush = ui.Confirm("\nリモートにプッシュしますか？", true)
	}

	if shouldPush {
		if err := pushTag(newTag); err != nil {
			fmt.Fprintf(os.Stderr, "エラー: タグのプッシュに失敗: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ リモートにプッシュしました: %s\n", newTag)
	}
}

// getCurrentTag は最新のタグを取得
func getCurrentTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// parseVersion はタグからバージョン番号を抽出
func parseVersion(tag string) (major, minor, patch int, err error) {
	// v1.2.3 または 1.2.3 の形式に対応
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("無効なバージョン形式: %s", tag)
	}

	major, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, err
	}

	minor, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, err
	}

	patch, err = strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, err
	}

	return major, minor, patch, nil
}

// normalizeVersionType はバージョンタイプを正規化
func normalizeVersionType(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "major", "m", "breaking":
		return "major"
	case "minor", "n", "feature", "f":
		return "minor"
	case "patch", "p", "bug", "b", "fix":
		return "patch"
	default:
		return ""
	}
}

// calculateNewVersion は新しいバージョン番号を計算
func calculateNewVersion(major, minor, patch int, versionType string) (newMajor, newMinor, newPatch int) {
	switch versionType {
	case "major":
		return major + 1, 0, 0
	case "minor":
		return major, minor + 1, 0
	case "patch":
		return major, minor, patch + 1
	default:
		// デフォルトは patch
		return major, minor, patch + 1
	}
}

// interactiveMode は対話的にバージョンタイプを選択
func interactiveMode(major, minor, patch int) string {
	fmt.Println("\n新しいタグのタイプを選択してください:")
	fmt.Printf("  [1] major   - v%d.0.0 (破壊的変更)\n", major+1)
	fmt.Printf("  [2] minor   - v%d.%d.0 (機能追加)\n", major, minor+1)
	fmt.Printf("  [3] patch   - v%d.%d.%d (バグ修正)\n", major, minor, patch+1)
	fmt.Print("選択 (1-3): ")

	var input string
	fmt.Scanln(&input)

	switch strings.TrimSpace(input) {
	case "1":
		return "major"
	case "2":
		return "minor"
	case "3":
		return "patch"
	default:
		fmt.Println("無効な選択です。patch を使用します。")
		return "patch"
	}
}

// createTag はタグを作成
func createTag(tag, message string) error {
	var cmd *exec.Cmd
	if message != "" {
		// アノテーテッドタグ
		cmd = exec.Command("git", "tag", "-a", tag, "-m", message)
	} else {
		// ライトウェイトタグ
		cmd = exec.Command("git", "tag", tag)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// pushTag はタグをリモートにプッシュ
func pushTag(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// printHelp はヘルプメッセージを表示
func printHelp() {
	help := `git new-tag - セマンティックバージョニングに従って新しいタグを作成

使い方:
  git new-tag [type] [options]

バージョンタイプ:
  major, m, breaking    メジャーバージョンアップ (v1.2.3 → v2.0.0)
  minor, n, feature, f  マイナーバージョンアップ (v1.2.3 → v1.3.0)
  patch, p, bug, b, fix パッチバージョンアップ   (v1.2.3 → v1.2.4)

オプション:
  -h, --help            このヘルプを表示
  -m, --message <msg>   タグメッセージを指定（アノテーテッドタグを作成）
  --push                作成後に自動的にリモートへプッシュ
  --dry-run             実際には作成せず、次のバージョンだけを表示

例:
  git new-tag                      # 対話的にタイプを選択
  git new-tag feature              # 機能追加（minor）
  git new-tag f                    # 機能追加の省略形
  git new-tag bug                  # バグ修正（patch）
  git new-tag b                    # バグ修正の省略形
  git new-tag major                # 破壊的変更
  git new-tag feature --push       # 作成してプッシュ
  git new-tag bug -m "Fix issue"   # メッセージ付きで作成
  git new-tag minor --dry-run      # 確認のみ

バージョンタイプのマッピング:
  機能追加: feature, f → minor
  バグ修正: bug, b, fix → patch
  破壊的変更: major, m, breaking → major
`
	fmt.Print(help)
}
