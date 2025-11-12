package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// StashInfo はスタッシュの詳細情報を表す構造体
//
// フィールド:
//   - Index: スタッシュのインデックス番号。例: stash@{0} の場合は 0
//   - Name: スタッシュの名前。例: "stash@{0}"
//   - Files: スタッシュに含まれるファイルのリスト。例: []string{"main.go", "README.md"}
//   - Hash: スタッシュの内容から計算されたハッシュ値（重複検出用）
//
// 使用箇所:
//   - main: スタッシュ情報の収集と重複検出
//   - findDuplicates: 重複グループの作成
type StashInfo struct {
	Index int
	Name  string
	Files []string
	Hash  string
}

// main は重複するスタッシュを検出して削除するメイン処理
//
// 処理フロー:
//  1. ヘルプオプション(-h)のチェック
//  2. 全スタッシュの一覧を取得
//  3. 各スタッシュの詳細情報（ファイル一覧、内容ハッシュ）を取得
//  4. 内容が完全に同一のスタッシュグループを検出
//  5. 重複グループを表示
//  6. ユーザー確認後、各グループの最新以外を削除
//  7. 削除結果のサマリーを表示
//
// 使用するgitコマンド:
//   - git stash list: スタッシュ一覧を取得
//   - git stash show --name-only <stash>: スタッシュのファイル一覧を取得
//   - git show <stash>:<file>: スタッシュの特定ファイルの内容を取得
//   - git hash-object --stdin: 内容からハッシュ値を計算
//   - git stash drop stash@{N}: 指定インデックスのスタッシュを削除
//
// 実装の詳細:
//   - ファイル構成と内容の両方が同一のスタッシュを重複とみなす
//   - 重複グループ内では最新（インデックスが最小）のみを保持
//   - インデックスが大きい順に削除してインデックスのずれを防止
//
// 終了コード:
//   - 0: 正常終了（削除成功、スタッシュなし、重複なし、キャンセル）
//   - 1: エラー発生（スタッシュ取得失敗など）
func main() {
	// -h オプションのチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			printHelp()
			return
		}
	}

	fmt.Println("スタッシュを分析しています...")

	// 全スタッシュの一覧を取得
	stashes, err := getAllStashes()
	if err != nil {
		fmt.Printf("エラー: スタッシュ一覧の取得に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if len(stashes) == 0 {
		fmt.Println("スタッシュが存在しません。")
		os.Exit(0)
	}

	if len(stashes) == 1 {
		fmt.Println("スタッシュが1つしかないため、圧縮の必要はありません。")
		os.Exit(0)
	}

	fmt.Printf("合計 %d 個のスタッシュが見つかりました。\n\n", len(stashes))

	// 各スタッシュの詳細情報を取得
	stashInfos := make([]StashInfo, 0, len(stashes))
	for i, stashName := range stashes {
		files, err := getStashFiles(stashName)
		if err != nil {
			fmt.Printf("警告: %s のファイル一覧取得に失敗しました: %v\n", stashName, err)
			continue
		}

		hash, err := getStashContentHash(stashName, files)
		if err != nil {
			fmt.Printf("警告: %s の内容ハッシュ取得に失敗しました: %v\n", stashName, err)
			continue
		}

		stashInfos = append(stashInfos, StashInfo{
			Index: i,
			Name:  stashName,
			Files: files,
			Hash:  hash,
		})
	}

	// 重複を検出（ファイル構成と内容が同じもの）
	duplicateGroups := findDuplicates(stashInfos)

	if len(duplicateGroups) == 0 {
		fmt.Println("✓ 重複するスタッシュは見つかりませんでした。")
		os.Exit(0)
	}

	// 重複を表示
	fmt.Printf("重複するスタッシュグループが %d 個見つかりました:\n\n", len(duplicateGroups))

	totalToDelete := 0
	for groupIdx, group := range duplicateGroups {
		fmt.Printf("グループ %d: %d 個の重複\n", groupIdx+1, len(group))
		for _, info := range group {
			fmt.Printf("  - %s (%d ファイル)\n", info.Name, len(info.Files))
		}
		totalToDelete += len(group) - 1 // 最新以外を削除
		fmt.Println()
	}

	fmt.Printf("合計 %d 個のスタッシュを削除します（各グループの最新のみを保持）。\n", totalToDelete)

	// 破壊的操作なのでEnterでno
	if !ui.Confirm("続行しますか?", false) {
		fmt.Println("キャンセルしました。")
		os.Exit(0)
	}

	// 削除実行（インデックスが大きい方から削除して、インデックスのずれを防ぐ）
	fmt.Println("\n重複スタッシュを削除しています...")
	deletedCount := 0
	failedCount := 0

	// 削除対象のスタッシュを収集
	toDelete := make([]StashInfo, 0)
	for _, group := range duplicateGroups {
		// グループ内で最新（インデックスが最小）を保持
		// インデックス1以降を削除対象に追加
		for i := 1; i < len(group); i++ {
			toDelete = append(toDelete, group[i])
		}
	}

	// インデックスの大きい順（古い順）にソートして削除
	sort.Slice(toDelete, func(i, j int) bool {
		return toDelete[i].Index > toDelete[j].Index
	})

	// 削除実行
	for _, stashToDelete := range toDelete {
		if err := deleteStash(stashToDelete.Index); err != nil {
			fmt.Printf("✗ %s の削除に失敗しました: %v\n", stashToDelete.Name, err)
			failedCount++
		} else {
			fmt.Printf("✓ %s を削除しました\n", stashToDelete.Name)
			deletedCount++
		}
	}

	// 結果サマリー
	fmt.Printf("\n完了: %d 個のスタッシュを削除しました", deletedCount)
	if failedCount > 0 {
		fmt.Printf(" (%d 個失敗)", failedCount)
	}
	fmt.Println()

	// 残りのスタッシュ数を表示
	remainingStashes, _ := getAllStashes()
	fmt.Printf("残りのスタッシュ数: %d\n", len(remainingStashes))
}

