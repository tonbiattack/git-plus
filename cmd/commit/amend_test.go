package commit

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestAmendCmd_CommandSetup はamendコマンドの設定をテストします
func TestAmendCmd_CommandSetup(t *testing.T) {
	// コマンドが正しく設定されているか確認
	if amendCmd.Use != "amend" {
		t.Errorf("amendCmd.Use = %q, want %q", amendCmd.Use, "amend")
	}

	if amendCmd.Short == "" {
		t.Error("amendCmd.Short should not be empty")
	}

	if amendCmd.Long == "" {
		t.Error("amendCmd.Long should not be empty")
	}

	if amendCmd.Example == "" {
		t.Error("amendCmd.Example should not be empty")
	}

	// DisableFlagParsing が true であることを確認
	if !amendCmd.DisableFlagParsing {
		t.Error("amendCmd.DisableFlagParsing should be true")
	}
}

// TestAmendCmd_InRootCmd はamendコマンドがRootCmdに登録されていることを確認します
func TestAmendCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "amend" {
			found = true
			break
		}
	}

	if !found {
		t.Error("amendCmd should be registered in RootCmd")
	}
}

// TestAmendCmd_HasRunE はRunE関数が設定されていることを確認します
func TestAmendCmd_HasRunE(t *testing.T) {
	if amendCmd.RunE == nil {
		t.Error("amendCmd.RunE should not be nil")
	}
}

// TestAmendCmd_ExampleContent は例の内容が正しいことを確認します
func TestAmendCmd_ExampleContent(t *testing.T) {
	examples := amendCmd.Example

	// 重要な例が含まれていることを確認
	expectedExamples := []string{
		"git amend",
		"--no-edit",
		"--reset-author",
	}

	for _, expected := range expectedExamples {
		if !containsString(examples, expected) {
			t.Errorf("amendCmd.Example should contain %q", expected)
		}
	}
}

// containsString は文字列が含まれているかを確認するヘルパー関数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || containsString(s[1:], substr)))
}
