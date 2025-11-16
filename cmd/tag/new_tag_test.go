package tag

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestNewTagCmdDefinition はnew-tagコマンドの定義をテストします
func TestNewTagCmdDefinition(t *testing.T) {
	if newTagCmd == nil {
		t.Fatal("newTagCmd should not be nil")
	}

	if newTagCmd.Use != "new-tag [type]" {
		t.Errorf("newTagCmd.Use = %q, want %q", newTagCmd.Use, "new-tag [type]")
	}

	if newTagCmd.Short == "" {
		t.Error("newTagCmd.Short should not be empty")
	}

	if newTagCmd.Long == "" {
		t.Error("newTagCmd.Long should not be empty")
	}

	if newTagCmd.Example == "" {
		t.Error("newTagCmd.Example should not be empty")
	}

	if newTagCmd.RunE == nil {
		t.Error("newTagCmd.RunE should not be nil")
	}
}

// TestNewTagCmdFlags はnew-tagコマンドのフラグをテストします
func TestNewTagCmdFlags(t *testing.T) {
	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"message flag", "message", "m"},
		{"push flag", "push", "p"},
		{"dry-run flag", "dry-run", "d"},
		{"release flag", "release", "r"},
		{"release-draft flag", "release-draft", "D"},
		{"release-prerelease flag", "release-prerelease", "P"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := newTagCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %q not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

// TestExtractVersion はextractVersion関数をテストします
func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		expectedMajor int
		expectedMinor int
		expectedPatch int
		expectError   bool
	}{
		{
			name:          "通常のvプレフィックス付きタグ",
			tag:           "v1.2.3",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
			expectError:   false,
		},
		{
			name:          "vプレフィックスなしタグ",
			tag:           "1.2.3",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
			expectError:   false,
		},
		{
			name:          "大きなバージョン番号",
			tag:           "v100.200.300",
			expectedMajor: 100,
			expectedMinor: 200,
			expectedPatch: 300,
			expectError:   false,
		},
		{
			name:          "ゼロパディング",
			tag:           "v0.0.1",
			expectedMajor: 0,
			expectedMinor: 0,
			expectedPatch: 1,
			expectError:   false,
		},
		{
			name:        "無効な形式",
			tag:         "invalid",
			expectError: true,
		},
		{
			name:        "不完全なバージョン",
			tag:         "v1.2",
			expectError: true,
		},
		{
			name:        "空文字列",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := extractVersion(tt.tag)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if major != tt.expectedMajor {
				t.Errorf("major = %d, want %d", major, tt.expectedMajor)
			}

			if minor != tt.expectedMinor {
				t.Errorf("minor = %d, want %d", minor, tt.expectedMinor)
			}

			if patch != tt.expectedPatch {
				t.Errorf("patch = %d, want %d", patch, tt.expectedPatch)
			}
		})
	}
}