// getAllStashes は全スタッシュの名前リストを取得する
//
// 戻り値:
//   - []string: スタッシュ名のリスト。例: []string{"stash@{0}", "stash@{1}"}
//   - error: エラー情報（git stash list の実行失敗など）
//
// 使用するgitコマンド:
//   - git stash list: すべてのスタッシュを一覧表示
//
// 実装の詳細:
//   - 出力形式: "stash@{N}: WIP on branch: message"
//   - ":" で分割して stash@{N} 部分のみを抽出
//   - 空行は無視
func getAllStashes() ([]string, error) {
	output, err := gitcmd.Run("stash", "list")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	stashes := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// "stash@{0}: WIP on branch: message" から "stash@{0}" を抽出
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 0 {
			stashes = append(stashes, strings.TrimSpace(parts[0]))
		}
	}

	return stashes, nil
}

// getStashFiles はstashに含まれるファイルのリストを取得する
//
// パラメータ:
//   - stash: スタッシュ名。例: "stash@{0}"
//
// 戻り値:
//   - []string: ファイル名のリスト（ソート済み）
//   - error: エラー情報（git stash show の実行失敗など）
//
// 使用するgitコマンド:
//   - git stash show --name-only <stash>: スタッシュ内のファイル名のみを表示
//
// 実装の詳細:
//   - ファイル名をアルファベット順にソート（重複比較の一貫性のため）
//   - 空行は無視
func getStashFiles(stash string) ([]string, error) {
	output, err := gitcmd.Run("stash", "show", "--name-only", "-u", stash)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	seen := make(map[string]struct{})
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, exists := seen[line]; exists {
			continue
		}
		seen[line] = struct{}{}
		files = append(files, line)
	}

	sort.Strings(files)
	return files, nil
}

