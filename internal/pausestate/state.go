// ================================================================================
// Package pausestate - Git Pause/Resume状態管理
// ================================================================================
// このパッケージは、git pauseコマンドとgit resumeコマンドの状態管理を提供します。
//
// 機能:
// git pauseコマンドで作業を一時保存した際の状態を、
// ~/.git-plus/pause-state.json に保存・読み込み・削除します。
//
// 保存される情報:
// - FromBranch: pause前のブランチ名
// - ToBranch: pauseで切り替えた先のブランチ名
// - StashRef: スタッシュの参照名
// - StashMessage: スタッシュメッセージ
// - Timestamp: pause実行日時
//
// ユースケース:
// 1. git pause: 現在の作業をスタッシュして別ブランチに切り替え → 状態を保存
// 2. git resume: 保存された状態を読み込んで元のブランチに戻り、スタッシュを適用
//
// ファイル構造:
// ~/.git-plus/pause-state.json (JSON形式で状態を保存)
// ================================================================================
package pausestate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PauseState は、git pauseコマンドで保存される状態を保持する構造体です。
//
// フィールド:
// - FromBranch: pause実行前のブランチ名（復元時にこのブランチに戻る）
// - ToBranch: pause実行後に切り替えたブランチ名
// - StashRef: 作成されたスタッシュの参照名（例: stash@{0}）
// - StashMessage: スタッシュに付けられたメッセージ
// - Timestamp: pause実行日時（いつpauseしたかを記録）
type PauseState struct {
	FromBranch   string    `json:"from_branch"`    // pause前のブランチ名
	ToBranch     string    `json:"to_branch"`      // pause後のブランチ名
	StashRef     string    `json:"stash_ref"`      // スタッシュ参照名
	StashMessage string    `json:"stash_message"`  // スタッシュメッセージ
	Timestamp    time.Time `json:"timestamp"`      // pause実行日時
}

// getStateFilePath は、pause状態を保存するJSONファイルのパスを取得します。
//
// この関数は、以下の処理を実行します：
// 1. ユーザーのホームディレクトリを取得
// 2. ~/.git-plus ディレクトリを作成（存在しない場合）
// 3. ~/.git-plus/pause-state.json のパスを返す
//
// 戻り値:
// - string: 状態ファイルのフルパス
// - error: ホームディレクトリの取得失敗、またはディレクトリ作成失敗
//
// パーミッション:
// .git-plusディレクトリは 0755 (rwxr-xr-x) で作成されます。
func getStateFilePath() (string, error) {
	// ユーザーのホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	// ~/.git-plus ディレクトリのパスを構築
	gitPlusDir := filepath.Join(homeDir, ".git-plus")

	// .git-plus ディレクトリが存在しない場合は作成
	// MkdirAll は既に存在する場合でもエラーを返しません
	if err := os.MkdirAll(gitPlusDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリの作成に失敗 %s: %w", gitPlusDir, err)
	}

	// 状態ファイルのフルパスを返す
	return filepath.Join(gitPlusDir, "pause-state.json"), nil
}

// Save は、pause状態をJSONファイルに保存します。
//
// git pauseコマンド実行時に呼び出され、現在の作業状態を保存します。
// JSONファイルはインデント付きで保存されるため、人間が読みやすい形式になります。
//
// パラメータ:
// - state: 保存するPauseState構造体のポインタ
//
// 戻り値:
// - error: ファイルパス取得失敗、JSONエンコード失敗、ファイル書き込み失敗
//
// ファイルパーミッション:
// 保存されるJSONファイルは 0644 (rw-r--r--) で作成されます。
//
// 使用例:
//
//	state := &pausestate.PauseState{
//	    FromBranch: "feature-xxx",
//	    ToBranch: "main",
//	    StashRef: "stash@{0}",
//	    StashMessage: "WIP on feature-xxx",
//	    Timestamp: time.Now(),
//	}
//	err := pausestate.Save(state)
func Save(state *PauseState) error {
	// 状態ファイルのパスを取得
	filePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	// PauseState構造体をインデント付きJSONに変換
	// インデントは2スペース
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON のエンコードに失敗: %w", err)
	}

	// JSONデータをファイルに書き込み
	// パーミッション: 0644 (所有者: 読み書き, グループ: 読み取り, その他: 読み取り)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("ファイルの書き込みに失敗 %s: %w", filePath, err)
	}

	return nil
}

