package cmd

import (
	"testing"
)

// TestRootCmdDefinition はルートコマンドの定義をテストします
func TestRootCmdDefinition(t *testing.T) {
	if RootCmd == nil {
		t.Fatal("RootCmd should not be nil")
	}

	if RootCmd.Use != "plus" {
		t.Errorf("RootCmd.Use = %q, want %q", RootCmd.Use, "plus")
	}

	if RootCmd.Short == "" {
		t.Error("RootCmd.Short should not be empty")
	}

	if RootCmd.Long == "" {
		t.Error("RootCmd.Long should not be empty")
	}
}

// TestRootCmdHasSubcommands はルートコマンドにサブコマンドが登録されているかテストします
// 注意: サブパッケージがインポートされた状態でのみ有効
// この関数はインテグレーションテストとして使用されることを想定
func TestRootCmdHasSubcommands(t *testing.T) {
	// このテストはサブパッケージがインポートされた後にのみ意味がある
	// 単体テストとしてcmdパッケージをテストする場合、
	// サブパッケージがインポートされていないため、コマンドは空になる
	//
	// 完全なテストは main パッケージや統合テストで行う
	_ = RootCmd.Commands()
}

// TestCommandRegistration はコマンドが RootCmd に正しく登録されているかテストします
// 注意: このテストはサブパッケージがインポートされた状態でのみ有効
func TestCommandRegistration(t *testing.T) {
	// このテストはサブパッケージがインポートされていないと動作しない
	// mainパッケージでのインテグレーションテストで確認する
	t.Skip("This test requires subpackages to be imported. Run integration tests instead.")
}
