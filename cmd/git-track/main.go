package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// main は現在のブランチにトラッキングブランチを設定するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. 現在のブランチ名を取得
//  3. コマンドライン引数からリモート名とブランチ名を取得（デフォルト: origin、現在のブランチ名）
//  4. リモートブランチの存在を確認
//  5a. リモートブランチが存在しない場合: git push --set-upstream を実行
//  5b. リモートブランチが存在する場合: git branch --set-upstream-to を実行
//
// 使用するgitコマンド:
//  - git rev-parse --abbrev-ref HEAD: 現在のブランチ名を取得
//  - git show-ref --verify refs/remotes/<リモート>/<ブランチ>: リモートブランチの存在確認
//  - git push --set-upstream <リモート> <ブランチ>: リモートブランチを作成してトラッキング設定
//  - git branch --set-upstream-to=<リモート>/<ブランチ>: トラッキングブランチを設定
//
// 実装の詳細:
//  - デフォルトリモート: origin
//  - デフォルトブランチ: 現在のブランチ名
//  - リモートブランチがない場合は自動的にプッシュしてトラッキング設定
//
// 終了コード:
//  - 0: 正常終了（トラッキング設定成功）
//  - 1: エラー発生（ブランチ名取得失敗、リモート確認失敗、設定失敗）
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// 現在のブランチ名を取得
	currentBranch, err := getCurrentBranch()
	if err != nil {
		fmt.Println("現在のブランチの取得に失敗しました:", err)
		os.Exit(1)
	}

	// リモート名を引数から取得、デフォルトは origin
	remote := "origin"
	if len(os.Args) >= 2 {
		remote = os.Args[1]
	}

	// ブランチ名を引数から取得、デフォルトは現在のブランチ名
	remoteBranch := currentBranch
	if len(os.Args) >= 3 {
		remoteBranch = os.Args[2]
	}

	// リモートブランチが存在するか確認
	remoteRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
	exists, err := remoteRefExists(remoteRef)
	if err != nil {
		fmt.Println("リモートブランチの確認に失敗しました:", err)
		os.Exit(1)
	}

	if !exists {
		fmt.Printf("リモートブランチ %s が見つかりません。\n", remoteRef)
		fmt.Printf("git push --set-upstream %s %s を実行します...\n\n", remote, remoteBranch)

		// git push --set-upstream を実行
		if err := gitcmd.RunWithIO("push", "--set-upstream", remote, remoteBranch); err != nil {
			fmt.Println("\nプッシュに失敗しました:", err)
			os.Exit(1)
		}

		fmt.Printf("\nブランチ '%s' を '%s' にプッシュし、トラッキングブランチを設定しました。\n", currentBranch, remoteRef)
		return
	}

	// upstream を設定
	upstreamRef := fmt.Sprintf("%s/%s", remote, remoteBranch)
	if err := gitcmd.RunWithIO("branch", "--set-upstream-to="+upstreamRef, currentBranch); err != nil {
		fmt.Println("トラッキングブランチの設定に失敗しました:", err)
		os.Exit(1)
	}

	fmt.Printf("ブランチ '%s' のトラッキングブランチを '%s' に設定しました。\n", currentBranch, upstreamRef)
}

// getCurrentBranch は現在チェックアウトされているブランチ名を取得する
//
// 戻り値:
//  - string: ブランチ名。例: "main", "feature/new-feature"
//  - error: エラー情報（git rev-parse の実行失敗など）
//
// 使用するgitコマンド:
//  - git rev-parse --abbrev-ref HEAD: 現在のブランチ名を短縮形式で取得
//
// 実装の詳細:
//  - detached HEAD の場合は "HEAD" が返される
//  - 出力から改行を除去
func getCurrentBranch() (string, error) {
	output, err := gitcmd.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// remoteRefExists はリモート参照が存在するかを確認する
//
// パラメータ:
//  - ref: リモート参照名。例: "origin/main"
//
// 戻り値:
//  - bool: true（存在する）、false（存在しない）
//  - error: エラー情報（git show-ref の実行失敗など、終了コード1以外のエラー）
//
// 使用するgitコマンド:
//  - git show-ref --verify --quiet refs/remotes/<ref>: リモート参照の存在を確認
//
// 実装の詳細:
//  - --quiet: 標準出力を抑制
//  - --verify: 参照の存在確認のみ
//  - 終了コード0: 存在する、1: 存在しない、その他: エラー
func remoteRefExists(ref string) (bool, error) {
	err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/%s", ref))
	if err != nil {
		if gitcmd.IsExitError(err, 1) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//  - トラッキングブランチ設定機能の使い方を説明
//  - リモートブランチがない場合の自動プッシュ機能を明記
//  - 使用例を複数パターン提示
func printHelp() {
	help := `git track - トラッキングブランチを設定

使い方:
  git track                    # origin/<現在のブランチ名> をトラッキング
  git track <リモート名>       # <リモート名>/<現在のブランチ名> をトラッキング
  git track <リモート名> <ブランチ名>  # <リモート名>/<ブランチ名> をトラッキング

説明:
  現在のブランチに対してトラッキングブランチを設定します。
  リモートブランチが存在しない場合は、自動的に
  git push --set-upstream を実行してリモートブランチを作成し、
  トラッキング設定を行います。

オプション:
  -h                    このヘルプを表示

例:
  git track                    # origin/<現在のブランチ> をトラッキング
  git track upstream           # upstream/<現在のブランチ> をトラッキング
  git track origin feature-123 # origin/feature-123 をトラッキング

注意:
  - リモートブランチがない場合は自動でプッシュされます
  - git pull 実行時のトラッキング情報エラーを解決できます
`
	fmt.Print(help)
}
