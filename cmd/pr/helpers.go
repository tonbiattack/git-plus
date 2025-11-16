// ================================================================================
// helpers.go
// ================================================================================
// このファイルは pr パッケージで使用されるヘルパー関数を提供します。
// ================================================================================

package pr

import (
	"os/exec"
	"strings"
)

// getBranchCurrent は現在チェックアウトされているブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名（空白や改行は除去されます）
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git branch --show-current コマンドを実行してブランチ名を取得します。
func getBranchCurrent() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkUncommittedChanges は未コミットの変更があるかどうかを確認します。
//
// 戻り値:
//   - bool: 未コミットの変更がある場合は true、ない場合は false
//   - error: git コマンドの実行に失敗した場合のエラー情報
//
// 内部処理:
//   git status --porcelain コマンドを実行し、出力があるかどうかで判定します。
//   変更がない場合は空の出力が返されます。
func checkUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// createStashWithMessage は指定されたメッセージで stash を作成します。
//
// パラメータ:
//   - message: stash に付けるメッセージ
//
// 戻り値:
//   - string: 作成された stash の参照（SHA-1 ハッシュ）
//   - error: stash の作成に失敗した場合のエラー情報
//
// 内部処理:
//   1. git stash push -m "<メッセージ>" で変更を stash に保存
//   2. git rev-parse stash@{0} で最新の stash の参照を取得
func createStashWithMessage(message string) (string, error) {
	cmd := exec.Command("git", "stash", "push", "-m", message)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	cmd = exec.Command("git", "rev-parse", "stash@{0}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
