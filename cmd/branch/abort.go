// ================================================================================
// abort.go
// ================================================================================
// このファイルは git の拡張コマンド abort を実装しています。
//
// 【概要】
// 進行中の Git 操作（rebase / merge / cherry-pick / revert）を安全に中止します。
// 引数を指定しない場合は現在の状態を判定し、該当する操作を自動で選択します。
//
// 【使用例】
//
//	git abort             # 状態から自動検出して中止
//	git abort merge       # マージを強制的に中止
//	git abort rebase      # リベースを強制的に中止
//
// ================================================================================
package branch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// abortCmd は進行中のGit操作を中止するコマンドです
var abortCmd = &cobra.Command{
	Use:   "abort [merge|rebase|cherry-pick|revert]",
	Short: "進行中のGit操作を安全に中止",
	Long: `進行中の rebase / merge / cherry-pick / revert を安全に中止します。

引数を指定しない場合は現在の状態を判定して自動的に操作を選択します。`,
	Example: `  git abort           # 自動検出して中止
  git abort merge   # マージを中止
  git abort rebase  # リベースを中止`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAbortCommand,
}

// runAbortCommand は abort コマンドのメイン処理です
func runAbortCommand(_ *cobra.Command, args []string) error {
	var (
		operation string
		err       error
	)

	if len(args) > 0 {
		operation, err = normalizeAbortOperation(args[0])
		if err != nil {
			return err
		}
	} else {
		operation, err = detectAbortOperation()
		if err != nil {
			return err
		}
	}

	label := abortOperationLabel(operation)
	fmt.Printf("%sを中止します...\n", label)

	if err := abortOperation(operation); err != nil {
		return fmt.Errorf("%sの中止に失敗しました: %w", label, err)
	}

	fmt.Println("中止が完了しました。")
	return nil
}

// normalizeAbortOperation はユーザー入力をサポートする操作名に変換します
func normalizeAbortOperation(op string) (string, error) {
	// 文字列の正規化処理
	// - `strings.TrimSpace` で前後の空白を除去
	// - `strings.ToLower` で大文字小文字を統一
	// - `strings.ReplaceAll` でアンダースコアをハイフンに置換
	// これにより、ユーザー入力のバリエーション（例: " ReBase", "cherrypick"）を
	// 受け付けやすくしています。
	normalized := strings.ToLower(strings.TrimSpace(op))
	// 入力の区切り文字を統一するため、アンダースコアをハイフンに置換します。
	// 例: "cherry_pick" -> "cherry-pick" として扱うことで、
	// ユーザーがアンダースコア/ハイフンどちらを使っても同一操作として扱えるようにします。
	normalized = strings.ReplaceAll(normalized, "_", "-")

	// switch 文で受け付ける操作名を決定します。
	// - 複数の case を列挙することで同義の入力を一つの正規形にまとめています。
	// - 成功時は正規化された操作名（例: "rebase"）を返し、エラー時は説明付きで返します。
	switch normalized {
	case "merge":
		return "merge", nil
	case "rebase":
		return "rebase", nil
	case "cherry", "cherry-pick", "cherrypick":
		// "cherry" を許容して "cherry-pick" に統一
		return "cherry-pick", nil
	case "revert":
		return "revert", nil
	default:
		// サポート外の操作の場合はエラーを返す
		return "", fmt.Errorf("サポートされていない操作です: %s", op)
	}
}

