// ================================================================================
// stash_cleanup.go
// ================================================================================
// このファイルは git の拡張コマンド stash-cleanup コマンドを実装しています。
//
// 【概要】
// stash-cleanup コマンドは、重複するスタッシュを検出して削除する機能を提供します。
// ファイル構成と内容が完全に同一のスタッシュをグループ化し、古い重複を削除します。
//
// 【主な機能】
// - 全スタッシュの詳細な分析
// - ファイル構成と内容に基づく重複検出
// - 重複スタッシュのグループ化と表示
// - 各グループから最新のスタッシュ（インデックスが最小）のみを保持
// - 古い重複スタッシュの自動削除
// - 削除前の確認プロンプト
//
// 【使用例】
//   git stash-cleanup  # 重複スタッシュを検出して削除
//
// 【重複判定の仕組み】
// 1. 各スタッシュのファイル一覧を取得
// 2. 各ファイルの内容（worktree, index, untracked）を取得
// 3. ファイル一覧と内容を結合してハッシュ値を計算
// 4. 同じハッシュ値を持つスタッシュを重複と判定
// ================================================================================

package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
	"github.com/tonbiattack/git-plus/internal/ui"
)

// StashInfo はスタッシュの詳細情報を表す構造体です。
type StashInfo struct {
	Index int      // スタッシュのインデックス（stash@{N} の N）
	Name  string   // スタッシュの名前（例: stash@{0}）
	Files []string // スタッシュに含まれるファイルのパス一覧
	Hash  string   // スタッシュの内容を表すハッシュ値（重複判定に使用）
}

