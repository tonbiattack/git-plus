package stash

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/gitcmd"
)

// StashEntry はスタッシュの詳細情報を表す構造体
type StashEntry struct {
	Index   int
	Ref     string
	Message string
	Files   []string
	Branch  string
}

var stashSelectCmd = &cobra.Command{
	Use:   "stash-select",
	Short: "インタラクティブにスタッシュを選択・操作",
	Long: `スタッシュされている変更を一覧表示し、インタラクティブに選択して操作できます。
各スタッシュのファイル一覧を確認しながら、apply（適用）、pop（適用して削除）、
drop（削除）、show（差分表示）などの操作を実行できます。`,
	Example: `  git stash-select`,
	RunE: func(c *cobra.Command, args []string) error {
		// スタッシュ一覧を取得
		stashes, err := getStashList()
		if err != nil {
			return fmt.Errorf("スタッシュ一覧の取得に失敗しました: %w", err)
		}

		if len(stashes) == 0 {
			fmt.Println("スタッシュが存在しません。")
			return nil
		}

		// スタッシュ一覧を表示
		fmt.Printf("スタッシュ一覧 (%d 個):\n\n", len(stashes))
		for i, stash := range stashes {
			fmt.Printf("%d. %s\n", i+1, stash.Ref)
			fmt.Printf("   ブランチ: %s\n", stash.Branch)
			fmt.Printf("   メッセージ: %s\n", stash.Message)
			fmt.Printf("   ファイル: ")
			if len(stash.Files) == 0 {
				fmt.Println("(なし)")
			} else if len(stash.Files) <= 3 {
				// 3つ以下なら1行で表示
				fmt.Println(strings.Join(stash.Files, ", "))
			} else {
				// 4つ以上なら最初の3つと残り数を表示
				fmt.Printf("%s ... (他%d個)\n", strings.Join(stash.Files[:3], ", "), len(stash.Files)-3)
			}
			fmt.Println()
		}

		// スタッシュを選択
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("選択してください (番号を入力、Enterでキャンセル): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("キャンセルしました。")
			return nil
		}

		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(stashes) {
			return fmt.Errorf("無効な番号です。1から%dの範囲で入力してください", len(stashes))
		}

		selectedStash := stashes[selection-1]

		// 選択したスタッシュの詳細を表示
		fmt.Printf("\n選択されたスタッシュ: %s\n", selectedStash.Ref)
		fmt.Printf("メッセージ: %s\n", selectedStash.Message)
		fmt.Printf("ブランチ: %s\n\n", selectedStash.Branch)

		fmt.Println("変更されたファイル:")
		if len(selectedStash.Files) == 0 {
			fmt.Println("  (ファイル情報を取得できませんでした)")
		} else {
			for _, file := range selectedStash.Files {
				fmt.Printf("  - %s\n", file)
			}
		}

		// 操作メニューを表示
		fmt.Println("\n操作を選択してください:")
		fmt.Println("  [a]pply  - スタッシュを適用（スタッシュは残す）")
		fmt.Println("  [p]op    - スタッシュを適用して削除")
		fmt.Println("  [d]rop   - スタッシュを削除")
		fmt.Println("  [s]how   - 差分を表示")
		fmt.Println("  [c]ancel - キャンセル")
		fmt.Print("\n選択 (a/p/d/s/c): ")

		action, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("入力の読み込みに失敗しました: %w", err)
		}

		action = strings.TrimSpace(strings.ToLower(action))

		switch action {
		case "a", "apply":
			fmt.Printf("\nスタッシュを適用しています: %s\n", selectedStash.Ref)
			if err := applyStash(selectedStash.Ref); err != nil {
				return fmt.Errorf("スタッシュの適用に失敗しました: %w", err)
			}
			fmt.Println("✓ スタッシュを適用しました")

		case "p", "pop":
			fmt.Printf("\nスタッシュを適用して削除しています: %s\n", selectedStash.Ref)
			if err := popStash(selectedStash.Ref); err != nil {
				return fmt.Errorf("スタッシュのpopに失敗しました: %w", err)
			}
			fmt.Println("✓ スタッシュを適用して削除しました")

		case "d", "drop":
			fmt.Printf("\nスタッシュを削除しています: %s\n", selectedStash.Ref)
			if err := dropStash(selectedStash.Ref); err != nil {
				return fmt.Errorf("スタッシュの削除に失敗しました: %w", err)
			}
			fmt.Println("✓ スタッシュを削除しました")

		case "s", "show":
			fmt.Printf("\nスタッシュの差分を表示しています: %s\n\n", selectedStash.Ref)
			if err := showStash(selectedStash.Ref); err != nil {
				return fmt.Errorf("スタッシュの表示に失敗しました: %w", err)
			}

		case "c", "cancel", "":
			fmt.Println("キャンセルしました。")
			return nil

		default:
			return fmt.Errorf("無効な操作です: %s", action)
		}

		return nil
	},
}

// getStashList はすべてのスタッシュ情報を取得します
func getStashList() ([]StashEntry, error) {
	// git stash list で一覧を取得
	output, err := gitcmd.Run("stash", "list", "--format=%gd|%gs")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	stashes := make([]StashEntry, 0, len(lines))

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// フォーマット: stash@{0}|WIP on branch: message
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		ref := parts[0]
		message := parts[1]

		// ブランチ名を抽出
		branch := extractBranch(message)

		// ファイル一覧を取得
		files, _ := getStashFiles(ref)

		stashes = append(stashes, StashEntry{
			Index:   i,
			Ref:     ref,
			Message: message,
			Files:   files,
			Branch:  branch,
		})
	}

	return stashes, nil
}

// extractBranch はメッセージからブランチ名を抽出します
func extractBranch(message string) string {
	// "WIP on branch-name: commit message" 形式を想定
	if strings.HasPrefix(message, "WIP on ") {
		parts := strings.SplitN(message[7:], ":", 2)
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if strings.HasPrefix(message, "On ") {
		parts := strings.SplitN(message[3:], ":", 2)
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return "(unknown)"
}

// getStashFiles はスタッシュのファイル一覧を取得します
func getStashFiles(ref string) ([]string, error) {
	output, err := gitcmd.Run("stash", "show", "--name-only", ref)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, line)
	}

	return files, nil
}

// applyStash はスタッシュを適用します（削除しない）
func applyStash(ref string) error {
	return gitcmd.RunWithIO("stash", "apply", ref)
}

// popStash はスタッシュを適用して削除します
func popStash(ref string) error {
	return gitcmd.RunWithIO("stash", "pop", ref)
}

// dropStash はスタッシュを削除します
func dropStash(ref string) error {
	return gitcmd.RunQuiet("stash", "drop", ref)
}

// showStash はスタッシュの差分を表示します
func showStash(ref string) error {
	return gitcmd.RunWithIO("stash", "show", "-p", ref)
}

func init() {
	cmd.RootCmd.AddCommand(stashSelectCmd)
}
