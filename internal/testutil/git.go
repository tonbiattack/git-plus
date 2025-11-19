// ================================================================================
// Package testutil - テスト用ユーティリティ
// ================================================================================
// このパッケージは、テストで使用する共通のヘルパー関数を提供します。
//
// 提供する機能:
// - GitRepo: テスト用の一時Gitリポジトリを作成・操作
// - ファイル作成、コミット、ブランチ操作などのユーティリティ
//
// 使用目的:
// Git Plus のコマンドをテストする際に、実際のGitリポジトリを
// 安全に作成・操作するためのヘルパーを提供します。
// ================================================================================
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// GitRepo はテスト用の一時Gitリポジトリを表す構造体
type GitRepo struct {
	Dir string // リポジトリのディレクトリパス
	t   *testing.T
}

// NewGitRepo は新しい一時Gitリポジトリを作成します。
//
// この関数は以下の処理を実行します：
// 1. 一時ディレクトリを作成（テスト終了時に自動削除）
// 2. git init を実行
// 3. user.email と user.name を設定
//
// 戻り値:
// - *GitRepo: テスト用リポジトリへのポインタ
//
// 使用例:
//
//	func TestMyCommand(t *testing.T) {
//	    repo := testutil.NewGitRepo(t)
//	    repo.CreateFile("README.md", "# Test")
//	    repo.Commit("Initial commit")
//	}
func NewGitRepo(t *testing.T) *GitRepo {
	t.Helper()

	// 一時ディレクトリを作成（テスト終了時に自動クリーンアップ）
	dir := t.TempDir()

	repo := &GitRepo{Dir: dir, t: t}

	// git init
	if _, err := repo.Git("init"); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// 基本設定
	cmds := [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		// コミット時の警告を抑制
		{"config", "init.defaultBranch", "main"},
		// テスト環境ではコミット署名を無効化
		{"config", "commit.gpgsign", "false"},
		{"config", "tag.gpgsign", "false"},
	}

	for _, args := range cmds {
		if _, err := repo.Git(args...); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	return repo
}

// Git はGitコマンドを実行し、出力を返します。
//
// パラメータ:
// - args: git コマンドの引数
//
// 戻り値:
// - string: コマンドの出力（stdout + stderr）
// - error: エラー（コマンド失敗時）
//
// 使用例:
//
//	output, err := repo.Git("status")
//	output, err := repo.Git("branch", "-a")
func (r *GitRepo) Git(args ...string) (string, error) {
	r.t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// MustGit はGitコマンドを実行し、エラー時にテストを失敗させます。
//
// パラメータ:
// - args: git コマンドの引数
//
// 戻り値:
// - string: コマンドの出力
//
// 注意: エラー時にはテストが即座に失敗します。
func (r *GitRepo) MustGit(args ...string) string {
	r.t.Helper()

	output, err := r.Git(args...)
	if err != nil {
		r.t.Fatalf("Git command failed: git %v\nOutput: %s\nError: %v", args, output, err)
	}
	return output
}

// CreateFile はリポジトリにファイルを作成します。
//
// パラメータ:
// - name: ファイル名（相対パス）
// - content: ファイルの内容
//
// サブディレクトリを含むパスも使用できます（例: "src/main.go"）。
// 必要なディレクトリは自動的に作成されます。
//
// 使用例:
//
//	repo.CreateFile("README.md", "# Test Project")
//	repo.CreateFile("src/main.go", "package main")
func (r *GitRepo) CreateFile(name, content string) {
	r.t.Helper()

	path := filepath.Join(r.Dir, name)

	// 親ディレクトリを作成
	dir := filepath.Dir(path)
	// パーミッションの意味:
	// - 0755 (8進数): 所有者に読み/書き/実行 (rwx=7)、
	//   グループとその他に読み/実行 (r-x=5) を許可します。
	//   ディレクトリに対しては実行ビット(x)が「入る/検索できる」権限を意味します。
	if err := os.MkdirAll(dir, 0755); err != nil {
		r.t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// ファイル作成時のパーミッション:
	// - 0644 (8進数): 所有者に読み/書き (rw=6)、
	//   グループとその他に読み (r=4) を許可します。
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		r.t.Fatalf("Failed to create file %s: %v", name, err)
	}
}

// Commit は現在の変更をコミットします。
//
// パラメータ:
// - message: コミットメッセージ
//
// この関数は以下を実行します：
// 1. git add . （全ての変更をステージング）
// 2. git commit -m message
//
// 使用例:
//
//	repo.CreateFile("file.txt", "content")
//	repo.Commit("Add file.txt")
func (r *GitRepo) Commit(message string) {
	r.t.Helper()

	r.MustGit("add", ".")
	r.MustGit("commit", "-m", message)
}

// CreateBranch は新しいブランチを作成します。
//
// パラメータ:
// - name: ブランチ名
//
// 注意: ブランチを作成しますが、切り替えは行いません。
// 切り替えが必要な場合は CheckoutBranch を使用してください。
func (r *GitRepo) CreateBranch(name string) {
	r.t.Helper()
	r.MustGit("branch", name)
}

// CheckoutBranch は指定したブランチに切り替えます。
//
// パラメータ:
// - name: ブランチ名
func (r *GitRepo) CheckoutBranch(name string) {
	r.t.Helper()
	r.MustGit("checkout", name)
}

// CreateAndCheckoutBranch は新しいブランチを作成して切り替えます。
//
// パラメータ:
// - name: ブランチ名
//
// git checkout -b と同等の操作を行います。
func (r *GitRepo) CreateAndCheckoutBranch(name string) {
	r.t.Helper()
	r.MustGit("checkout", "-b", name)
}

// CurrentBranch は現在のブランチ名を返します。
func (r *GitRepo) CurrentBranch() string {
	r.t.Helper()
	output := r.MustGit("branch", "--show-current")
	// 末尾の改行を削除
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
	return output
}

// CreateTag はタグを作成します。
//
// パラメータ:
// - name: タグ名
// - message: タグメッセージ（注釈付きタグ）
func (r *GitRepo) CreateTag(name, message string) {
	r.t.Helper()
	r.MustGit("tag", "-a", name, "-m", message)
}

// CreateLightweightTag は軽量タグを作成します。
//
// パラメータ:
// - name: タグ名
func (r *GitRepo) CreateLightweightTag(name string) {
	r.t.Helper()
	r.MustGit("tag", name)
}

// StashPush は現在の変更をスタッシュします。
//
// パラメータ:
// - message: スタッシュメッセージ（空文字列の場合はデフォルトメッセージ）
func (r *GitRepo) StashPush(message string) {
	r.t.Helper()
	if message == "" {
		r.MustGit("stash", "push")
	} else {
		r.MustGit("stash", "push", "-m", message)
	}
}

// StashPop は最新のスタッシュを適用します。
func (r *GitRepo) StashPop() {
	r.t.Helper()
	r.MustGit("stash", "pop")
}

// HasUncommittedChanges は未コミットの変更があるかどうかを返します。
func (r *GitRepo) HasUncommittedChanges() bool {
	r.t.Helper()
	output, _ := r.Git("status", "--porcelain")
	return len(output) > 0
}

// CommitCount はコミット数を返します。
func (r *GitRepo) CommitCount() int {
	r.t.Helper()
	output := r.MustGit("rev-list", "--count", "HEAD")
	var count int
	// 改行を含む可能性があるため、数字のみを抽出
	for _, c := range output {
		if c >= '0' && c <= '9' {
			count = count*10 + int(c-'0')
		}
	}
	return count
}

// Path はリポジトリ内のファイルの絶対パスを返します。
//
// パラメータ:
// - name: 相対ファイルパス
//
// 戻り値:
// - string: 絶対パス
func (r *GitRepo) Path(name string) string {
	return filepath.Join(r.Dir, name)
}

// ReadFile はリポジトリ内のファイルを読み込みます。
//
// パラメータ:
// - name: 相対ファイルパス
//
// 戻り値:
// - string: ファイルの内容
func (r *GitRepo) ReadFile(name string) string {
	r.t.Helper()
	path := filepath.Join(r.Dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		r.t.Fatalf("Failed to read file %s: %v", name, err)
	}
	return string(data)
}

// FileExists はファイルが存在するかどうかを返します。
//
// パラメータ:
// - name: 相対ファイルパス
//
// 戻り値:
// - bool: ファイルが存在する場合は true
func (r *GitRepo) FileExists(name string) bool {
	r.t.Helper()
	path := filepath.Join(r.Dir, name)
	_, err := os.Stat(path)
	return err == nil
}
