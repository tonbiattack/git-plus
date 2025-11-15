// ================================================================================
// new_tag.go
// ================================================================================
// このファイルは git-plus の new-tag コマンドを実装しています。
//
// 【概要】
// new-tag コマンドは、セマンティックバージョニング（SemVer）に従って
// 新しいタグを作成する機能を提供します。現在のバージョンから自動的に
// 次のバージョンを計算します。
//
// 【主な機能】
// - セマンティックバージョニング（major.minor.patch）のサポート
// - バージョンタイプの対話的選択（major/minor/patch）
// - バージョンタイプの短縮形と別名のサポート
//   - major/m/breaking: メジャーバージョンアップ（破壊的変更）
//   - minor/n/feature/f: マイナーバージョンアップ（機能追加）
//   - patch/p/bug/b/fix: パッチバージョンアップ（バグ修正）
// - タグメッセージの指定（-m オプション）
// - 作成後の自動プッシュ（--push オプション）
// - プッシュ後の自動リリース作成（--release オプション）
// - リリースのドラフト作成（--release-draft オプション）
// - プレリリース作成（--release-prerelease オプション）
// - ドライラン（--dry-run オプション）
//
// 【使用例】
//   git-plus new-tag                      # 対話的にタイプを選択
//   git-plus new-tag feature              # 機能追加（minor）
//   git-plus new-tag bug                  # バグ修正（patch）
//   git-plus new-tag major                # 破壊的変更
//   git-plus new-tag feature --push       # 作成してプッシュ
//   git-plus new-tag bug -m "Fix issue"   # メッセージ付きで作成
//   git-plus new-tag minor --dry-run      # 確認のみ
//   git-plus new-tag feature --push --release              # タグ作成、プッシュ、リリース作成
//   git-plus new-tag bug --push --release --release-draft  # ドラフトリリースとして作成
//
// 【バージョン形式】
// - 形式: v<major>.<minor>.<patch>
// - 例: v1.2.3 → v1.3.0 (minor), v1.2.4 (patch), v2.0.0 (major)
// ================================================================================

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	tagMessage         string // タグメッセージ（アノテーテッドタグ用）
	tagPush            bool   // 作成後に自動的にリモートへプッシュするフラグ
	tagDryRun          bool   // 実際には作成せず、次のバージョンだけを表示するフラグ
	tagRelease         bool   // プッシュ後に自動的にGitHubリリースを作成するフラグ
	tagReleaseDraft    bool   // リリースをドラフトとして作成するフラグ
	tagReleasePrerelease bool   // リリースをプレリリースとして作成するフラグ
)

// newTagCmd は new-tag コマンドの定義です。
// セマンティックバージョニングに従って新しいタグを作成します。
var newTagCmd = &cobra.Command{
	Use:   "new-tag [type]",
	Short: "セマンティックバージョニングに従って新しいタグを作成",
	Long: `セマンティックバージョニング（SemVer）に従って新しいタグを作成します。
引数なしで実行すると対話的にバージョンタイプを選択できます。
--releaseフラグを使用すると、タグプッシュ後に自動的にGitHubリリースも作成できます。`,
	Example: `  git-plus new-tag                      # 対話的にタイプを選択
  git-plus new-tag feature              # 機能追加（minor）
  git-plus new-tag f                    # 機能追加の省略形
  git-plus new-tag bug                  # バグ修正（patch）
  git-plus new-tag b                    # バグ修正の省略形
  git-plus new-tag major                # 破壊的変更
  git-plus new-tag feature --push       # 作成してプッシュ
  git-plus new-tag bug -m "Fix issue"   # メッセージ付きで作成
  git-plus new-tag minor --dry-run      # 確認のみ
  git-plus new-tag feature --push --release              # タグ作成、プッシュ、リリース作成
  git-plus new-tag bug --push --release --release-draft  # ドラフトリリースとして作成`,
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

			// --release フラグが指定されている場合、または対話モードで確認された場合にリリースを作成
			shouldRelease := tagRelease
			if !tagRelease && len(args) == 0 {
				// 対話モードの場合はリリースを作成するか確認
				shouldRelease = ui.Confirm("\nGitHubリリースを作成しますか？", false)
			}

			if shouldRelease {
				// GitHub CLI の確認
				if !checkGitHubCLIInstalled() {
					fmt.Println("警告: GitHub CLI (gh) がインストールされていないため、リリースを作成できません")
					fmt.Println("インストール方法: https://cli.github.com/")
					return nil
				}

				fmt.Printf("\nGitHubリリースを作成中...\n")
				if err := createReleaseFromTag(newTag, tagReleaseDraft, tagReleasePrerelease); err != nil {
					return fmt.Errorf("リリースの作成に失敗: %w", err)
				}
				fmt.Printf("✓ GitHubリリースを作成しました\n")
				fmt.Printf("詳細を確認するには: gh release view %s --web\n", newTag)
			}
		}

		return nil
	},
}

// getLatestTag は最新のタグを取得します。
//
// 戻り値:
//   - string: 最新のタグ名（例: v1.2.3）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git describe --tags --abbrev=0 コマンドで最新のタグを取得します。
func getLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractVersion はタグからバージョン番号を抽出します。
//
// パラメータ:
//   - tag: タグ名（例: v1.2.3 または 1.2.3）
//
// 戻り値:
//   - major: メジャーバージョン
//   - minor: マイナーバージョン
//   - patch: パッチバージョン
//   - err: 解析に失敗した場合のエラー情報
//
// 内部処理:
//   正規表現を使用してバージョン番号を抽出します。
//   v プレフィックスの有無に対応しています。
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

