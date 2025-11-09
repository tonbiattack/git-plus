package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// main はブランチを作成または再作成するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. ブランチ名の引数チェック
//  3. ブランチの存在確認
//  4. ブランチが存在する場合:
//     a. ユーザーに操作を選択させる（recreate/switch/cancel）
//     b. switch の場合: 既存ブランチに切り替えて終了
//     c. cancel の場合: 処理を中止して終了
//     d. recreate の場合: ブランチを削除してから作成（下に続く）
//  5. ブランチが存在しない、またはrecreateが選択された場合:
//     a. 既存ブランチを強制削除（存在しない場合はエラーを無視）
//     b. 新しいブランチを作成して切り替え
//
// 終了コード:
//  - 0: 正常終了（作成成功、切り替え成功、またはキャンセル）
//  - 1: エラー発生（引数不足、ブランチ確認失敗、作成/切り替え失敗など）
func main() {
	// -h オプションのチェック
	// コマンドライン引数に -h が含まれている場合はヘルプを表示して終了
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	// ブランチ名の引数チェック
	// 引数が指定されていない場合はエラーメッセージを表示して終了
	if len(os.Args) < 2 {
		fmt.Println("ブランチ名を指定してください。")
		os.Exit(1)
	}
	branch := os.Args[1]

	// 指定されたブランチが既に存在するかチェック
	exists, err := branchExists(branch)
	if err != nil {
		fmt.Println("ブランチの存在確認に失敗しました:", err)
		os.Exit(1)
	}

	// ブランチが既に存在する場合の処理
	if exists {
		// ユーザーに操作を選択させる（recreate/switch/cancel）
		action, err := askForAction(branch)
		if err != nil {
			fmt.Println("入力の読み込みに失敗しました:", err)
			os.Exit(1)
		}

		// キャンセルが選択された場合は終了
		if action == "cancel" {
			fmt.Println("処理を中止しました。")
			return
		}

		// 既存ブランチへの切り替えが選択された場合
		if action == "switch" {
			switchCmd := exec.Command("git", "checkout", branch)
			switchCmd.Stdout = os.Stdout // git の出力をそのまま表示
			switchCmd.Stderr = os.Stderr // git のエラー出力もそのまま表示
			if err := switchCmd.Run(); err != nil {
				fmt.Println("ブランチの切り替えに失敗しました:", err)
				os.Exit(1)
			}
			fmt.Printf("ブランチ %s に切り替えました。\n", branch)
			return
		}
		// action == "recreate" の場合は下に続く（ブランチを削除してから作成）
	}

	// 既存ブランチを強制削除
	// git branch -D で強制削除（未マージでも削除される）
	// ブランチが存在しない場合のエラーは無視
	delCmd := exec.Command("git", "branch", "-D", branch)
	delCmd.Stdout = os.Stdout
	delCmd.Stderr = os.Stderr
	if err := delCmd.Run(); err != nil && !isNotFound(err) {
		fmt.Println("ブランチの削除に失敗しました:", err)
		os.Exit(1)
	}

	// 新しいブランチを作成して切り替え
	// git checkout -b で新しいブランチを作成し、同時に切り替え
	createCmd := exec.Command("git", "checkout", "-b", branch)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		fmt.Println("ブランチ作成に失敗しました:", err)
		os.Exit(1)
	}

	fmt.Printf("ブランチ %s を作成しました。\n", branch)
}