// TestNormalizeVersionTypeName はnormalizeVersionTypeName関数をテストします
func TestNormalizeVersionTypeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// major関連
		{"major", "major", "major"},
		{"m", "m", "major"},
		{"breaking", "breaking", "major"},
		{"MAJOR", "MAJOR", "major"},
		{"M", "M", "major"},

		// minor関連
		{"minor", "minor", "minor"},
		{"n", "n", "minor"},
		{"feature", "feature", "minor"},
		{"f", "f", "minor"},
		{"MINOR", "MINOR", "minor"},
		{"N", "N", "minor"},
		{"FEATURE", "FEATURE", "minor"},
		{"F", "F", "minor"},

		// patch関連
		{"patch", "patch", "patch"},
		{"p", "p", "patch"},
		{"bug", "bug", "patch"},
		{"b", "b", "patch"},
		{"fix", "fix", "patch"},
		{"PATCH", "PATCH", "patch"},
		{"P", "P", "patch"},
		{"BUG", "BUG", "patch"},
		{"B", "B", "patch"},
		{"FIX", "FIX", "patch"},

		// 無効な入力
		{"invalid", "invalid", ""},
		{"empty", "", ""},
		{"unknown", "unknown", ""},

		// 空白付き入力
		{"前後空白付きmajor", "  major  ", "major"},
		{"前後空白付きminor", "  feature  ", "minor"},
		{"前後空白付きpatch", "  bug  ", "patch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVersionTypeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersionTypeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestComputeNewVersion はcomputeNewVersion関数をテストします
func TestComputeNewVersion(t *testing.T) {
	tests := []struct {
		name          string
		major         int
		minor         int
		patch         int
		versionType   string
		expectedMajor int
		expectedMinor int
		expectedPatch int
	}{
		{
			name:          "major update",
			major:         1,
			minor:         2,
			patch:         3,
			versionType:   "major",
			expectedMajor: 2,
			expectedMinor: 0,
			expectedPatch: 0,
		},
		{
			name:          "minor update",
			major:         1,
			minor:         2,
			patch:         3,
			versionType:   "minor",
			expectedMajor: 1,
			expectedMinor: 3,
			expectedPatch: 0,
		},
		{
			name:          "patch update",
			major:         1,
			minor:         2,
			patch:         3,
			versionType:   "patch",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 4,
		},
		{
			name:          "unknown type defaults to patch",
			major:         1,
			minor:         2,
			patch:         3,
			versionType:   "unknown",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 4,
		},
		{
			name:          "major from zero",
			major:         0,
			minor:         0,
			patch:         1,
			versionType:   "major",
			expectedMajor: 1,
			expectedMinor: 0,
			expectedPatch: 0,
		},
		{
			name:          "minor from zero patch",
			major:         1,
			minor:         0,
			patch:         0,
			versionType:   "minor",
			expectedMajor: 1,
			expectedMinor: 1,
			expectedPatch: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newMajor, newMinor, newPatch := computeNewVersion(tt.major, tt.minor, tt.patch, tt.versionType)

			if newMajor != tt.expectedMajor {
				t.Errorf("newMajor = %d, want %d", newMajor, tt.expectedMajor)
			}

			if newMinor != tt.expectedMinor {
				t.Errorf("newMinor = %d, want %d", newMinor, tt.expectedMinor)
			}

			if newPatch != tt.expectedPatch {
				t.Errorf("newPatch = %d, want %d", newPatch, tt.expectedPatch)
			}
		})
	}
}

// TestNewTagCommandRegistration はnew-tagコマンドがRootCmdに登録されているかテストします
func TestNewTagCommandRegistration(t *testing.T) {
	c, _, err := cmd.RootCmd.Find([]string{"new-tag"})
	if err != nil {
		t.Errorf("new-tag command not found: %v", err)
	}
	if c == nil {
		t.Error("new-tag command is nil")
	}
	if c.Name() != "new-tag" {
		t.Errorf("Command name = %q, want %q", c.Name(), "new-tag")
	}
}

// TestVersionTypeAliases はバージョンタイプのエイリアスをテストします
func TestVersionTypeAliases(t *testing.T) {
	// major のエイリアス
	majorAliases := []string{"major", "m", "breaking"}
	for _, alias := range majorAliases {
		result := normalizeVersionTypeName(alias)
		if result != "major" {
			t.Errorf("%q should normalize to 'major', got %q", alias, result)
		}
	}

	// minor のエイリアス
	minorAliases := []string{"minor", "n", "feature", "f"}
	for _, alias := range minorAliases {
		result := normalizeVersionTypeName(alias)
		if result != "minor" {
			t.Errorf("%q should normalize to 'minor', got %q", alias, result)
		}
	}

	// patch のエイリアス
	patchAliases := []string{"patch", "p", "bug", "b", "fix"}
	for _, alias := range patchAliases {
		result := normalizeVersionTypeName(alias)
		if result != "patch" {
			t.Errorf("%q should normalize to 'patch', got %q", alias, result)
		}
	}
}
