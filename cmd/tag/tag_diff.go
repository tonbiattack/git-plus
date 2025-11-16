/*
Package tag は git の拡張コマンドのうち、タグ関連コマンドを定義します。

このファイル (tag_diff.go) は、2つのタグ間の差分を取得するコマンドを提供します。
指定された2つのタグ間のコミット差分をファイルに出力します。

主な機能:
  - 2つのタグ間のコミット差分の取得
  - マージコミットの自動除外
  - ファイルへの差分出力
  - コミット数の集計

使用例:

	git tag-diff V4.2.00.00 V4.3.00.00  # 2つのタグ間の差分を取得
*/
package tag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// tagDiffCmd は2つのタグ間のコミット差分を取得するコマンドです。
// git log を使用して、指定されたタグ間の差分をファイルに出力します。
var tagDiffCmd = &cobra.Command{
	Use:   "tag-diff <古いタグ> <新しいタグ>",
	Short: "2つのタグ間の差分を取得",
	Long: `2つのタグ間のコミット差分を取得し、ファイルに出力します。
Mergeコミットは自動的に除外されます。
出力形式: - コミットメッセージ (作成者名, 日付)`,
	Example: `  git tag-diff V4.2.00.00 V4.3.00.00`,
	Args:    cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		oldTag := args[0]
		newTag := args[1]

		// 出力ファイル名の自動生成
		// 例: tag_diff_v1.0.0_to_v2.0.0.txt
		outputFile := fmt.Sprintf("tag_diff_%s_to_%s.txt", oldTag, newTag)

		// タグの存在確認
		// 両方のタグが存在することを確認してから差分を取得
		if err := verifyTagExists(oldTag); err != nil {
			return fmt.Errorf("タグ '%s' が存在しません", oldTag)
		}
		if err := verifyTagExists(newTag); err != nil {
			return fmt.Errorf("タグ '%s' が存在しません", newTag)
		}

		// git logコマンドの実行
		// --no-merges: マージコミットを除外
		// --pretty=format: 出力フォーマットを指定
		tagRange := fmt.Sprintf("%s..%s", oldTag, newTag)
		output, err := gitcmd.Run("log", tagRange, "--no-merges", "--pretty=format:- %s (%an, %ad)", "--date=iso")
		if err != nil {
			return fmt.Errorf("git logの実行に失敗しました: %w", err)
		}

		// 出力が空の場合は差分なし
		if len(output) == 0 {
			fmt.Printf("タグ %s と %s の間に差分はありません。\n", oldTag, newTag)
			return nil
		}

		// ファイルに書き込み
		absPath, err := filepath.Abs(outputFile)
		if err != nil {
			return fmt.Errorf("ファイルパスの取得に失敗しました: %w", err)
		}

		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			return fmt.Errorf("ファイルへの書き込みに失敗しました: %w", err)
		}

		// サマリー表示
		commits := strings.Split(string(output), "\n")
		fmt.Printf("✓ タグ %s と %s の差分を %s に出力しました。\n", oldTag, newTag, absPath)
		fmt.Printf("  コミット数: %d\n", len(commits))
		return nil
	},
}

// verifyTagExists は指定されたタグが存在するかを確認します。
//
// パラメータ:
//
//	tag: 確認するタグ名
//
// 戻り値:
//
//	タグが存在する場合は nil、存在しない場合はエラー
func verifyTagExists(tag string) error {
	return gitcmd.RunQuiet("rev-parse", "--verify", tag)
}

// init はコマンドの初期化を行います。
// tagDiffCmd を RootCmd に登録することで、CLI から実行可能にします。
func init() {
	cmd.RootCmd.AddCommand(tagDiffCmd)
}
