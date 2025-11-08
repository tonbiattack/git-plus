package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type StashInfo struct {
	Index int
	Name  string
	Files []string
	Hash  string
}

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
	fmt.Print("続行しますか? (y/N): ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	if response != "y" && response != "yes" {
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
func getAllStashes() ([]string, error) {
	cmd := exec.Command("git", "stash", "list")
	output, err := cmd.Output()
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
func getStashFiles(stash string) ([]string, error) {
	cmd := exec.Command("git", "stash", "show", "--name-only", stash)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	sort.Strings(files)
	return files, nil
}

// getStashContentHash はスタッシュの内容からハッシュを生成する
func getStashContentHash(stash string, files []string) (string, error) {
	var buffer bytes.Buffer

	// ファイル名を含める（ソート済み）
	buffer.WriteString(strings.Join(files, "\n"))
	buffer.WriteString("\n---\n")

	// 各ファイルの内容を連結
	for _, file := range files {
		content, err := getStashFileContent(stash, file)
		if err != nil {
			// ファイル取得エラーは内容に反映
			buffer.WriteString(fmt.Sprintf("ERROR:%s\n", file))
			continue
		}
		buffer.Write(content)
		buffer.WriteString("\n")
	}

	// SHA-1ハッシュを計算
	cmd := exec.Command("git", "hash-object", "--stdin")
	cmd.Stdin = &buffer
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// getStashFileContent はstashの特定のファイルの内容を取得する
func getStashFileContent(stash, file string) ([]byte, error) {
	ref := fmt.Sprintf("%s:%s", stash, file)
	cmd := exec.Command("git", "show", ref)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// findDuplicates は重複するスタッシュグループを検出する
func findDuplicates(stashInfos []StashInfo) [][]StashInfo {
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
			sort.Slice(group, func(i, j int) bool {
				return group[i].Index < group[j].Index
			})
			duplicateGroups = append(duplicateGroups, group)
		}
	}

	return duplicateGroups
}

// deleteStash は指定したインデックスのスタッシュを削除する
func deleteStash(index int) error {
	// git stash drop stash@{N}
	cmd := exec.Command("git", "stash", "drop", fmt.Sprintf("stash@{%d}", index))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}

	return nil
}

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
