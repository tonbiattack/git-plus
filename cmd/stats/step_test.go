package stats

import (
	"os"
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestStepCmd_CommandSetup はstepコマンドの設定をテストします
func TestStepCmd_CommandSetup(t *testing.T) {
	if stepCmd.Use != "step" {
		t.Errorf("stepCmd.Use = %q, want %q", stepCmd.Use, "step")
	}

	if stepCmd.Short == "" {
		t.Error("stepCmd.Short should not be empty")
	}

	if stepCmd.Long == "" {
		t.Error("stepCmd.Long should not be empty")
	}

	if stepCmd.Example == "" {
		t.Error("stepCmd.Example should not be empty")
	}
}

// TestStepCmd_Flags はフラグが正しく設定されていることを確認します
func TestStepCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"since", "s"},
		{"until", "u"},
		{"weeks", "w"},
		{"months", "m"},
		{"years", "y"},
		{"include-initial", "i"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := stepCmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.name)
				return
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// TestFormatNum は数値フォーマット関数をテストします
func TestFormatNum(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{1000, "1,000"},
		{10000, "10,000"},
		{100000, "100,000"},
		{1000000, "1,000,000"},
		{1234567, "1,234,567"},
		{-1, "-1"},
		{-1000, "-1,000"},
		{-1234567, "-1,234,567"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatNum(tt.input)
			if result != tt.expected {
				t.Errorf("formatNum(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestAuthorStats_Fields はAuthorStats構造体をテストします
func TestAuthorStats_Fields(t *testing.T) {
	stats := AuthorStats{
		Name:          "Test User",
		Added:         100,
		Deleted:       50,
		Net:           50,
		Modified:      150,
		CurrentCode:   500,
		Commits:       10,
		AvgCommitSize: 15.0,
	}

	if stats.Name != "Test User" {
		t.Errorf("AuthorStats.Name = %q, want %q", stats.Name, "Test User")
	}

	if stats.Added != 100 {
		t.Errorf("AuthorStats.Added = %d, want %d", stats.Added, 100)
	}

	if stats.Deleted != 50 {
		t.Errorf("AuthorStats.Deleted = %d, want %d", stats.Deleted, 50)
	}

	if stats.Net != 50 {
		t.Errorf("AuthorStats.Net = %d, want %d", stats.Net, 50)
	}

	if stats.Modified != 150 {
		t.Errorf("AuthorStats.Modified = %d, want %d", stats.Modified, 150)
	}

	if stats.CurrentCode != 500 {
		t.Errorf("AuthorStats.CurrentCode = %d, want %d", stats.CurrentCode, 500)
	}

	if stats.Commits != 10 {
		t.Errorf("AuthorStats.Commits = %d, want %d", stats.Commits, 10)
	}

	if stats.AvgCommitSize != 15.0 {
		t.Errorf("AuthorStats.AvgCommitSize = %f, want %f", stats.AvgCommitSize, 15.0)
	}
}

// TestStepCmd_InRootCmd はstepコマンドがRootCmdに登録されていることを確認します
func TestStepCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "step" {
			found = true
			break
		}
	}

	if !found {
		t.Error("stepCmd should be registered in RootCmd")
	}
}

// TestCollectAuthorStats_SingleAuthor は単一作成者の統計収集をテストします
func TestCollectAuthorStats_SingleAuthor(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test\nLine 2\nLine 3\n")
	repo.Commit("Initial commit")

	// 追加のコミットを作成
	repo.CreateFile("file1.txt", "content\n")
	repo.Commit("Add file1")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	stats := collectAuthorStats("", "", false)

	if len(stats) == 0 {
		t.Error("collectAuthorStats() returned empty stats")
		return
	}

	// 少なくとも1人の作成者がいることを確認
	if stats[0].Name == "" {
		t.Error("Author name should not be empty")
	}

	// コミット数が2以上であることを確認
	totalCommits := 0
	for _, s := range stats {
		totalCommits += s.Commits
	}
	if totalCommits < 2 {
		t.Errorf("Total commits = %d, want >= 2", totalCommits)
	}
}

// TestCollectAuthorStats_ExcludeInitial は初回コミット除外をテストします
func TestCollectAuthorStats_ExcludeInitial(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成（大量の行を追加）
	repo.CreateFile("README.md", "# Test\nLine 2\nLine 3\nLine 4\nLine 5\n")
	repo.Commit("Initial commit")

	// 追加のコミットを作成
	repo.CreateFile("file1.txt", "small\n")
	repo.Commit("Add small file")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// 初回コミットを含める場合
	statsWithInitial := collectAuthorStats("", "", false)

	// 初回コミットを除外する場合
	statsWithoutInitial := collectAuthorStats("", "", true)

	// 初回コミットを除外した場合、追加行数が少ないことを確認
	totalAddedWith := 0
	for _, s := range statsWithInitial {
		totalAddedWith += s.Added
	}

	totalAddedWithout := 0
	for _, s := range statsWithoutInitial {
		totalAddedWithout += s.Added
	}

	if totalAddedWithout >= totalAddedWith {
		t.Errorf("Excluding initial commit should reduce total added lines: with=%d, without=%d", totalAddedWith, totalAddedWithout)
	}
}

// TestGetTotalLines_EmptyRepo は空のリポジトリでの総行数をテストします
func TestGetTotalLines_EmptyRepo(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 初期コミットを作成
	repo.CreateFile("README.md", "# Test")
	repo.Commit("Initial commit")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	lines := getTotalLines()

	// 少なくとも1行があることを確認
	if lines < 1 {
		t.Errorf("getTotalLines() = %d, want >= 1", lines)
	}
}

// TestGetTotalLines_MultipleFiles は複数ファイルの総行数をテストします
func TestGetTotalLines_MultipleFiles(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 複数のファイルを作成
	repo.CreateFile("file1.txt", "line1\nline2\nline3\n")
	repo.CreateFile("file2.txt", "a\nb\n")
	repo.MustGit("add", ".")
	repo.Commit("Add files")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	lines := getTotalLines()

	// file1.txt: 3行, file2.txt: 2行 = 合計5行
	if lines != 5 {
		t.Errorf("getTotalLines() = %d, want %d", lines, 5)
	}
}

// TestGetValidCommits_AllCommits は全コミットの取得をテストします
func TestGetValidCommits_AllCommits(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// 複数のコミットを作成
	repo.CreateFile("file1.txt", "content1")
	repo.Commit("Commit 1")

	repo.CreateFile("file2.txt", "content2")
	repo.Commit("Commit 2")

	repo.CreateFile("file3.txt", "content3")
	repo.Commit("Commit 3")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	commits := getValidCommits("", "")

	// 3つのコミットが存在することを確認
	if len(commits) != 3 {
		t.Errorf("getValidCommits() returned %d commits, want %d", len(commits), 3)
	}

	// すべてのコミットハッシュが40文字であることを確認
	for hash := range commits {
		if len(hash) != 40 {
			t.Errorf("Commit hash %q should be 40 characters", hash)
		}
	}
}

// TestGetCodeByAuthor_SingleAuthor は単一作成者のコード貢献をテストします
func TestGetCodeByAuthor_SingleAuthor(t *testing.T) {
	repo := testutil.NewGitRepo(t)

	// ファイルを作成してコミット
	repo.CreateFile("file1.txt", "line1\nline2\nline3\n")
	repo.Commit("Add file")

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	authorLines := getCodeByAuthor("", "")

	// 少なくとも1人の作成者がいることを確認
	if len(authorLines) == 0 {
		t.Error("getCodeByAuthor() returned empty map")
		return
	}

	// 合計行数が3であることを確認
	totalLines := 0
	for _, lines := range authorLines {
		totalLines += lines
	}

	if totalLines != 3 {
		t.Errorf("Total lines by author = %d, want %d", totalLines, 3)
	}
}
