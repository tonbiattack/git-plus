package stats

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
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
