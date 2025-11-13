package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

var trackCmd = &cobra.Command{
	Use:   "track [リモート名] [ブランチ名]",
	Short: "トラッキングブランチを設定",
	Long: `現在のブランチに対してトラッキングブランチを設定します。
リモートブランチが存在しない場合は、自動的に
git push --set-upstream を実行してリモートブランチを作成し、
トラッキング設定を行います。`,
	Example: `  git-plus track                    # origin/<現在のブランチ> をトラッキング
  git-plus track upstream           # upstream/<現在のブランチ> をトラッキング
  git-plus track origin feature-123 # origin/feature-123 をトラッキング`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Printf("リモートブランチ %s が見つかりません。\n", remoteRef)
			fmt.Printf("git push --set-upstream %s %s を実行します...\n\n", remote, remoteBranch)

			if err := gitcmd.RunWithIO("push", "--set-upstream", remote, remoteBranch); err != nil {
				return fmt.Errorf("プッシュに失敗しました: %w", err)
			}

			fmt.Printf("\nブランチ '%s' を '%s' にプッシュし、トラッキングブランチを設定しました。\n", currentBranch, remoteRef)
			return nil
		}

		// upstream を設定
		upstreamRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
		if err := gitcmd.RunWithIO("branch", "--set-upstream-to="+upstreamRef, currentBranch); err != nil {
			return fmt.Errorf("トラッキングブランチの設定に失敗しました: %w", err)
		}

		fmt.Printf("ブランチ '%s' のトラッキングブランチを '%s' に設定しました。\n", currentBranch, upstreamRef)
		return nil
	},
}

func fetchCurrentBranch() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func checkRemoteRefExists(ref string) (bool, error) {
	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/%s", ref))
	if err != nil {
		if gitcmd.IsExitError(err, 1) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
