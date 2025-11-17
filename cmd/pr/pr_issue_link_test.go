package pr

import (
	"testing"

	"github.com/tonbiattack/git-plus/cmd"
)

// TestPrIssueLinkCmd_CommandSetup はpr-issue-linkコマンドの設定をテストします
func TestPrIssueLinkCmd_CommandSetup(t *testing.T) {
	if prIssueLinkCmd.Use != "pr-issue-link [--base ベースブランチ] [--issue イシュー番号]" {
		t.Errorf("prIssueLinkCmd.Use = %q, want %q", prIssueLinkCmd.Use, "pr-issue-link [--base ベースブランチ] [--issue イシュー番号]")
	}

	if prIssueLinkCmd.Short == "" {
		t.Error("prIssueLinkCmd.Short should not be empty")
	}

	if prIssueLinkCmd.Long == "" {
		t.Error("prIssueLinkCmd.Long should not be empty")
	}

	if prIssueLinkCmd.Example == "" {
		t.Error("prIssueLinkCmd.Example should not be empty")
	}
}

// TestPrIssueLinkCmd_InRootCmd はpr-issue-linkコマンドがRootCmdに登録されていることを確認します
func TestPrIssueLinkCmd_InRootCmd(t *testing.T) {
	found := false
	for _, c := range cmd.RootCmd.Commands() {
		if c.Use == "pr-issue-link [--base ベースブランチ] [--issue イシュー番号]" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prIssueLinkCmd should be registered in RootCmd")
	}
}

// TestPrIssueLinkCmd_HasRunE はRunE関数が設定されていることを確認します
func TestPrIssueLinkCmd_HasRunE(t *testing.T) {
	if prIssueLinkCmd.RunE == nil {
		t.Error("prIssueLinkCmd.RunE should not be nil")
	}
}

// TestPrIssueLinkCmd_HasFlags はフラグが正しく設定されていることを確認します
func TestPrIssueLinkCmd_HasFlags(t *testing.T) {
	// base フラグの確認
	baseFlag := prIssueLinkCmd.Flags().Lookup("base")
	if baseFlag == nil {
		t.Error("prIssueLinkCmd should have 'base' flag")
	} else {
		if baseFlag.Shorthand != "b" {
			t.Errorf("base flag shorthand = %q, want %q", baseFlag.Shorthand, "b")
		}
	}

	// issue フラグの確認
	issueFlag := prIssueLinkCmd.Flags().Lookup("issue")
	if issueFlag == nil {
		t.Error("prIssueLinkCmd should have 'issue' flag")
	} else {
		if issueFlag.Shorthand != "i" {
			t.Errorf("issue flag shorthand = %q, want %q", issueFlag.Shorthand, "i")
		}
	}

	// title フラグの確認
	titleFlag := prIssueLinkCmd.Flags().Lookup("title")
	if titleFlag == nil {
		t.Error("prIssueLinkCmd should have 'title' flag")
	} else {
		if titleFlag.Shorthand != "t" {
			t.Errorf("title flag shorthand = %q, want %q", titleFlag.Shorthand, "t")
		}
	}

	// body フラグの確認
	bodyFlag := prIssueLinkCmd.Flags().Lookup("body")
	if bodyFlag == nil {
		t.Error("prIssueLinkCmd should have 'body' flag")
	}
}

// TestBuildPRBody_SingleIssue は単一のIssueで本文が正しく構築されることをテストします
func TestBuildPRBody_SingleIssue(t *testing.T) {
	issues := []int{123}
	body := buildPRBody("", issues)
	expected := "Closes #123"

	if body != expected {
		t.Errorf("buildPRBody() = %q, want %q", body, expected)
	}
}

// TestBuildPRBody_MultipleIssues は複数のIssueで本文が正しく構築されることをテストします
func TestBuildPRBody_MultipleIssues(t *testing.T) {
	issues := []int{123, 456, 789}
	body := buildPRBody("", issues)
	expected := "Closes #123\nCloses #456\nCloses #789"

	if body != expected {
		t.Errorf("buildPRBody() = %q, want %q", body, expected)
	}
}

// TestBuildPRBody_WithUserBody はユーザー本文とIssueが正しく結合されることをテストします
func TestBuildPRBody_WithUserBody(t *testing.T) {
	issues := []int{42}
	userBody := "This is a fix for the authentication bug."
	body := buildPRBody(userBody, issues)
	expected := "This is a fix for the authentication bug.\n\nCloses #42"

	if body != expected {
		t.Errorf("buildPRBody() = %q, want %q", body, expected)
	}
}

// TestBuildPRBody_NoIssues はIssueがない場合にユーザー本文のみが返されることをテストします
func TestBuildPRBody_NoIssues(t *testing.T) {
	issues := []int{}
	userBody := "Just a simple PR"
	body := buildPRBody(userBody, issues)
	expected := "Just a simple PR"

	if body != expected {
		t.Errorf("buildPRBody() = %q, want %q", body, expected)
	}
}

// TestBuildPRBody_EmptyAll は両方空の場合に空文字列が返されることをテストします
func TestBuildPRBody_EmptyAll(t *testing.T) {
	issues := []int{}
	body := buildPRBody("", issues)
	expected := ""

	if body != expected {
		t.Errorf("buildPRBody() = %q, want %q", body, expected)
	}
}
