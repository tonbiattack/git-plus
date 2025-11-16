// ================================================================================
// helpers.go
// ================================================================================
// このファイルは release パッケージで使用されるヘルパー関数を提供します。
// ================================================================================

package release

import (
	"os/exec"
	"strings"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// checkGitHubCLIInstalled はGitHub CLIがインストールされているかを確認します。
//
// 戻り値:
//   - bool: インストールされている場合はtrue
func checkGitHubCLIInstalled() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// verifyTagExists は指定されたタグが存在するかを確認します。
//
// パラメータ:
//   - tag: 確認するタグ名
//
// 戻り値:
//   - error: タグが存在しない場合はエラー
func verifyTagExists(tag string) error {
	return gitcmd.RunQuiet("rev-parse", "--verify", tag)
}

// getLatestTag は最新のタグを取得します。
//
// 戻り値:
//   - string: 最新のタグ名
//   - error: エラーが発生した場合のエラー情報
func getLatestTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
