package pausestate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PauseState は git pause の状態を保持する構造体
type PauseState struct {
	FromBranch   string    `json:"from_branch"`
	ToBranch     string    `json:"to_branch"`
	StashRef     string    `json:"stash_ref"`
	StashMessage string    `json:"stash_message"`
	Timestamp    time.Time `json:"timestamp"`
}

// getStateFilePath は状態ファイルのパスを取得
func getStateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	gitPlusDir := filepath.Join(homeDir, ".git-plus")

	// .git-plus ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(gitPlusDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリの作成に失敗 %s: %w", gitPlusDir, err)
	}

	return filepath.Join(gitPlusDir, "pause-state.json"), nil
}

// Save は状態をファイルに保存
func Save(state *PauseState) error {
	filePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON のエンコードに失敗: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗 %s: %w", filePath, err)
	}

	return nil
}

// Load は状態をファイルから読み込み
func Load() (*PauseState, error) {
	filePath, err := getStateFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 状態ファイルが存在しない場合は nil を返す
		}
		return nil, fmt.Errorf("ファイルの読み込みに失敗 %s: %w", filePath, err)
	}

	var state PauseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("JSON のデコードに失敗: %w", err)
	}

	return &state, nil
}

// Delete は状態ファイルを削除
func Delete() error {
	filePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("ファイルの削除に失敗 %s: %w", filePath, err)
		}
	}

	return nil
}

// Exists は状態ファイルが存在するか確認
func Exists() (bool, error) {
	filePath, err := getStateFilePath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