// Load は、保存されたpause状態をJSONファイルから読み込みます。
//
// git resumeコマンド実行時に呼び出され、保存された作業状態を復元します。
//
// 戻り値:
// - *PauseState: 読み込まれた状態（ファイルが存在しない場合はnil）
// - error: ファイルパス取得失敗、ファイル読み込み失敗、JSONデコード失敗
//
// 注意:
// ファイルが存在しない場合、エラーではなく (nil, nil) を返します。
// これにより、呼び出し側でpauseされていない状態を判別できます。
//
// 使用例:
//
//	state, err := pausestate.Load()
//	if err != nil {
//	    return err
//	}
//	if state == nil {
//	    fmt.Println("保存されたpause状態がありません")
//	    return nil
//	}
func Load() (*PauseState, error) {
	// 状態ファイルのパスを取得
	filePath, err := getStateFilePath()
	if err != nil {
		return nil, err
	}

	// JSONファイルを読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		// ファイルが存在しない場合は nil を返す（エラーではない）
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("ファイルの読み込みに失敗 %s: %w", filePath, err)
	}

	// JSONデータをPauseState構造体にデコード
	var state PauseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("JSON のデコードに失敗: %w", err)
	}

	return &state, nil
}

// Delete は、保存されたpause状態ファイルを削除します。
//
// git resumeコマンドが正常に完了した後に呼び出され、
// 使用済みの状態ファイルをクリーンアップします。
//
// 戻り値:
// - error: ファイルパス取得失敗、ファイル削除失敗
//
// 注意:
// ファイルが存在しない場合はエラーを返しません。
// これにより、冪等性（何度実行しても同じ結果）が保証されます。
//
// 使用例:
//
//	if err := pausestate.Delete(); err != nil {
//	    fmt.Println("警告: 状態ファイルの削除に失敗しました:", err)
//	}
func Delete() error {
	// 状態ファイルのパスを取得
	filePath, err := getStateFilePath()
	if err != nil {
		return err
	}

	// ファイルを削除
	if err := os.Remove(filePath); err != nil {
		// ファイルが存在しない場合はエラーを返さない
		if !os.IsNotExist(err) {
			return fmt.Errorf("ファイルの削除に失敗 %s: %w", filePath, err)
		}
	}

	return nil
}

// Exists は、pause状態ファイルが存在するかどうかを確認します。
//
// git resumeコマンド実行前に呼び出され、
// 復元可能な状態が保存されているかをチェックします。
//
// 戻り値:
// - bool: true = ファイルが存在する, false = ファイルが存在しない
// - error: ファイルパス取得失敗、Stat失敗（存在確認以外のエラー）
//
// 使用例:
//
//	exists, err := pausestate.Exists()
//	if err != nil {
//	    return err
//	}
//	if !exists {
//	    fmt.Println("保存されたpause状態がありません")
//	    return nil
//	}
func Exists() (bool, error) {
	// 状態ファイルのパスを取得
	filePath, err := getStateFilePath()
	if err != nil {
		return false, err
	}

	// ファイルの存在を確認
	_, err = os.Stat(filePath)
	if err != nil {
		// ファイルが存在しない場合は false を返す（エラーではない）
		if os.IsNotExist(err) {
			return false, nil
		}
		// その他のエラーの場合はエラーを返す
		return false, err
	}

	// ファイルが存在する
	return true, nil
}