// detectAbortOperation は現在のGitディレクトリから進行中の操作を判定します
func detectAbortOperation() (string, error) {
	// getGitDir で .git ディレクトリの絶対パスを取得します。
	// エラーがあれば検出不能としてそのまま返します。
	gitDir, err := getGitDir()
	if err != nil {
		return "", err
	}

	// rebase は 2 種類の作業ディレクトリを持つため両方を確認します。
	// - rebase-apply: 非対話的/メールベースの rebase で使われる場合がある
	// - rebase-merge: 対話的 rebase 等で使われる場合がある
	rebaseDirs := []string{"rebase-apply", "rebase-merge"}
	for _, dir := range rebaseDirs {
		// filepath.Join は複数のパス要素を OS に依存しない形で結合します。
		if pathExists(filepath.Join(gitDir, dir)) {
			// 見つかった時点で rebase が進行中と判定
			return "rebase", nil
		}
	}

	// CHERRY_PICK_HEAD が存在すればチェリーピック中
	if pathExists(filepath.Join(gitDir, "CHERRY_PICK_HEAD")) {
		return "cherry-pick", nil
	}

	// REVERT_HEAD が存在すればリバート中
	if pathExists(filepath.Join(gitDir, "REVERT_HEAD")) {
		return "revert", nil
	}

	// MERGE_HEAD が存在すればマージ中
	if pathExists(filepath.Join(gitDir, "MERGE_HEAD")) {
		return "merge", nil
	}

	// どの操作も検出できない場合はエラーを返して引数による指定を促す
	return "", fmt.Errorf("中止できる操作が検出されませんでした。引数で操作を指定してください")
}

// abortOperation は指定された操作を実際に中止します
func abortOperation(operation string) error {
	// 実際の Git コマンドを実行する箇所。
	// gitcmd.RunWithIO は呼び出し元の標準入出力に接続してコマンドを実行するため、
	// ユーザー対話やエラー出力がそのまま端末に表示されます。
	switch operation {
	case "merge":
		return gitcmd.RunWithIO("merge", "--abort")
	case "rebase":
		return gitcmd.RunWithIO("rebase", "--abort")
	case "cherry-pick":
		return gitcmd.RunWithIO("cherry-pick", "--abort")
	case "revert":
		return gitcmd.RunWithIO("revert", "--abort")
	default:
		// 想定外の操作名が来た場合は明示的にエラーを返す
		return fmt.Errorf("未対応の操作です: %s", operation)
	}
}

// abortOperationLabel は日本語の表示名を返します
func abortOperationLabel(operation string) string {
	// 表示用に日本語ラベルを返すヘルパー関数
	// switch 文で対応する日本語を返し、未対応の文字列はそのまま返します。
	switch operation {
	case "merge":
		return "マージ"
	case "rebase":
		return "リベース"
	case "cherry-pick":
		return "チェリーピック"
	case "revert":
		return "リバート"
	default:
		return operation
	}
}

// getGitDir は現在のリポジトリの .git ディレクトリへの絶対パスを返します
func getGitDir() (string, error) {
	// git rev-parse --git-dir はリポジトリの .git ディレクトリのパスを返します。
	// - 絶対パスが返る場合と相対パスが返る場合がある（サブモジュール等）ため、
	//   相対パスだった場合はカレントディレクトリと結合して絶対パスに直します。
	output, err := gitcmd.Run("rev-parse", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("Gitディレクトリの取得に失敗しました: %w", err)
	}

	// gitcmd.Run の返す値は出力（末尾に改行が含まれることがある）なので
	// strings.TrimSpace で余分な空白や改行を削除します。
	dir := strings.TrimSpace(string(output))

	// filepath.IsAbs で絶対パスかどうか判定します。
	// - 絶対パスならそのまま返す
	if filepath.IsAbs(dir) {
		return dir, nil
	}

	// 相対パスの場合は現在の作業ディレクトリを取得して結合します。
	// - os.Getwd は現在のカレントワーキングディレクトリの絶対パスを返します。
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("カレントディレクトリの取得に失敗しました: %w", err)
	}

	// filepath.Join で OS に依存しない形でパス結合
	return filepath.Join(cwd, dir), nil
}

// pathExists はファイルまたはディレクトリの存在を確認します
func pathExists(path string) bool {
	// 空文字列は存在しないとみなす
	if path == "" {
		return false
	}

	// os.Stat はファイル情報を返し、存在しない場合はエラーを返す。
	// - 存在する場合: err == nil
	// - 存在しない場合: err != nil（詳細を判定するには os.IsNotExist(err) を利用可能）
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	cmd.RootCmd.AddCommand(abortCmd)
}
