package branch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tonbiattack/git-plus/internal/testutil"
)

// TestNormalizeAbortOperation はユーザー入力の正規化をテストします
//
// 詳細 (文法/意図):
//   - このテストは表形式 (table-driven) で複数の入力ケースを定義し、
//     各ケースごとに期待される正規化結果を検証します。
//   - `t.Parallel()` を呼んでいるため、サブテストは並列実行される可能性があります。
//     そのためサブテスト内で使用するループ変数を再バインドしてクロージャキャプチャ問題を避けています。
//   - テストの構造:
//     1. `tests` スライスに入力と期待値を列挙
//     2. for-range で各ケースを取り出し、`t.Run` でサブテストを作成
//     3. サブテスト内で `normalizeAbortOperation` を呼び出し、結果と期待値を比較
//
// 文法レベルのポイント:
// - table-driven テストは新しいケースを追加しやすく、期待値を明示しやすい。
// - 並列化 (`t.Parallel()`) はテスト速度の向上に寄与するが、クロージャのキャプチャに注意が必要。
func TestNormalizeAbortOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"merge", "merge"},
		{"Rebase", "rebase"},
		{"CHERRY-PICK", "cherry-pick"},
		{"cherry_pick", "cherry-pick"},
		{"cherry", "cherry-pick"},
		{"revert", "revert"},
	}

	for _, tt := range tests {
		// ループ変数を新しいローカル変数に再バインドします。
		// 理由:
		// - Go の for-range ではループ変数が再利用されるため、クロージャがその変数を参照すると
		//   並列実行時（t.Parallel()）にすべてのサブテストが同じ最終値を参照してしまう可能性があります。
		// - 各イテレーションごとに `tt := tt` で新しい変数に再バインドすることで、クロージャは
		//   そのイテレーション固有の値を捕捉し、並列サブテストでも安全に動作します。
		tt := tt

		// サブテスト: 各入力ケースを個別のサブテストとして実行します。
		// - 第一引数はテスト名（ここでは入力文字列）
		// - 無名関数内で再度 `t.Parallel()` を呼ぶことで各サブテスト自身も並列化されます。
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			// 実際の呼び出しと検証
			actual, err := normalizeAbortOperation(tt.input)
			if err != nil {
				t.Fatalf("normalizeAbortOperation(%q) returned error: %v", tt.input, err)
			}
			if actual != tt.expected {
				t.Fatalf("normalizeAbortOperation(%q) = %q, want %q", tt.input, actual, tt.expected)
			}
		})
	}

	if _, err := normalizeAbortOperation("unknown"); err == nil {
		t.Fatalf("normalizeAbortOperation should fail for unsupported operation")
	}
}

// TestDetectAbortOperation_Merge はMERGE_HEAD検出をテストします
func TestDetectAbortOperation_Merge(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "MERGE_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "merge" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "merge")
		}
	})
}

// TestDetectAbortOperation_Rebase は rebase ディレクトリ検出をテストします
func TestDetectAbortOperation_Rebase(t *testing.T) {
	withRepo(t, func(gitDir string) {
		createDir(t, filepath.Join(gitDir, "rebase-merge"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "rebase" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "rebase")
		}
	})
}

// TestDetectAbortOperation_Revert は REVERT_HEAD を検出することをテストします
func TestDetectAbortOperation_Revert(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "REVERT_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "revert" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "revert")
		}
	})
}

// TestDetectAbortOperation_CherryPick は CHERRY_PICK_HEAD を検出することをテストします
func TestDetectAbortOperation_CherryPick(t *testing.T) {
	withRepo(t, func(gitDir string) {
		writeIndicator(t, filepath.Join(gitDir, "CHERRY_PICK_HEAD"))
		op, err := detectAbortOperation()
		if err != nil {
			t.Fatalf("detectAbortOperation returned error: %v", err)
		}
		if op != "cherry-pick" {
			t.Fatalf("detectAbortOperation = %q, want %q", op, "cherry-pick")
		}
	})
}

// TestDetectAbortOperation_NoOp は操作が検出されない場合にエラーとなることをテストします
func TestDetectAbortOperation_NoOp(t *testing.T) {
	withRepo(t, func(_ string) {
		if _, err := detectAbortOperation(); err == nil {
			t.Fatalf("detectAbortOperation should fail when no operation is detected")
		}
	})
}

// withRepo は一時Gitリポジトリでコールバックを実行します
func withRepo(t *testing.T, fn func(gitDir string)) {
	// withRepo は一時的なテスト用 Git リポジトリを作成し、指定のコールバックを実行します。
	//
	// 文法/意図:
	// - テストで必要な前処理（リポジトリ作成、初期コミット）を集約することで各テストの冗長性を減らす。
	// - 作成したリポジトリのディレクトリにカレントディレクトリを移動してからコールバックを呼ぶ。
	//   これにより、`git` コマンドや `getGitDir` の動作が期待通りにリポジトリを参照できる。
	// - 終了時に `defer` を使って元の作業ディレクトリへ復帰させることで、テスト間の副作用を防止する。
	//
	// 実装の流れ:
	// 1. `testutil.NewGitRepo` で一時リポジトリを作成
	// 2. 簡単なファイルを作ってコミット（`git init` 後の最小セットアップ）
	// 3. 現在の作業ディレクトリを保存し、テスト用リポジトリへ `chdir`
	// 4. コールバックに `.git` ディレクトリのパスを渡す
	t.Helper()

	repo := testutil.NewGitRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial")

	// 元の作業ディレクトリを取得して保存
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// テスト終了後に元のディレクトリへ戻す（副作用のクリーンアップ）
	defer func() { _ = os.Chdir(oldDir) }()

	// テスト用リポジトリのルートへ移動する
	if err := os.Chdir(repo.Dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// コールバックには `.git` ディレクトリの絶対パスを渡す
	fn(filepath.Join(repo.Dir, ".git"))
}

// writeIndicator は指定されたパスに空ファイルを作成します
func writeIndicator(t *testing.T, path string) {
	// 文法/意図:
	// - Git の進行中操作判定では特定のファイル（例: MERGE_HEAD, CHERRY_PICK_HEAD）や
	//   ディレクトリ（rebase-merge 等）の存在を確認します。このヘルパーはその指標ファイルを
	//   テスト用に作成するためのものです。
	// - 第2引数 `path` は作成するインジケータファイルのパス（通常は <repo>/.git/<NAME>）。
	t.Helper()

	// ファイルを書き込む際のパーミッションはテスト内のみの利用なので 0o600 を指定しています。
	// - 所有者に読み/書き、他に権限なし（セキュリティ的に最小限の権限）
	if err := os.WriteFile(path, []byte("dummy"), 0o600); err != nil {
		t.Fatalf("failed to write indicator file: %v", err)
	}
}

// createDir は指定されたディレクトリを作成します
func createDir(t *testing.T, path string) {
	// 文法/意図:
	// - テスト内で rebase 等の進行中状態を模倣するためにディレクトリを作成するユーティリティ。
	// - `os.MkdirAll` を使うことで中間ディレクトリが存在しなくても確実に作成できます。
	// - パーミッション 0o755 は所有者に書き込みを許可し、グループ/その他に読み/実行を許可します。
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
}
