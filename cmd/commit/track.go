/*
Package commit は git の拡張コマンドのうち、コミット関連のコマンドを定義します。

このファイル (track.go) は、トラッキングブランチを設定するコマンドを提供します。
現在のブランチに対してリモートトラッキングブランチを設定し、
必要に応じてリモートブランチを作成します。

主な機能:
  - トラッキングブランチの設定
  - リモートブランチの自動作成
  - git push --set-upstream の自動実行

使用例:
  git track                    # origin/<現在のブランチ> をトラッキング
  git track upstream           # upstream/<現在のブランチ> をトラッキング
  git track origin feature-123 # origin/feature-123 をトラッキング
*/
package commit

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// trackCmd は現在のブランチにトラッキングブランチを設定するコマンドです。
// リモートブランチが存在しない場合は、自動的に作成します。
var trackCmd = &cobra.Command{
	Use:   "track [リモート名] [ブランチ名]",
	Short: "トラッキングブランチを設定",
	Long: `現在のブランチに対してトラッキングブランチを設定します。
リモートブランチが存在しない場合は、自動的に
git push --set-upstream を実行してリモートブランチを作成し、
トラッキング設定を行います。`,
	Example: `  git track                    # origin/<現在のブランチ> をトラッキング
  git track upstream           # upstream/<現在のブランチ> をトラッキング
  git track origin feature-123 # origin/feature-123 をトラッキング`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		// 現在のブランチ名を取得
		currentBranch, err := fetchCurrentBranch()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗しました: %w", err)
		}

		// リモート名を引数から取得、デフォルトは origin
		remote := "origin"
		if len(args) >= 1 {
			remote = args[0]
		}

		// ブランチ名を引数から取得、デフォルトは現在のブランチ名
		remoteBranch := currentBranch
		if len(args) >= 2 {
			remoteBranch = args[1]
		}

		// リモートブランチが存在するか確認
		remoteRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
		exists, err := checkRemoteRefExists(remoteRef)
		if err != nil {
			return fmt.Errorf("リモートブランチの確認に失敗しました: %w", err)
		}

		if !exists {
			// リモートブランチが存在しない場合は、プッシュして作成する
			fmt.Printf("リモートブランチ %s が見つかりません。\n", remoteRef)
			fmt.Printf("git push --set-upstream %s %s を実行します...\n\n", remote, remoteBranch)

			if err := gitcmd.RunWithIO("push", "--set-upstream", remote, remoteBranch); err != nil {
				return fmt.Errorf("プッシュに失敗しました: %w", err)
			}

			fmt.Printf("\nブランチ '%s' を '%s' にプッシュし、トラッキングブランチを設定しました。\n", currentBranch, remoteRef)
			return nil
		}

		// リモートブランチが存在する場合は、upstream を設定
		upstreamRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
		if err := gitcmd.RunWithIO("branch", "--set-upstream-to="+upstreamRef, currentBranch); err != nil {
			return fmt.Errorf("トラッキングブランチの設定に失敗しました: %w", err)
		}

		fmt.Printf("ブランチ '%s' のトラッキングブランチを '%s' に設定しました。\n", currentBranch, upstreamRef)
		return nil
	},
}

// fetchCurrentBranch は現在のブランチ名を取得します。
//
// 戻り値:
//   - string: 現在のブランチ名
//   - error: エラーが発生した場合はエラーオブジェクト
func fetchCurrentBranch() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// checkRemoteRefExists はリモート参照が存在するかを確認します。
//
// パラメータ:
//   ref: 確認するリモート参照（例: "origin/main"）
//
// 戻り値:
//   - bool: 存在する場合は true、存在しない場合は false
//   - error: エラーが発生した場合はエラーオブジェクト
func checkRemoteRefExists(ref string) (bool, error) {
	// refs/remotes/<ref> の形式で存在確認
	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/%s", ref))
	if err != nil {
		// 終了コード 1 は参照が見つからないことを示す
		if gitcmd.IsExitError(err, 1) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// init はコマンドの初期化を行います。
// trackCmd を RootCmd に登録することで、CLI から実行可能にします。
func init() {
	cmd.RootCmd.AddCommand(trackCmd)
}