// normalizeVersionTypeName はバージョンタイプ名を正規化します。
//
// パラメータ:
//   - input: ユーザーが入力したバージョンタイプ
//
// 戻り値:
//   - string: 正規化されたバージョンタイプ（"major", "minor", "patch"）
//            無効な入力の場合は空文字列
//
// サポートする入力:
//   - major: major, m, breaking
//   - minor: minor, n, feature, f
//   - patch: patch, p, bug, b, fix
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

// computeNewVersion は新しいバージョン番号を計算します。
//
// パラメータ:
//   - major: 現在のメジャーバージョン
//   - minor: 現在のマイナーバージョン
//   - patch: 現在のパッチバージョン
//   - versionType: バージョンアップのタイプ（"major", "minor", "patch"）
//
// 戻り値:
//   - newMajor: 新しいメジャーバージョン
//   - newMinor: 新しいマイナーバージョン
//   - newPatch: 新しいパッチバージョン
//
// ルール:
//   - major: メジャーを+1、マイナーとパッチを0にリセット
//   - minor: マイナーを+1、パッチを0にリセット
//   - patch: パッチを+1
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

// interactiveVersionSelection は対話的にバージョンタイプを選択します。
//
// パラメータ:
//   - major: 現在のメジャーバージョン
//   - minor: 現在のマイナーバージョン
//   - patch: 現在のパッチバージョン
//
// 戻り値:
//   - string: 選択されたバージョンタイプ（"major", "minor", "patch"）
//
// 内部処理:
//   各バージョンタイプの説明と新しいバージョンの例を表示し、
//   ユーザーに選択を促します。無効な選択の場合は patch をデフォルトとします。
func interactiveVersionSelection(major, minor, patch int) string {
	fmt.Println("\n新しいタグのタイプを選択してください:")
	fmt.Printf("  [1] major   - v%d.0.0 (破壊的変更)\n", major+1)
	fmt.Printf("  [2] minor   - v%d.%d.0 (機能追加)\n", major, minor+1)
	fmt.Printf("  [3] patch   - v%d.%d.%d (バグ修正)\n", major, minor, patch+1)
	fmt.Print("選択 (1-3): ")

	var input string
	fmt.Scanln(&input)

	switch ui.NormalizeNumberInput(input) {
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

// makeTag は指定されたタグを作成します。
//
// パラメータ:
//   - tag: 作成するタグ名（例: v1.2.3）
//   - message: タグメッセージ（空文字列の場合は軽量タグを作成）
//
// 戻り値:
//   - error: タグの作成に失敗した場合のエラー情報
//
// 内部処理:
//   メッセージが指定されている場合は git tag -a <tag> -m <message> で
//   アノテーテッドタグを作成し、そうでない場合は git tag <tag> で
//   軽量タグを作成します。
func makeTag(tag, message string) error {
	var cmd *exec.Cmd
	if message != "" {
		cmd = exec.Command("git", "tag", "-a", tag, "-m", message)
	} else {
		cmd = exec.Command("git", "tag", tag)
	}
	return cmd.Run()
}

// pushTagToRemote は指定されたタグをリモートリポジトリにプッシュします。
//
// パラメータ:
//   - tag: プッシュするタグ名
//
// 戻り値:
//   - error: プッシュに失敗した場合のエラー情報
//
// 内部処理:
//   git push origin <tag> コマンドでリモートにタグをプッシュします。
func pushTagToRemote(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	return cmd.Run()
}

// createReleaseFromTag は指定されたタグからGitHubリリースを作成します。
//
// パラメータ:
//   - tag: リリースを作成するタグ
//   - draft: ドラフトとして作成するかどうか
//   - prerelease: プレリリースとして作成するかどうか
//
// 戻り値:
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   gh release create コマンドで GitHub リリースを作成します。
//   --generate-notes オプションで自動的にリリースノートを生成します。
func createReleaseFromTag(tag string, draft, prerelease bool) error {
	args := []string{"release", "create", tag, "--generate-notes"}

	if draft {
		args = append(args, "--draft")
	}
	if prerelease {
		args = append(args, "--prerelease")
	}

	cmd := exec.Command("gh", args...)

	// 出力をキャプチャして表示
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// エラーメッセージを表示
		if stderr.Len() > 0 {
			return fmt.Errorf("%w\n%s", err, stderr.String())
		}
		return err
	}

	// 成功メッセージを表示
	if stdout.Len() > 0 {
		fmt.Println(stdout.String())
	}

	return nil
}

// init は new-tag コマンドを root コマンドに登録し、フラグを設定します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
//
// 設定されるフラグ:
//   -m, --message: タグメッセージを指定（アノテーテッドタグを作成）
//   --push: 作成後に自動的にリモートへプッシュ
//   --dry-run: 実際には作成せず、次のバージョンだけを表示
//   --release: プッシュ後に自動的にGitHubリリースを作成
//   --release-draft: リリースをドラフトとして作成
//   --release-prerelease: リリースをプレリリースとして作成
func init() {
	newTagCmd.Flags().StringVarP(&tagMessage, "message", "m", "", "タグメッセージを指定（アノテーテッドタグを作成）")
	newTagCmd.Flags().BoolVarP(&tagPush, "push", "p", false, "作成後に自動的にリモートへプッシュ")
	newTagCmd.Flags().BoolVar(&tagDryRun, "dry-run", false, "実際には作成せず、次のバージョンだけを表示")
	newTagCmd.Flags().BoolVar(&tagRelease, "release", false, "プッシュ後に自動的にGitHubリリースを作成")
	newTagCmd.Flags().BoolVar(&tagReleaseDraft, "release-draft", false, "リリースをドラフトとして作成")
	newTagCmd.Flags().BoolVar(&tagReleasePrerelease, "release-prerelease", false, "リリースをプレリリースとして作成")
	rootCmd.AddCommand(newTagCmd)
}
