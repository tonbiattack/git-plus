package main

import (
	"fmt"
	"os"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// main は現在のブランチをリモートブランチと同期するメイン処理
//
// 処理フロー:
//  1. コマンドライン引数を解析（ヘルプ、--continue、--abort、ブランチ名）
//  2. git fetch origin を実行して最新の変更を取得
//  3. 指定されたブランチ（デフォルト: main/master）を自動判定
//  4. git rebase origin/<ブランチ> を実行してリモートと同期
//  5. コンフリクト発生時は解決方法を案内
//
// 使用するgitコマンド:
//   - git fetch origin: リモートリポジトリから最新の変更を取得
//   - git rebase origin/<ブランチ>: 指定ブランチにリベース
//   - git rebase --continue: コンフリクト解決後にリベースを続行
//   - git rebase --abort: リベースを中止して元の状態に戻す
//
// コマンドライン引数:
//   - -h: ヘルプメッセージを表示
//   - --continue: コンフリクト解決後にリベースを続行
//   - --abort: リベースを中止
//   - [ブランチ名]: 同期先のブランチ（省略時は main/master を自動判定）
//
// 実装の詳細:
//   - rebase を使用するため、履歴がきれいに保たれる
//   - コンフリクト発生時は .git/rebase-merge ディレクトリの存在を確認
//   - デフォルトブランチの検出は origin/main -> origin/master の順に試行
//
// 終了コード:
//   - 0: 正常終了（同期完了）
//   - 1: エラー発生（fetch失敗、rebase失敗、コンフリクト発生など）
//
// 注意事項:
//   - すでにプッシュ済みのコミットがある場合、force push が必要になる可能性がある
//   - 共有ブランチでの使用には注意が必要
func main() {
	// ヘルプオプションのチェック
	// -h が指定された場合、使い方を表示して終了
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		printHelp()
		return
	}

	// --continue オプションの処理
	// コンフリクト解決後にユーザーが手動で実行するコマンド
	// git rebase --continue を実行して、中断していたリベースを再開する
	if len(os.Args) > 1 && os.Args[1] == "--continue" {
		if err := continueRebase(); err != nil {
			fmt.Fprintf(os.Stderr, "rebase の続行に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("同期が完了しました。")
		return
	}

	// --abort オプションの処理
	// リベースを中止して、リベース前の状態に戻す
	// git rebase --abort を実行して、全ての変更を破棄する
	if len(os.Args) > 1 && os.Args[1] == "--abort" {
		if err := abortRebase(); err != nil {
			fmt.Fprintf(os.Stderr, "rebase の中止に失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("同期を中止しました。")
		return
	}

	// ターゲットブランチの決定
	// コマンドライン引数でブランチが指定された場合はそれを使用
	// 指定がない場合は main/master を自動判定
	targetBranch := ""
	if len(os.Args) > 1 {
		// 引数として渡されたブランチ名を使用
		// 例: git sync develop -> targetBranch = "develop"
		targetBranch = os.Args[1]
	} else {
		// 引数がない場合は origin/main または origin/master を自動判定
		// detectDefaultBranch() は git show-ref で存在確認を行う
		branch, err := detectDefaultBranch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "デフォルトブランチの検出に失敗しました: %v\n", err)
			os.Exit(1)
		}
		targetBranch = branch
	}

	// git fetch origin を実行
	// リモートリポジトリから最新の変更を取得（ローカルファイルは変更しない）
	// これにより origin/<ブランチ> が最新の状態に更新される
	fmt.Printf("origin から最新の変更を取得しています...\n")
	if err := gitcmd.RunWithIO("fetch", "origin"); err != nil {
		fmt.Fprintf(os.Stderr, "fetch に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// git rebase origin/<ブランチ> を実行
	// 現在のブランチのコミットを origin/<ブランチ> の最新コミットの上に再配置
	// これにより履歴が一直線に保たれ、マージコミットが作成されない
	remoteBranch := fmt.Sprintf("origin/%s", targetBranch)
	fmt.Printf("%s にリベースしています...\n", remoteBranch)
	if err := gitcmd.RunWithIO("rebase", remoteBranch); err != nil {
		// rebase 中にコンフリクトが発生した場合
		// .git/rebase-merge または .git/rebase-apply ディレクトリが存在する
		// この場合、ユーザーにコンフリクト解決方法を案内する
		if isRebaseInProgress() {
			fmt.Println("\nコンフリクトが発生しました。")
			fmt.Println("コンフリクトを解決した後、以下のコマンドを実行してください:")
			fmt.Println("  git sync --continue    # 同期を続行")
			fmt.Println("  git sync --abort       # 同期を中止")
			os.Exit(1)
		}
		// コンフリクト以外のエラー（例: ネットワークエラー、権限エラーなど）
		fmt.Fprintf(os.Stderr, "rebase に失敗しました: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("同期が完了しました。(%s)\n", remoteBranch)
}

// detectDefaultBranch は origin のデフォルトブランチ (main または master) を検出する
//
// 戻り値:
//   - string: 検出されたブランチ名（"main" または "master"）
//   - error: 両方のブランチが存在しない場合のエラー
//
// 使用するgitコマンド:
//   - git show-ref --verify --quiet refs/remotes/origin/main: origin/main の存在確認
//   - git show-ref --verify --quiet refs/remotes/origin/master: origin/master の存在確認
//
// 実装の詳細:
//   - まず origin/main の存在を確認（最近のリポジトリは main を使用）
//   - origin/main が存在しない場合、origin/master を確認（古いリポジトリ）
//   - 両方とも存在しない場合はエラーを返す
//   - --quiet オプションで出力を抑制
//   - --verify オプションで厳密な参照確認
func detectDefaultBranch() (string, error) {
	// origin/main の存在確認
	// GitHub などの新しいリポジトリではデフォルトブランチが main になっている
	if err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/remotes/origin/main"); err == nil {
		return "main", nil
	}

	// origin/master の存在確認
	// 古いリポジトリや一部のサービスではまだ master が使われている
	if err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/remotes/origin/master"); err == nil {
		return "master", nil
	}

	// どちらも見つからない場合はエラー
	// リモートブランチが存在しないか、fetch していない可能性がある
	return "", fmt.Errorf("origin/main も origin/master も見つかりませんでした")
}

// isRebaseInProgress は rebase が進行中かどうかをチェックする
//
// 戻り値:
//   - bool: rebase が進行中の場合 true、それ以外は false
//
// 実装の詳細:
//   - .git/rebase-merge または .git/rebase-apply ディレクトリの存在を確認
//   - rebase-merge: 通常の rebase 時に作成される一時ディレクトリ
//   - git rebase, git rebase -i などで使用される
//   - コンフリクト解決中はこのディレクトリが残る
//   - rebase-apply: git am や古い形式の rebase で使用される一時ディレクトリ
//   - パッチ適用時に使用される
//   - git am --continue, git rebase --continue で処理が再開される
//   - これらのディレクトリが存在する = rebase が進行中（コンフリクト等で中断している状態）
//   - os.Stat() でディレクトリの存在を確認（エラーが返らなければ存在する）
//
// 使用箇所:
//   - main関数内で rebase 失敗時にコンフリクトかどうかを判定
func isRebaseInProgress() bool {
	// .git/rebase-merge ディレクトリの存在確認
	// git rebase でコンフリクトが発生すると作成される
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return true
	}
	// .git/rebase-apply ディレクトリの存在確認
	// git am や古い形式の rebase で使用される
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return true
	}
	// どちらも存在しない場合は rebase は進行していない
	return false
}

// continueRebase は rebase を続行する
//
// 戻り値:
//   - error: git rebase --continue の実行エラー（成功時は nil）
//
// 使用するgitコマンド:
//   - git rebase --continue: コンフリクト解決後にリベースを再開
//
// 実装の詳細:
//   - ユーザーがコンフリクトを解決した後に実行される
//   - git add でコンフリクトファイルをステージングした後に実行する必要がある
//   - 全てのコンフリクトが解決されるまで、この処理を繰り返す
//   - RunWithIO() を使用してgitの出力をそのまま表示
//
// 使用箇所:
//   - main関数内で --continue オプションが指定された場合
func continueRebase() error {
	return gitcmd.RunWithIO("rebase", "--continue")
}

// abortRebase は rebase を中止する
//
// 戻り値:
//   - error: git rebase --abort の実行エラー（成功時は nil）
//
// 使用するgitコマンド:
//   - git rebase --abort: リベースを中止して元の状態に戻す
//
// 実装の詳細:
//   - リベース開始前の状態に完全に戻す
//   - .git/rebase-merge ディレクトリを削除
//   - HEAD を元のコミットに戻す
//   - コンフリクト解決中の変更は全て破棄される
//   - RunWithIO() を使用してgitの出力をそのまま表示
//
// 使用箇所:
//   - main関数内で --abort オプションが指定された場合
func abortRebase() error {
	return gitcmd.RunWithIO("rebase", "--abort")
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//   - 全てのコマンドラインオプションの説明
//   - 使用例（基本的な同期、ブランチ指定、コンフリクト処理）
//   - 内部動作の説明（fetch -> rebase の流れ）
//   - 注意事項（force push の必要性、共有ブランチでの注意点）
//
// 使用箇所:
//   - main関数内で -h オプションが指定された場合
func printHelp() {
	help := `git sync - 現在のブランチを最新のリモートブランチと同期

使い方:
  git sync [ブランチ名]      # 指定されたブランチと同期（デフォルト: main/master）
  git sync --continue       # コンフリクト解決後に続行
  git sync --abort          # 同期を中止

説明:
  現在のブランチを最新の origin/<ブランチ> と同期します。
  内部的に git rebase を使用するため、履歴がきれいに保たれます。

引数:
  [ブランチ名]              同期先のブランチ（省略時は main/master を自動判定）

オプション:
  --continue                コンフリクト解決後に rebase を続行
  --abort                   同期を中止して元の状態に戻す
  -h                        このヘルプを表示

例:
  git sync                  # origin/main (または origin/master) と同期
  git sync develop          # origin/develop と同期
  git sync --continue       # コンフリクト解決後に続行
  git sync --abort          # 同期を中止

内部動作:
  1. git fetch origin を実行
  2. 指定されたブランチ（デフォルト: main/master）を自動判定
  3. git rebase origin/<ブランチ> を実行
  4. コンフリクト発生時は修正方法を案内

注意:
  - rebase を使用するため、すでにプッシュ済みのコミットがある場合は
    force push が必要になる可能性があります
  - 共有ブランチでは使用に注意してください
`
	fmt.Print(help)
}
