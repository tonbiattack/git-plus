// Package terminal provides utilities for managing terminal state and signal handling.
package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// State represents the saved terminal state.
type State struct {
	settings string
	saved    bool
}

// SaveState saves the current terminal state using stty.
// Returns a State that can be used to restore the terminal later.
func SaveState() *State {
	state := &State{}

	// Save current terminal settings
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err == nil {
		state.settings = strings.TrimSpace(string(output))
		state.saved = true
	}

	return state
}

// Restore restores the terminal to its saved state.
func (s *State) Restore() {
	if !s.saved || s.settings == "" {
		return
	}

	cmd := exec.Command("stty", s.settings)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// ResetToSane resets the terminal to a sane state.
// This is useful when the terminal is left in an unknown state.
func ResetToSane() {
	cmd := exec.Command("stty", "sane")
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// EditorResult represents the result of an editor operation.
type EditorResult struct {
	Cancelled bool   // User cancelled the operation (e.g., Ctrl+C)
	Error     error  // Error that occurred during the operation
}

// RunEditorWithProtection runs an editor command with terminal state protection.
// It saves the terminal state before running the editor and restores it after,
// even if the editor is interrupted by a signal.
func RunEditorWithProtection(editorCmd *exec.Cmd) *EditorResult {
	result := &EditorResult{}

	// Save terminal state
	termState := SaveState()

	// Set up signal handler for graceful cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to receive editor completion
	done := make(chan error, 1)

	// Run editor in a goroutine
	go func() {
		done <- editorCmd.Run()
	}()

	// Wait for either editor completion or signal
	select {
	case err := <-done:
		// Editor finished normally
		signal.Stop(sigChan)
		close(sigChan)

		if err != nil {
			// Check if the error indicates the process was killed/interrupted
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.Signaled() {
						result.Cancelled = true
						result.Error = fmt.Errorf("エディタが中断されました")
					} else if status.ExitStatus() != 0 {
						// Non-zero exit code might indicate cancellation
						result.Error = err
					}
				}
			} else {
				result.Error = err
			}
		}

	case sig := <-sigChan:
		// Signal received, kill the editor process
		signal.Stop(sigChan)
		close(sigChan)

		if editorCmd.Process != nil {
			// Send the same signal to the editor process
			_ = editorCmd.Process.Signal(sig)
			// Wait for the process to finish
			<-done
		}

		result.Cancelled = true
		result.Error = fmt.Errorf("操作がキャンセルされました (signal: %v)", sig)
	}

	// Always restore terminal state
	termState.Restore()

	// Additional safety: reset to sane state if something went wrong
	if result.Cancelled {
		ResetToSane()
	}

	return result
}

// IsWaitRequiredEditor checks if the editor requires a --wait flag.
// Some editors like VSCode open and return immediately unless --wait is specified.
func IsWaitRequiredEditor(editorName string) bool {
	waitEditors := []string{
		"code",
		"code-insiders",
		"codium",
		"vscodium",
		"subl",
		"sublime_text",
		"atom",
		"gedit",
	}

	baseName := strings.ToLower(editorName)
	// Extract base name if it's a path
	if idx := strings.LastIndex(baseName, "/"); idx >= 0 {
		baseName = baseName[idx+1:]
	}
	if idx := strings.LastIndex(baseName, "\\"); idx >= 0 {
		baseName = baseName[idx+1:]
	}

	for _, editor := range waitEditors {
		if strings.Contains(baseName, editor) {
			return true
		}
	}

	return false
}

// AddWaitFlagIfNeeded adds --wait flag to the editor command if needed.
// Returns the modified editor string.
func AddWaitFlagIfNeeded(editor string) string {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return editor
	}

	// Check if this editor needs --wait
	if !IsWaitRequiredEditor(parts[0]) {
		return editor
	}

	// Check if --wait or -w is already present
	hasWait := false
	for _, part := range parts {
		if part == "--wait" || part == "-w" {
			hasWait = true
			break
		}
	}

	if !hasWait {
		// Add --wait after the editor name
		return parts[0] + " --wait " + strings.Join(parts[1:], " ")
	}

	return editor
}
