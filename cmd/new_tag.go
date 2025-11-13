package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	tagMessage string
	tagPush    bool
	tagDryRun  bool
)

var newTagCmd = &cobra.Command{
	Use:   "new-tag [type]",
	Short: "セマンティックバージョニングに従って新しいタグを作成",
	Long: `セマンティックバージョニング（SemVer）に従って新しいタグを作成します。
引数なしで実行すると対話的にバージョンタイプを選択できます。`,
	Example: `  git-plus new-tag                      # 対話的にタイプを選択
  git-plus new-tag feature              # 機能追加（minor）
  git-plus new-tag f                    # 機能追加の省略形
  git-plus new-tag bug                  # バグ修正（patch）
  git-plus new-tag b                    # バグ修正の省略形
  git-plus new-tag major                # 破壊的変更
  git-plus new-tag feature --push       # 作成してプッシュ
  git-plus new-tag bug -m "Fix issue"   # メッセージ付きで作成
  git-plus new-tag minor --dry-run      # 確認のみ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 最新タグを取得
		currentTag, err := getLatestTag()
		if err != nil {
			fmt.Println("エラー: 最新タグの取得に失敗しました")
			fmt.Println("タグが存在しない可能性があります。最初のタグを手動で作成してください。")
			return err
		}

		// バージョンを解析
		major, minor, patch, err := extractVersion(currentTag)
		if err != nil {
			fmt.Printf("エラー: バージョンの解析に失敗: %v\n", err)
			fmt.Printf("現在のタグ: %s\n", currentTag)
			return err
		}

		fmt.Printf("現在のタグ: %s\n", currentTag)

		var versionType string

		// 引数の解析
		if len(args) == 0 {
			// 対話的モード
			versionType = interactiveVersionSelection(major, minor, patch)
		} else {
			// コマンドライン引数からタイプを取得
			versionType = normalizeVersionTypeName(args[0])
			if versionType == "" {
				return fmt.Errorf("無効なバージョンタイプ: %s\n使用可能なタイプ: major, minor, patch, feature, bug, fix, m, n, p, f, b", args[0])
			}
		}

		// 新しいバージョンを計算
		newMajor, newMinor, newPatch := computeNewVersion(major, minor, patch, versionType)
		newTag := fmt.Sprintf("v%d.%d.%d", newMajor, newMinor, newPatch)
		versionTypeDisplay := strings.ToUpper(versionType)

		fmt.Printf("新しいタグ: %s (%s)\n", newTag, versionTypeDisplay)
		if tagMessage != "" {
			fmt.Printf("メッセージ: %s\n", tagMessage)
		}

		// --dry-run の場合はここで終了
		if tagDryRun {
			fmt.Println("(--dry-run のため、タグは作成されません)")
			return nil
		}

		// 確認プロンプト
		if !ui.Confirm("タグを作成しますか？", true) {
			fmt.Println("キャンセルしました")
			return nil
		}

		// タグを作成
		if err := makeTag(newTag, tagMessage); err != nil {
			return fmt.Errorf("タグの作成に失敗: %w", err)
		}

		fmt.Printf("✓ タグを作成しました: %s\n", newTag)

		// --push オプションが指定されている場合、または対話モードでプッシュ確認がYesの場合
		shouldPush := tagPush
		if !tagPush && len(args) == 0 {
			// 対話モードの場合はプッシュするか確認
			shouldPush = ui.Confirm("\nリモートにプッシュしますか？", true)
		}

		if shouldPush {
			if err := pushTagToRemote(newTag); err != nil {
				return fmt.Errorf("タグのプッシュに失敗: %w", err)
			}
			fmt.Printf("✓ リモートにプッシュしました: %s\n", newTag)
		}

		return nil
	},
}

func getLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func extractVersion(tag string) (major, minor, patch int, err error) {
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

func normalizeVersionTypeName(input string) string {
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

func computeNewVersion(major, minor, patch int, versionType string) (newMajor, newMinor, newPatch int) {
	switch versionType {
	case "major":
		return major + 1, 0, 0
	case "minor":
		return major, minor + 1, 0
	case "patch":
		return major, minor, patch + 1
	default:
		return major, minor, patch + 1
	}
}

func interactiveVersionSelection(major, minor, patch int) string {
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

func makeTag(tag, message string) error {
	var cmd *exec.Cmd
	if message != "" {
		cmd = exec.Command("git", "tag", "-a", tag, "-m", message)
	} else {
		cmd = exec.Command("git", "tag", tag)
	}
	return cmd.Run()
}

func pushTagToRemote(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	return cmd.Run()
}

func init() {
	newTagCmd.Flags().StringVarP(&tagMessage, "message", "m", "", "タグメッセージを指定（アノテーテッドタグを作成）")
	newTagCmd.Flags().BoolVar(&tagPush, "push", false, "作成後に自動的にリモートへプッシュ")
	newTagCmd.Flags().BoolVar(&tagDryRun, "dry-run", false, "実際には作成せず、次のバージョンだけを表示")
	rootCmd.AddCommand(newTagCmd)
}