// stashCleanupCmd は stash-cleanup コマンドの定義です。
// 重複するスタッシュを検出して削除します。
var stashCleanupCmd = &cobra.Command{
	Use:   "stash-cleanup",
	Short: "重複するスタッシュを検出して削除",
	Long: `全てのスタッシュを分析し、ファイル構成と内容が完全に同一の
スタッシュを検出します。重複するスタッシュをグループ化して表示し、
削除確認のプロンプトを表示します。
各重複グループから最新のスタッシュ（インデックスが最小）のみを残し、
古い重複を削除します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("スタッシュを分析しています...")

		// 全スタッシュの一覧を取得
		stashes, err := getAllStashesList()
		if err != nil {
			return fmt.Errorf("スタッシュ一覧の取得に失敗しました: %w", err)
		}

		if len(stashes) == 0 {
			fmt.Println("スタッシュが存在しません。")
			return nil
		}

		if len(stashes) == 1 {
			fmt.Println("スタッシュが1つしかないため、圧縮の必要はありません。")
			return nil
		}

		fmt.Printf("合計 %d 個のスタッシュが見つかりました。\n\n", len(stashes))

		// 各スタッシュの詳細情報を取得
		stashInfos := make([]StashInfo, 0, len(stashes))
		for i, stashName := range stashes {
			files, err := getStashFilesList(stashName)
			if err != nil {
				fmt.Printf("警告: %s のファイル一覧取得に失敗しました: %v\n", stashName, err)
				continue
			}

			hash, err := getStashHash(stashName, files)
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

		// 重複を検出
		duplicateGroups := findDuplicateStashes(stashInfos)

		if len(duplicateGroups) == 0 {
			fmt.Println("✓ 重複するスタッシュは見つかりませんでした。")
			return nil
		}

		// 重複を表示
		fmt.Printf("重複するスタッシュグループが %d 個見つかりました:\n\n", len(duplicateGroups))

		totalToDelete := 0
		for groupIdx, group := range duplicateGroups {
			fmt.Printf("グループ %d: %d 個の重複\n", groupIdx+1, len(group))
			for _, info := range group {
				fmt.Printf("  - %s (%d ファイル)\n", info.Name, len(info.Files))
			}
			totalToDelete += len(group) - 1
			fmt.Println()
		}

		fmt.Printf("合計 %d 個のスタッシュを削除します（各グループの最新のみを保持）。\n", totalToDelete)

		if !ui.Confirm("続行しますか?", false) {
			fmt.Println("キャンセルしました。")
			return nil
		}

		// 削除実行
		fmt.Println("\n重複スタッシュを削除しています...")
		deletedCount := 0
		failedCount := 0

		// 削除対象のスタッシュを収集
		toDelete := make([]StashInfo, 0)
		for _, group := range duplicateGroups {
			for i := 1; i < len(group); i++ {
				toDelete = append(toDelete, group[i])
			}
		}

		// インデックスの大きい順にソート
		sort.Slice(toDelete, func(i, j int) bool {
			return toDelete[i].Index > toDelete[j].Index
		})

		// 削除実行
		for _, stashToDelete := range toDelete {
			if err := deleteStashByIndex(stashToDelete.Index); err != nil {
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
		remainingStashes, _ := getAllStashesList()
		fmt.Printf("残りのスタッシュ数: %d\n", len(remainingStashes))

		return nil
	},
}

// getAllStashesList は全てのスタッシュの名前一覧を取得します。
//
// 戻り値:
//   - []string: スタッシュ名のスライス（例: ["stash@{0}", "stash@{1}", ...]）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git stash list コマンドを実行し、各行から stash@{N} の形式で
//   スタッシュ名を抽出します。
func getAllStashesList() ([]string, error) {
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
		parts := strings.SplitN(line, ":", 2)
		if len(parts) > 0 {
			stashes = append(stashes, strings.TrimSpace(parts[0]))
		}
	}

	return stashes, nil
}

// getStashFilesList は指定されたスタッシュに含まれるファイル一覧を取得します。
//
// パラメータ:
//   - stash: スタッシュ名（例: stash@{0}）
//
// 戻り値:
//   - []string: ファイルパスのスライス（重複を除去し、ソート済み）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   git stash show --name-only -u <stash> コマンドでファイル一覧を取得します。
//   -u オプションにより untracked ファイルも含めます。
func getStashFilesList(stash string) ([]string, error) {
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

// getStashHash はスタッシュの内容を表すハッシュ値を計算します。
//
// パラメータ:
//   - stash: スタッシュ名（例: stash@{0}）
//   - files: スタッシュに含まれるファイルのパス一覧
//
// 戻り値:
//   - string: 計算されたハッシュ値（SHA-1）
//   - error: エラーが発生した場合のエラー情報
//
// 内部処理:
//   1. ファイル一覧を文字列として連結
//   2. 各ファイルについて以下の3つの状態の内容を取得:
//      - WORKTREE: 作業ディレクトリの状態（stash）
//      - INDEX: ステージングエリアの状態（stash^2）
//      - UNTRACKED: 追跡されていないファイル（stash^3）
//   3. 全ての内容を連結してバッファを作成
//   4. git hash-object --stdin でハッシュ値を計算
//
// 備考:
//   この方法により、ファイル構成と内容が完全に同一のスタッシュは
//   同じハッシュ値を持つため、重複判定が可能になります。
func getStashHash(stash string, files []string) (string, error) {
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
			content, found, err := getStashFileContentByRef(component.ref, file)
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

// getStashFileContentByRef は指定された参照からファイルの内容を取得します。
//
// パラメータ:
//   - stashRef: スタッシュの参照（例: stash@{0}, stash@{0}^2, stash@{0}^3）
//   - file: ファイルパス
//
// 戻り値:
//   - []byte: ファイルの内容
//   - bool: ファイルが存在する場合は true、存在しない場合は false
//   - error: エラーが発生した場合のエラー情報（存在しない場合は除く）
//
// 内部処理:
//   git show <stashRef>:<file> コマンドでファイルの内容を取得します。
//   終了コード 128 はファイルが存在しないことを示し、エラーとして扱いません。
func getStashFileContentByRef(stashRef, file string) ([]byte, bool, error) {
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

// findDuplicateStashes は重複するスタッシュのグループを検出します。
//
// パラメータ:
//   - stashInfos: 全スタッシュの詳細情報
//
// 戻り値:
//   - [][]StashInfo: 重複スタッシュのグループのスライス
//                    各グループはインデックスの昇順でソートされます
//
// 内部処理:
//   1. ハッシュ値をキーとして、同じハッシュを持つスタッシュをグループ化
//   2. 2つ以上のスタッシュを持つグループのみを抽出（重複グループ）
//   3. 各グループをインデックスの昇順でソート
func findDuplicateStashes(stashInfos []StashInfo) [][]StashInfo {
	hashMap := make(map[string][]StashInfo)

	for _, info := range stashInfos {
		hashMap[info.Hash] = append(hashMap[info.Hash], info)
	}

	duplicateGroups := make([][]StashInfo, 0)
	for _, group := range hashMap {
		if len(group) > 1 {
			sort.Slice(group, func(i, j int) bool {
				return group[i].Index < group[j].Index
			})
			duplicateGroups = append(duplicateGroups, group)
		}
	}

	return duplicateGroups
}

// deleteStashByIndex は指定されたインデックスのスタッシュを削除します。
//
// パラメータ:
//   - index: 削除するスタッシュのインデックス（stash@{N} の N）
//
// 戻り値:
//   - error: 削除に失敗した場合のエラー情報
//
// 内部処理:
//   git stash drop stash@{<index>} コマンドでスタッシュを削除します。
//
// 注意:
//   スタッシュを削除すると、それより後のスタッシュのインデックスが
//   繰り上がります。そのため、複数削除する場合は大きいインデックスから
//   順に削除する必要があります。
func deleteStashByIndex(index int) error {
	return gitcmd.RunQuiet("stash", "drop", fmt.Sprintf("stash@{%d}", index))
}

// init は stash-cleanup コマンドを root コマンドに登録します。
// この関数はパッケージの初期化時に自動的に呼び出されます。
func init() {
	rootCmd.AddCommand(stashCleanupCmd)
}