// branchExists は指定されたブランチが存在するかチェックする
//
// git show-ref コマンドを使用して、ローカルブランチの存在を確認する。
//
// パラメータ:
//  - name: チェックするブランチ名（例: "feature/awesome", "main"）
//
// 戻り値:
//  - bool: true = ブランチが存在、false = ブランチが存在しない
//  - error: git コマンドの実行エラー（終了コード1はブランチ不在として扱う）
//
// 使用する git コマンド:
//  git show-ref --verify --quiet refs/heads/<branch-name>
//
// オプションの説明:
//  --verify: 完全一致する参照のみを確認
//  --quiet: 出力を抑制（終了コードのみで判定）
//
// 終了コードの意味:
//  0: ブランチが存在
//  1: ブランチが存在しない
//  その他: エラー発生
func branchExists(name string) (bool, error) {
	// refs/heads/<branch-name> 形式の参照名を作成
	ref := fmt.Sprintf("refs/heads/%s", name)
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)

	if err := cmd.Run(); err != nil {
		// 終了コードが1の場合はブランチが存在しない
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil // ブランチが存在しない（エラーではない）
			}
		}
		// その他のエラーはそのまま返す
		return false, err
	}
	// 終了コード0: ブランチが存在
	return true, nil
}

// askForAction はブランチが既に存在する場合にユーザーに操作を選択させる
//
// ユーザーに3つの選択肢を提示し、標準入力から回答を読み取る:
//  - [r]ecreate: ブランチを削除して作り直す
//  - [s]witch: 既存のブランチに切り替える
//  - [c]ancel: 処理を中止する
//
// パラメータ:
//  - branch: 既に存在するブランチ名
//
// 戻り値:
//  - string: ユーザーが選択した操作（"recreate", "switch", "cancel"）
//  - error: 入力読み取り時のエラー（EOF は "cancel" として扱う）
//
// 入力パターン:
//  "r" または "recreate" → "recreate"
//  "s" または "switch" → "switch"
//  "c" または "cancel" または空入力 → "cancel"
//  その他 → "cancel"
//  EOF → "cancel"
func askForAction(branch string) (string, error) {
	fmt.Printf("ブランチ %s は既に存在します。どうしますか？ [r]ecreate/[s]witch/[c]ancel (r/s/c): ", branch)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// EOF（Ctrl+D など）の場合はキャンセルとして扱う
		if err == io.EOF {
			input = "c"
		} else {
			return "", err
		}
	}

	// 入力を小文字に変換し、前後の空白（改行含む）を削除
	answer := strings.ToLower(strings.TrimSpace(input))
	switch answer {
	case "r", "recreate":
		return "recreate", nil
	case "s", "switch":
		return "switch", nil
	case "c", "cancel", "":
		return "cancel", nil
	default:
		// 不明な入力はキャンセルとして扱う
		return "cancel", nil
	}
}

// isNotFound はエラーがブランチ不在エラーかどうかを判定する
//
// git branch -D コマンドが失敗した際、ブランチが存在しないことによる
// エラー（終了コード1）かどうかを判定する。
//
// パラメータ:
//  - err: チェックするエラー
//
// 戻り値:
//  - bool: true = ブランチ不在エラー（終了コード1）、false = その他のエラー
//
// 使用例:
//  削除コマンドが失敗しても、ブランチが元々存在しない場合はエラーとして扱わない
func isNotFound(err error) bool {
	exitErr, ok := err.(*exec.ExitError)
	return ok && exitErr.ExitCode() == 1
}

// printHelp はコマンドのヘルプメッセージを表示する
//
// 使い方、説明、オプション、例を含む詳細なヘルプを標準出力に表示する。
// -h オプションが指定された場合や、使い方が分からない場合にユーザーに情報を提供する。
func printHelp() {
	help := `git newbranch - ブランチを作成または再作成

使い方:
  git newbranch <ブランチ名>

説明:
  指定したブランチ名でブランチを作成します。
  既にブランチが存在する場合は、以下の選択肢が表示されます：
    [r]ecreate - ブランチを削除して作り直す
    [s]witch   - 既存のブランチに切り替える
    [c]ancel   - 処理を中止する

オプション:
  -h                    このヘルプを表示

例:
  git newbranch feature/awesome    # feature/awesome ブランチを作成
  git newbranch -h                 # ヘルプを表示

実装詳細:
  ブランチの存在確認: git show-ref --verify --quiet refs/heads/<branch-name>
  ブランチの削除: git branch -D <branch-name>
  ブランチの作成と切り替え: git checkout -b <branch-name>
`
	fmt.Print(help)
}
