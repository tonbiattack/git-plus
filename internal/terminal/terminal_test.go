package terminal

import (
	"os/exec"
	"testing"
)

func TestSaveState(t *testing.T) {
	// This test just verifies that SaveState doesn't panic
	state := SaveState()
	if state == nil {
		t.Error("SaveState should return a non-nil State")
	}
}

func TestState_Restore(t *testing.T) {
	// This test just verifies that Restore doesn't panic
	state := SaveState()
	state.Restore()

	// Test with empty state
	emptyState := &State{}
	emptyState.Restore() // Should not panic
}

func TestResetToSane(t *testing.T) {
	// This test just verifies that ResetToSane doesn't panic
	ResetToSane()
}

func TestIsWaitRequiredEditor(t *testing.T) {
	tests := []struct {
		name       string
		editorName string
		expected   bool
	}{
		{"VSCode", "code", true},
		{"VSCode Insiders", "code-insiders", true},
		{"VSCodium", "codium", true},
		{"Sublime Text", "subl", true},
		{"Sublime Text Full", "sublime_text", true},
		{"Atom", "atom", true},
		{"Gedit", "gedit", true},
		{"Vim", "vim", false},
		{"Vi", "vi", false},
		{"Nano", "nano", false},
		{"Emacs", "emacs", false},
		{"VSCode with path", "/usr/bin/code", true},
		{"Vim with path", "/usr/bin/vim", false},
		{"VSCode Windows path", "C:\\Program Files\\VSCode\\code.exe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWaitRequiredEditor(tt.editorName)
			if result != tt.expected {
				t.Errorf("IsWaitRequiredEditor(%q) = %v, want %v", tt.editorName, result, tt.expected)
			}
		})
	}
}

func TestAddWaitFlagIfNeeded(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		expected string
	}{
		{"VSCode without flag", "code", "code --wait "},
		{"VSCode with --wait", "code --wait", "code --wait"},
		{"VSCode with -w", "code -w", "code -w"},
		{"VSCode with other args", "code --new-window", "code --wait --new-window"},
		{"Vim no change", "vim", "vim"},
		{"Nano no change", "nano", "nano"},
		{"Empty string", "", ""},
		{"Sublime without flag", "subl", "subl --wait "},
		{"Sublime with flag", "subl --wait", "subl --wait"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddWaitFlagIfNeeded(tt.editor)
			if result != tt.expected {
				t.Errorf("AddWaitFlagIfNeeded(%q) = %q, want %q", tt.editor, result, tt.expected)
			}
		})
	}
}

func TestRunEditorWithProtection_SuccessfulCommand(t *testing.T) {
	// Test with a simple command that exits successfully
	cmd := exec.Command("true")
	result := RunEditorWithProtection(cmd)

	if result.Cancelled {
		t.Error("Expected Cancelled to be false for successful command")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
}

func TestRunEditorWithProtection_FailingCommand(t *testing.T) {
	// Test with a command that exits with error
	cmd := exec.Command("false")
	result := RunEditorWithProtection(cmd)

	if result.Cancelled {
		t.Error("Expected Cancelled to be false for failed command")
	}
	if result.Error == nil {
		t.Error("Expected error for failed command")
	}
}

func TestRunEditorWithProtection_NonExistentCommand(t *testing.T) {
	// Test with a command that doesn't exist
	cmd := exec.Command("nonexistent-command-12345")
	result := RunEditorWithProtection(cmd)

	if result.Error == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestEditorResult_Structure(t *testing.T) {
	// Test that EditorResult has the expected fields
	result := &EditorResult{
		Cancelled: true,
		Error:     nil,
	}

	if !result.Cancelled {
		t.Error("Expected Cancelled to be true")
	}
	if result.Error != nil {
		t.Error("Expected Error to be nil")
	}
}
