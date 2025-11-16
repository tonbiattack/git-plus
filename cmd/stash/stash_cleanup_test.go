package stash

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestStashCleanupCmd_CommandSetup はstash-cleanupコマンドの設定をテストします
func TestStashCleanupCmd_CommandSetup(t *testing.T) {
	if stashCleanupCmd.Use != "stash-cleanup" {
		t.Errorf("stashCleanupCmd.Use = %q, want %q", stashCleanupCmd.Use, "stash-cleanup")
	}

	if stashCleanupCmd.Short == "" {
		t.Error("stashCleanupCmd.Short should not be empty")
	}

	if stashCleanupCmd.Long == "" {
		t.Error("stashCleanupCmd.Long should not be empty")
	}
}

// TestStashCleanupCmd_InRootCmd はstash-cleanupコマンドがrootCmdに登録されていることを確認します
func TestStashCleanupCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "stash-cleanup" {
			found = true
			break
		}
	}

	if !found {
		t.Error("stashCleanupCmd should be registered in rootCmd")
	}
}

// TestStashInfo_Fields はStashInfo構造体をテストします
func TestStashInfo_Fields(t *testing.T) {
	info := StashInfo{
		Index: 0,
		Name:  "stash@{0}",
		Files: []string{"file1.txt", "file2.txt"},
		Hash:  "abc123",
	}

	if info.Index != 0 {
		t.Errorf("StashInfo.Index = %d, want %d", info.Index, 0)
	}

	if info.Name != "stash@{0}" {
		t.Errorf("StashInfo.Name = %q, want %q", info.Name, "stash@{0}")
	}

	if len(info.Files) != 2 {
		t.Errorf("StashInfo.Files length = %d, want %d", len(info.Files), 2)
	}

	if info.Hash != "abc123" {
		t.Errorf("StashInfo.Hash = %q, want %q", info.Hash, "abc123")
	}
}

// TestGetAllStashesList はスタッシュ一覧の取得をテストします
func TestGetAllStashesList(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// スタッシュがない状態
	stashes, err := getAllStashesList()
	if err != nil {
		t.Errorf("getAllStashesList returned error: %v", err)
	}

	if len(stashes) != 0 {
		t.Errorf("Expected 0 stashes, got %d", len(stashes))
	}
}

// TestGetAllStashesList_WithStashes はスタッシュがある場合をテストします
func TestGetAllStashesList_WithStashes(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	// 変更を作成してスタッシュ（追跡されたファイルを変更する）
	repo.CreateFile("file1.txt", "content1")
	repo.MustGit("add", "file1.txt")
	repo.CreateFile("file1.txt", "modified content")
	repo.StashPush("Test stash 1")

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	stashes, err := getAllStashesList()
	if err != nil {
		t.Errorf("getAllStashesList returned error: %v", err)
	}

	if len(stashes) != 1 {
		t.Errorf("Expected 1 stash, got %d", len(stashes))
	}

	if len(stashes) > 0 && stashes[0] != "stash@{0}" {
		t.Errorf("First stash = %q, want %q", stashes[0], "stash@{0}")
	}
}

// TestFindDuplicateStashes は重複検出をテストします
func TestFindDuplicateStashes(t *testing.T) {
	tests := []struct {
		name     string
		infos    []StashInfo
		expected int
	}{
		{
			name:     "空のリスト",
			infos:    []StashInfo{},
			expected: 0,
		},
		{
			name: "重複なし",
			infos: []StashInfo{
				{Index: 0, Hash: "abc123"},
				{Index: 1, Hash: "def456"},
				{Index: 2, Hash: "ghi789"},
			},
			expected: 0,
		},
		{
			name: "1つの重複グループ",
			infos: []StashInfo{
				{Index: 0, Hash: "abc123"},
				{Index: 1, Hash: "abc123"},
				{Index: 2, Hash: "def456"},
			},
			expected: 1,
		},
		{
			name: "複数の重複グループ",
			infos: []StashInfo{
				{Index: 0, Hash: "abc123"},
				{Index: 1, Hash: "abc123"},
				{Index: 2, Hash: "def456"},
				{Index: 3, Hash: "def456"},
			},
			expected: 2,
		},
		{
			name: "3つ以上の同じハッシュ",
			infos: []StashInfo{
				{Index: 0, Hash: "abc123"},
				{Index: 1, Hash: "abc123"},
				{Index: 2, Hash: "abc123"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := findDuplicateStashes(tt.infos)
			if len(groups) != tt.expected {
				t.Errorf("findDuplicateStashes() returned %d groups, want %d", len(groups), tt.expected)
			}
		})
	}
}

// TestFindDuplicateStashes_Sorting は重複グループがインデックス順でソートされることを確認します
func TestFindDuplicateStashes_Sorting(t *testing.T) {
	infos := []StashInfo{
		{Index: 2, Hash: "abc123"},
		{Index: 0, Hash: "abc123"},
		{Index: 1, Hash: "abc123"},
	}

	groups := findDuplicateStashes(infos)
	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	group := groups[0]
	if len(group) != 3 {
		t.Fatalf("Expected 3 items in group, got %d", len(group))
	}

	// インデックス順でソートされているはず
	if group[0].Index != 0 {
		t.Errorf("First item index = %d, want 0", group[0].Index)
	}
	if group[1].Index != 1 {
		t.Errorf("Second item index = %d, want 1", group[1].Index)
	}
	if group[2].Index != 2 {
		t.Errorf("Third item index = %d, want 2", group[2].Index)
	}
}
