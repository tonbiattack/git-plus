// ================================================================================
// tag_checkout.go
// ================================================================================
// このファイルは git-plus の tag-checkout コマンドを実装しています。
//
// 【概要】
// tag-checkout コマンドは、最新のタグを取得してチェックアウトする機能を
// 提供します。セマンティックバージョン順で最新のタグを表示し、確認後に
// チェックアウトします。
//
// 【主な機能】
// - セマンティックバージョン順で最新のタグを取得（--sort=-v:refname）
// - 最新N個のタグを表示（デフォルト: 10個）
// - 対話的にタグを選択してチェックアウト
// - 最新タグに自動チェックアウト（-y オプション）
//
// 【使用例】
//   git-plus tag-checkout                 # 最新10個のタグから選択
//   git-plus tag-checkout -n 5            # 最新5個のタグから選択
//   git-plus tag-checkout -y              # 最新タグに自動チェックアウト
//   git-plus tag-checkout --limit 20      # 最新20個のタグから選択
//
// 【ソート方法】
// git tag --sort=-v:refname を使用してセマンティックバージョン順で
// 新しいもの → 古いものの順に並べます。
// ================================================================================

package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/ui"
)

var (
	tagLimit   int  // 表示するタグの数
	autoYes    bool // 確認なしで最新タグにチェックアウト
	showLatest bool // 最新タグのみを表示して終了
)

// tagCheckoutCmd は tag-checkout コマンドの定義です。
// セマンティックバージョン順で最新のタグを取得してチェックアウトします。
var tagCheckoutCmd = &cobra.Command{
	Use:   "tag-checkout",
	Short: "最新のタグを取得してチェックアウト",
	Long: `セマンティックバージョン順で最新のタグを取得してチェックアウトします。
デフォルトでは最新10個のタグを表示し、選択してチェックアウトできます。`,
	Example: `  git-plus tag-checkout                 # 最新10個のタグから選択
  git-plus tag-checkout -n 5            # 最新5個のタグから選択
  git-plus tag-checkout -y              # 最新タグに自動チェックアウト
  git-plus tag-checkout --limit 20      # 最新20個のタグから選択
  git-plus tag-checkout --latest        # 最新タグを表示するのみ`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 全てのタグを最新順に取得
		tags, err := getTagsSortedByVersion()
		if err != nil {
			return fmt.Errorf("タグの取得に失敗しました: %w", err)
		}

		if len(tags) == 0 {
			fmt.Println("タグが見つかりませんでした")
			return nil
		}

		// 表示するタグの数を制限
		displayTags := tags
		if tagLimit > 0 && tagLimit < len(tags) {
			displayTags = tags[:tagLimit]
		}

		// --latest オプション: 最新タグのみを表示して終了
		if showLatest {
			fmt.Printf("最新のタグ: %s\n", displayTags[0])
			return nil
		}

		// -y オプション: 確認なしで最新タグにチェックアウト
		if autoYes {
			latestTag := displayTags[0]
			fmt.Printf("最新のタグ: %s\n", latestTag)
			return checkoutTag(latestTag)
		}

		// タグ一覧を表示
		fmt.Printf("最新のタグ（セマンティックバージョン順）:\n\n")
		for i, tag := range displayTags {
			fmt.Printf("  [%d] %s\n", i+1, tag)
		}
		fmt.Println()

		// タグを選択
		fmt.Print("チェックアウトするタグの番号を入力してください (Enterでキャンセル): ")
		var input string
		fmt.Scanln(&input)

		if input == "" {
			fmt.Println("キャンセルしました")
			return nil
		}

		// 入力を解析
		input = ui.NormalizeNumberInput(input)
		selectedIndex, err := strconv.Atoi(input)
		if err != nil || selectedIndex < 1 || selectedIndex > len(displayTags) {
			return fmt.Errorf("無効な番号です: %s", input)
		}

		selectedTag := displayTags[selectedIndex-1]

		// 確認プロンプト
		if !ui.Confirm(fmt.Sprintf("タグ '%s' にチェックアウトしますか？", selectedTag), true) {
			fmt.Println("キャンセルしました")
			return nil
		}

		// チェックアウト実行
		return checkoutTag(selectedTag)
	},
}

// getTagsSortedByVersion は全てのタグをセマンティックバージョン順（最新→古い）で取得します。
//
// 戻り値:
//   - []string: タグのリスト（最新順）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git tag --sort=-v:refname コマンドでセマンティックバージョン順に
//   ソートされたタグ一覧を取得します。
func getTagsSortedByVersion() ([]string, error) {
	cmd := exec.Command("git", "tag", "--sort=-v:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	tagsStr := strings.TrimSpace(string(output))
	if tagsStr == "" {
		return []string{}, nil
	}

	tags := strings.Split(tagsStr, "\n")
	return tags, nil
}

// checkoutTag は指定されたタグにチェックアウトします。
//
// パラメータ:
//   - tag: チェックアウトするタグ名
//
// 戻り値:
//   - error: チェックアウトに失敗した場合のエラー情報
//
// 内部処理:
//   git checkout <tag> コマンドを実行してタグにチェックアウトします。
//   チェックアウトの出力はユーザーに表示されます。
func checkoutTag(tag string) error {
	fmt.Printf("タグ '%s' にチェックアウト中...\n", tag)
	cmd := exec.Command("git", "checkout", tag)
	cmd.Stdout = nil
	cmd.Stderr = nil
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(string(output))
		return fmt.Errorf("チェックアウトに失敗しました: %w", err)
	}

	fmt.Println(string(output))
	fmt.Printf("✓ タグ '%s' にチェックアウトしました\n", tag)
	return nil
}

// init は tag-checkout コマンドを root コマンドに登録し、フラグを設定します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
//
// 設定されるフラグ:
//   -n, --limit: 表示するタグの数（デフォルト: 10）
//   -y, --yes: 確認なしで最新タグにチェックアウト
//   --latest: 最新タグのみを表示して終了
func init() {
	tagCheckoutCmd.Flags().IntVarP(&tagLimit, "limit", "n", 10, "表示するタグの数")
	tagCheckoutCmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "確認なしで最新タグにチェックアウト")
	tagCheckoutCmd.Flags().BoolVar(&showLatest, "latest", false, "最新タグのみを表示")
	rootCmd.AddCommand(tagCheckoutCmd)
}
