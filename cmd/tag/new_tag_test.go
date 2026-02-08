package tag

import (
	"errors"
	"os/exec"
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
		{"release-note flag", "release-note", ""},
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

// TestIsNoTagsDescribeError は git describe でタグが存在しない場合のエラー判定をテストします
func TestIsNoTagsDescribeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "detects no tags error",
			err: &exec.ExitError{
				Stderr: []byte("fatal: No names found, cannot describe anything.\n"),
			},
			want: true,
		},
		{
			name: "ignores other exit errors",
			err: &exec.ExitError{
				Stderr: []byte("fatal: not a git repository (or any of the parent directories): .git"),
			},
			want: false,
		},
		{
			name: "ignores generic errors",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNoTagsDescribeError(tt.err); got != tt.want {
				t.Fatalf("isNoTagsDescribeError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveTagMessage はタグメッセージの解決ロジックをテストします
func TestResolveTagMessage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		message  string
		expected string
	}{
		{
			name:     "custom message",
			tag:      "v1.2.3",
			message:  "Release notes",
			expected: "Release notes",
		},
		{
			name:     "trimmed message",
			tag:      "v1.2.3",
			message:  "  Custom  ",
			expected: "Custom",
		},
		{
			name:     "default message when empty",
			tag:      "v1.2.3",
			message:  "",
			expected: "Release v1.2.3",
		},
		{
			name:     "default message when whitespace",
			tag:      "v1.2.3",
			message:  "   ",
			expected: "Release v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTagMessage(tt.tag, tt.message); got != tt.expected {
				t.Errorf("resolveTagMessage(%q, %q) = %q, want %q", tt.tag, tt.message, got, tt.expected)
			}
		})
	}
}

// TestParseGitHubOriginURL は GitHub origin URL の解析をテストします
func TestParseGitHubOriginURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		expectErr bool
	}{
		{
			name:      "https url",
			input:     "https://github.com/example/project.git",
			wantOwner: "example",
			wantRepo:  "project",
			expectErr: false,
		},
		{
			name:      "ssh url",
			input:     "git@github.com:example/project.git",
			wantOwner: "example",
			wantRepo:  "project",
			expectErr: false,
		},
		{
			name:      "https url without .git",
			input:     "https://github.com/example/project",
			wantOwner: "example",
			wantRepo:  "project",
			expectErr: false,
		},
		{
			name:      "non github url",
			input:     "https://gitlab.com/example/project.git",
			expectErr: true,
		},
		{
			name:      "invalid format",
			input:     "https://github.com/example",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseGitHubOriginURL(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("parseGitHubOriginURL(%q) expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseGitHubOriginURL(%q) unexpected error: %v", tt.input, err)
			}

			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

// TestBuildGitHubCompareURL は差分リンク生成をテストします
func TestBuildGitHubCompareURL(t *testing.T) {
	url, err := buildGitHubCompareURL("https://github.com/example/project.git", "v1.2.3", "v1.2.4")
	if err != nil {
		t.Fatalf("buildGitHubCompareURL returned error: %v", err)
	}

	expected := "https://github.com/example/project/compare/v1.2.3...v1.2.4"
	if url != expected {
		t.Errorf("compare url = %q, want %q", url, expected)
	}
}

// TestBuildReleaseNotesWithPrefix はリリースノートの組み立てをテストします
func TestBuildReleaseNotesWithPrefix(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		prefix   string
		want     string
	}{
		{
			name:     "prepend to existing",
			existing: "Auto notes",
			prefix:   "2026-02-08 / PROJ-1234",
			want:     "2026-02-08 / PROJ-1234\n\nAuto notes",
		},
		{
			name:     "existing empty",
			existing: "",
			prefix:   "2026-02-08 / PROJ-1234",
			want:     "2026-02-08 / PROJ-1234",
		},
		{
			name:     "prefix empty",
			existing: "Auto notes",
			prefix:   "",
			want:     "Auto notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildReleaseNotesWithPrefix(tt.existing, tt.prefix); got != tt.want {
				t.Errorf("buildReleaseNotesWithPrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestResolveReleaseNote はリリースノートのデフォルト設定をテストします
func TestResolveReleaseNote(t *testing.T) {
	if got := resolveReleaseNote("2026-02-08 / PROJ-1234"); got != "2026-02-08 / PROJ-1234" {
		t.Errorf("resolveReleaseNote(custom) = %q", got)
	}

	got := resolveReleaseNote("")
	if got == "" {
		t.Error("resolveReleaseNote(empty) should not be empty")
	}
}