// getStashContentHash はスタッシュの内容からハッシュを生成する
//
// パラメータ:
//   - stash: スタッシュ名。例: "stash@{0}"
//   - files: スタッシュに含まれるファイルのリスト（ソート済み）
//
// 戻り値:
//   - string: SHA-1ハッシュ値（40文字の16進数文字列）
//   - error: エラー情報（ハッシュ計算失敗など）
//
// 使用するgitコマンド:
//   - git show <stash>:<file>: スタッシュの特定ファイルの内容を取得
//   - git hash-object --stdin: 標準入力からSHA-1ハッシュを計算
//
// 実装の詳細:
//   - ファイル名一覧と各ファイルの内容を連結してハッシュ化
//   - ファイル取得エラーは "ERROR:<ファイル名>" として内容に含める
//   - 同じ内容のスタッシュは同じハッシュ値になる
func getStashContentHash(stash string, files []string) (string, error) {
	var buffer bytes.Buffer

	buffer.WriteString(strings.Join(files, "\n"))
	buffer.WriteString("\n---\n")

	components := []struct {
		label string
		ref   string
	}{
		{label: "WORKTREE", ref: stash},
		{label: "INDEX", ref: fmt.Sprintf("%s^2", stash)},
		{label: "UNTRACKED", ref: fmt.Sprintf("%s^3", stash)},
	}

	for _, file := range files {
		buffer.WriteString("FILE:" + file + "\n")
		for _, component := range components {
			content, found, err := getStashFileContent(component.ref, file)
			if err != nil {
				return "", err
			}
			if !found {
				continue
			}

			buffer.WriteString("PART:" + component.label + "\n")
			buffer.Write(content)
			if len(content) == 0 || content[len(content)-1] != '\n' {
				buffer.WriteString("\n")
			}
		}
	}

	cmd := exec.Command("git", "hash-object", "--stdin")
	cmd.Stdin = &buffer
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getStashFileContent はstashの特定のファイルの内容を取得する
//
// パラメータ:
//   - stash: スタッシュ名。例: "stash@{0}"
//   - file: ファイルパス。例: "main.go"
//
// 戻り値:
//   - []byte: ファイルの内容
//   - error: エラー情報（git show の実行失敗など）
//
// 使用するgitコマンド:
//   - git show <stash>:<file>: スタッシュの特定ファイルの内容を表示
//
// 実装の詳細:
//   - ref形式: "stash@{0}:main.go"
//   - バイナリファイルも取得可能
func getStashFileContent(stashRef, file string) ([]byte, bool, error) {
	if stashRef == "" {
		return nil, false, nil
	}

	ref := fmt.Sprintf("%s:%s", stashRef, file)
	output, err := gitcmd.Run("show", ref)
	if err != nil {
		if gitcmd.IsExitError(err, 128) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return output, true, nil
}

// findDuplicates は重複するスタッシュグループを検出する
//
// パラメータ:
//   - stashInfos: すべてのスタッシュ情報のスライス
//
// 戻り値:
//   - [][]StashInfo: 重複するスタッシュのグループのスライス
//
// 実装の詳細:
//   - ハッシュ値が同じスタッシュをグループ化
//   - 2個以上のグループのみを重複として返す
//   - 各グループ内はインデックス順（小さい=新しい）にソート
func findDuplicates(stashInfos []StashInfo) [][]StashInfo {
	// make(map[string][]StashInfo) の文法解説:
	// - map[キー型]値型: マップ（連想配列）の型定義
	// - map[string][]StashInfo: キーがstring、値が[]StashInfo（StashInfoのスライス）
	// - make(): マップを初期化して使用可能な状態にする（nilでない空マップを作成）
	// 用途: 同じハッシュ値を持つスタッシュをグループ化するため
	//       例: {"abc123": [stash1, stash2], "def456": [stash3]}
	hashMap := make(map[string][]StashInfo)

	// ハッシュでグループ化
	for _, info := range stashInfos {
		hashMap[info.Hash] = append(hashMap[info.Hash], info)
	}

	// 重複があるグループのみを抽出
	duplicateGroups := make([][]StashInfo, 0)
	for _, group := range hashMap {
		if len(group) > 1 {
			// インデックス順にソート（小さい=新しい）
			// sort.Slice文法: sort.Slice(スライス, 比較関数)
			// - 第1引数: ソート対象のスライス
			// - 第2引数: 無名関数 func(i, j int) bool
			//   - i, j: 比較する2要素のインデックス
			//   - 戻り値: i番目がj番目より前に来るべき場合にtrueを返す
			//   - group[i].Index < group[j].Index で昇順ソート（小→大）
			sort.Slice(group, func(i, j int) bool {
				return group[i].Index < group[j].Index
			})
			duplicateGroups = append(duplicateGroups, group)
		}
	}

	return duplicateGroups
}

// deleteStash は指定したインデックスのスタッシュを削除する
//
// パラメータ:
//   - index: スタッシュのインデックス番号。例: 0 は stash@{0} を指す
//
// 戻り値:
//   - error: エラー情報（git stash drop の実行失敗など）
//
// 使用するgitコマンド:
//   - git stash drop stash@{N}: 指定インデックスのスタッシュを削除
//
// 実装の詳細:
//   - インデックスから stash@{N} 形式の参照を生成
//   - 削除失敗時はエラーメッセージを含めて返す
func deleteStash(index int) error {
	// git stash drop stash@{N}
	err := gitcmd.RunQuiet("stash", "drop", fmt.Sprintf("stash@{%d}", index))
	if err != nil {
		return err
	}

	return nil
}

// printHelp はヘルプメッセージを表示する
//
// 実装の詳細:
//   - 重複スタッシュの検出と削除機能の使い方を説明
//   - 重複の判定基準（ファイル構成と内容）を明記
//   - 削除前の確認プロセスを説明
func printHelp() {
	help := `git stash-cleanup - 重複するスタッシュを検出して削除

使い方:
  git stash-cleanup

説明:
  全てのスタッシュを分析し、ファイル構成と内容が完全に同一の
  スタッシュを検出します。重複するスタッシュをグループ化して表示し、
  削除確認のプロンプトを表示します。
  各重複グループから最新のスタッシュ（インデックスが最小）のみを残し、
  古い重複を削除します。

オプション:
  -h                    このヘルプを表示

注意:
  - 削除前に確認プロンプトが表示されます
  - y または yes を入力すると削除が実行されます
  - 各重複グループで最新のもののみが残ります
`
	fmt.Print(help)
}
