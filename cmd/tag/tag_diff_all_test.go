package tag

import (
	"testing"
	"time"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestTagDiffAllCmd_CommandSetup はtag-diff-allコマンドの設定をテストします
func TestTagDiffAllCmd_CommandSetup(t *testing.T) {
	if tagDiffAllCmd.Use != "tag-diff-all" {
		t.Errorf("tagDiffAllCmd.Use = %q, want %q", tagDiffAllCmd.Use, "tag-diff-all")
	}

	if tagDiffAllCmd.Short == "" {
		t.Error("tagDiffAllCmd.Short should not be empty")
	}

	if tagDiffAllCmd.Long == "" {
		t.Error("tagDiffAllCmd.Long should not be empty")
	}

	if tagDiffAllCmd.Example == "" {
		t.Error("tagDiffAllCmd.Example should not be empty")
	}
}

// TestTagDiffAllCmd_Flags はフラグが正しく設定されていることを確認します
func TestTagDiffAllCmd_Flags(t *testing.T) {
	tests := []struct {
		name      string
		shorthand string
	}{
		{"prefix", "p"},
		{"output", "o"},
		{"split", "s"},
		{"limit", "l"},
		{"reverse", "r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tagDiffAllCmd.Flags().Lookup(tt.name)
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

// TestTagDiffAllCmd_InRootCmd はtag-diff-allコマンドがRootCmdに登録されていることを確認します
func TestTagDiffAllCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "tag-diff-all" {
			found = true
			break
		}
	}

	if !found {
		t.Error("tagDiffAllCmd should be registered in RootCmd")
	}
}

// TestTagInfo_Fields はtagInfo構造体をテストします
func TestTagInfo_Fields(t *testing.T) {
	testTime := time.Now()
	info := tagInfo{
		Name: "v1.0.0",
		Date: testTime,
	}

	if info.Name != "v1.0.0" {
		t.Errorf("tagInfo.Name = %q, want %q", info.Name, "v1.0.0")
	}

	if !info.Date.Equal(testTime) {
		t.Errorf("tagInfo.Date = %v, want %v", info.Date, testTime)
	}
}

// TestFilterTagsByPrefix はタグのプレフィックスフィルタリングをテストします
func TestFilterTagsByPrefix(t *testing.T) {
	now := time.Now()
	tags := []tagInfo{
		{Name: "v1.0.0", Date: now},
		{Name: "v1.1.0", Date: now},
		{Name: "release-1.0", Date: now},
		{Name: "v2.0.0", Date: now},
		{Name: "release-2.0", Date: now},
	}

	tests := []struct {
		prefix       string
		expectedLen  int
		expectedTags []string
	}{
		{"v", 3, []string{"v1.0.0", "v1.1.0", "v2.0.0"}},
		{"v1", 2, []string{"v1.0.0", "v1.1.0"}},
		{"release", 2, []string{"release-1.0", "release-2.0"}},
		{"xyz", 0, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			filtered := filterTagsByPrefix(tags, tt.prefix)
			if len(filtered) != tt.expectedLen {
				t.Errorf("filterTagsByPrefix with prefix %q returned %d tags, want %d", tt.prefix, len(filtered), tt.expectedLen)
			}
		})
	}
}

// TestGenerateTagPairs はタグペアの生成をテストします
func TestGenerateTagPairs(t *testing.T) {
	now := time.Now()
	tags := []tagInfo{
		{Name: "v1.0.0", Date: now.Add(-3 * time.Hour)},
		{Name: "v1.1.0", Date: now.Add(-2 * time.Hour)},
		{Name: "v2.0.0", Date: now.Add(-1 * time.Hour)},
	}

	// 通常順（古い→新しい）
	tagDiffReverse = false
	pairs := generateTagPairs(tags)

	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(pairs))
	}

	if len(pairs) >= 1 {
		if pairs[0][0].Name != "v1.0.0" || pairs[0][1].Name != "v1.1.0" {
			t.Errorf("First pair should be v1.0.0 → v1.1.0")
		}
	}

	if len(pairs) >= 2 {
		if pairs[1][0].Name != "v1.1.0" || pairs[1][1].Name != "v2.0.0" {
			t.Errorf("Second pair should be v1.1.0 → v2.0.0")
		}
	}
}

// TestGenerateTagPairs_Reverse は逆順のタグペア生成をテストします
func TestGenerateTagPairs_Reverse(t *testing.T) {
	now := time.Now()
	tags := []tagInfo{
		{Name: "v1.0.0", Date: now.Add(-3 * time.Hour)},
		{Name: "v1.1.0", Date: now.Add(-2 * time.Hour)},
		{Name: "v2.0.0", Date: now.Add(-1 * time.Hour)},
	}

	// 逆順（新しい→古い）
	tagDiffReverse = true
	pairs := generateTagPairs(tags)

	if len(pairs) != 2 {
		t.Errorf("Expected 2 pairs, got %d", len(pairs))
	}

	if len(pairs) >= 1 {
		if pairs[0][0].Name != "v1.1.0" || pairs[0][1].Name != "v2.0.0" {
			t.Errorf("First pair should be v1.1.0 → v2.0.0 in reverse mode")
		}
	}

	// 元に戻す
	tagDiffReverse = false
}

// TestParseGitDate はgitの日付パースをテストします
func TestParseGitDate(t *testing.T) {
	tests := []struct {
		dateStr string
		valid   bool
	}{
		{"2024-01-15 10:30:45 +0900", true},
		{"2024-01-15 10:30:45 -0700", true},
		{"invalid date", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			_, err := parseGitDate(tt.dateStr)
			if tt.valid && err != nil {
				t.Errorf("parseGitDate(%q) returned error: %v", tt.dateStr, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("parseGitDate(%q) should return error", tt.dateStr)
			}
		})
	}
}
