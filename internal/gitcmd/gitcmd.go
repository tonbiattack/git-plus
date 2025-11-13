// ================================================================================
// Package gitcmd - Gitコマンド実行ユーティリティ
// ================================================================================
// このパッケージは、Gitコマンドを実行するための共通ユーティリティ関数を提供します。
//
// 提供する機能:
// - Run(): Gitコマンドを実行して出力を取得
// - RunWithIO(): Gitコマンドを実行して標準入出力をリダイレクト
// - RunQuiet(): Gitコマンドを静かに実行（出力なし）
// - IsExitError(): Gitコマンドの終了コードをチェック
//
// 使用目的:
// すべてのサブコマンドで共通して使用するGit操作を一元化し、
// コードの重複を避けてメンテナンス性を向上させます。
//
// 設計思想:
// - シンプルなインターフェース: exec.Commandのラッパーとして機能
// - エラーハンドリング: 呼び出し側でエラーを適切に処理できるよう設計
// - 柔軟性: 出力の取得、リダイレクト、抑制など複数のモードを提供
// ================================================================================
package gitcmd

import (
	"os"
	"os/exec"
)

// Run は指定された引数で git コマンドを実行し、出力を返す
//
// パラメータ:
//   - args: git コマンドの引数（例: "branch", "--show-current"）
//
// 戻り値:
//   - []byte: コマンドの標準出力
//   - error: コマンド実行エラー
//
// 使用例:
//
//	output, err := gitcmd.Run("branch", "--show-current")
//	if err != nil {
//	    return err
//	}
//	branch := string(output)
func Run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	return cmd.Output()
}

// RunWithIO は指定された引数で git コマンドを実行し、
// 標準入出力・標準エラー出力を親プロセスにリダイレクトする
//
// この関数は、git コマンドの出力をリアルタイムでユーザーに表示したい場合や、
// ユーザーからの入力が必要な場合（例: エディタの起動、コンフリクト解決）に使用します。
//
// パラメータ:
//   - args: git コマンドの引数（例: "rebase", "--continue"）
//
// 戻り値:
//   - error: コマンド実行エラー
//
// 使用例:
//
//	err := gitcmd.RunWithIO("rebase", "origin/main")
//	if err != nil {
//	    return err
//	}
func RunWithIO(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunQuiet は git コマンドを静かに実行する（出力なし）
//
// コマンドの成功/失敗のみを確認したい場合に使用します。
// 標準出力・標準エラー出力は破棄されます。
//
// パラメータ:
//   - args: git コマンドの引数（例: "show-ref", "--verify", "--quiet", "refs/heads/main"）
//
// 戻り値:
//   - error: コマンド実行エラー（終了コードが0以外の場合）
//
// 使用例:
//
//	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/heads/main")
//	if err != nil {
//	    // ブランチが存在しない
//	}
func RunQuiet(args ...string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

// IsExitError はエラーが特定の終了コードを持つ終了エラーかどうかをチェックする
//
// git コマンドの終了コードをチェックする場合に使用します。
//
// パラメータ:
//   - err: チェックするエラー
//   - code: 期待する終了コード
//
// 戻り値:
//   - bool: true = 指定した終了コードのエラー、false = それ以外
//
// 使用例:
//
//	if gitcmd.IsExitError(err, 1) {
//	    // 終了コード1のエラー（例: ブランチが存在しない）
//	}
func IsExitError(err error, code int) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode() == code
	}
	return false
}
