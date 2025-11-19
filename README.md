# Git Plus

Git の日常操作を少しだけ楽にするための拡張コマンド集です。元々 Bash で書いていたスクリプトを Go で書き直し、Cobra フレームワークを使用して単一のバイナリとして配布できるようにしました。

**このツールは `git` コマンドの拡張として動作します。** すべてのコマンドは `git newbranch`、`git pr-merge` のように、`git` コマンドのサブコマンドとして実行します。

## 特徴

- **Git の拡張コマンド**: `git-xxx` 形式のシンボリックリンクにより、`git xxx` として自然に呼び出せる
- **単一バイナリ**: Go で実装された単一のバイナリから、すべてのコマンドが利用可能
- **豊富なコマンド**: ブランチ、タグ、コミット、スタッシュ、PR、Issue などの多様なコマンド
- **対話的な操作**: 多くのコマンドが対話的な選択を提供し、安全で直感的に操作可能
- **GitHub CLI 連携**: プルリクエスト、Issue、リリースなどで GitHub CLI と連携

## コマンド一覧

コマンドはジャンル別に分類されています。詳細は各ドキュメントを参照してください。

### ブランチ操作

ブランチの作成、切り替え、削除、同期など。

- `git newbranch` - ブランチを削除して作り直し、トラッキングブランチとして設定
- `git rename-branch` - 現在のブランチ名を安全に変更し、--push でリモートも更新
- `git delete-local-branches` - マージ済みローカルブランチをまとめて削除
- `git recent` - 最近使用したブランチを時系列で表示して切り替え
- `git sync` - リモートのデフォルトブランチと同期（rebase使用）
- `git abort` - 進行中の rebase / merge / cherry-pick / revert を自動判定して中止

[詳細はこちら](doc/commands/branch.md)

### タグ操作

タグの作成、削除、チェックアウト、差分取得など。

- `git reset-tag` - タグをローカルとリモートから削除して再作成
- `git tag-diff` - 2つのタグ間の差分を取得してファイルに出力
- `git tag-diff-all` - 全タグ間の差分を一括取得してファイルに出力
- `git tag-checkout` - セマンティックバージョン順で最新タグをチェックアウト
- `git new-tag` - セマンティックバージョニングに従って新しいタグを自動生成

[詳細はこちら](doc/commands/tag.md)

### コミット操作

コミットの修正、スカッシュ、取り消し、トラッキング設定など。

- `git amend` - 直前のコミットを `git commit --amend` で再編集
- `git squash` - 直近の複数コミットを対話的にスカッシュ
- `git undo-last-commit` - 直近のコミットを取り消し（変更内容は残す）
- `git track` - トラッキングブランチを設定（リモートブランチがなければ自動プッシュ）

[詳細はこちら](doc/commands/commit.md)

### スタッシュ操作

スタッシュの整理、選択、一時保存と復元など。

- `git stash-cleanup` - 重複するスタッシュを自動削除
- `git stash-select` - スタッシュをインタラクティブに選択して操作
- `git pause` - 作業を一時保存してブランチを切り替え
- `git resume` - git pause で保存した作業を復元

[詳細はこちら](doc/commands/stash.md)

### プルリクエスト

プルリクエストの作成、マージ、チェックアウト、一覧表示など。

- `git pr-create-merge` - PR作成→マージ→ブランチ削除→最新取得を一気に実行
- `git pr-list` - プルリクエスト一覧を表示（`gh pr list` のラッパー）
- `git pr-merge` - プルリクエストをマージ（`gh pr merge` のラッパー）
- `git pr-checkout` - 最新または指定されたPRをチェックアウト
- `git pr-browse` - プルリクエストをブラウザで開く（`gh pr view --web` のラッパー）
- `git pr-issue-link` - PRとIssueを紐づけて作成（Closes #番号を自動追加）

[詳細はこちら](doc/commands/pull-request.md)

### リポジトリ管理

リポジトリの作成、クローン、ブラウザで開く、他人のリポジトリ一覧など。

- `git create-repository` - GitHubリポジトリの作成→クローン→VSCode起動を自動化
- `git clone-org` - GitHub組織のリポジトリを一括クローン
- `git batch-clone` - ファイルに記載されたリポジトリをまとめてクローン
- `git browse` - 現在のリポジトリをブラウザで開く
- `git repo-others` - ローカルにクローン済みの他人のリポジトリを一覧表示

[詳細はこちら](doc/commands/repository.md)

### Issue管理

GitHubのIssueの作成、編集、閲覧など。

- `git issue-list` - Issueの一覧から詳細表示・編集・コメント追加・クローズ・一括クローズ・新規作成を統合操作
- `git issue-create` - エディタでIssueを作成
- `git issue-edit` - Issueの一覧を表示して編集/閲覧/コメント追加/クローズ
- `git issue-bulk-close` - 複数のIssueを同じコメントで一括クローズ

[詳細はこちら](doc/commands/issue.md)

### リリース管理

GitHubリリースノートの自動生成など。

- `git release-notes` - 既存のタグからGitHubリリースノートを自動生成

[詳細はこちら](doc/commands/release.md)

### 統計・分析

リポジトリのステップ数やユーザーごとの貢献度など。

- `git step` - リポジトリ全体のステップ数とユーザーごとの貢献度を11の指標で集計

[詳細はこちら](doc/commands/stats.md)

### ワークツリー操作

git worktreeを使った並行開発の支援。

- `git worktree-new` - 新しいブランチをworktreeとして別ディレクトリに作成しVSCodeを開く
- `git worktree-switch` - 既存worktreeの一覧から選択してVSCodeを開く
- `git worktree-delete` - 既存worktreeの一覧から選択して削除

[詳細はこちら](doc/commands/worktree.md)

## インストール

### 推奨: リポジトリをクローンしてグローバルコマンドとして利用

`go install` がネットワーク環境やプロキシの影響で動作しないことがあるため、リポジトリを直接クローンして利用する方法を推奨します。

**1. リポジトリをクローン**

```bash
git clone https://github.com/tonbiattack/git-plus.git
cd git-plus
```

**2. セットアップスクリプトを実行（推奨）**

**Linux / macOS の場合:**

```bash
./setup.sh
```

**Windows (PowerShell) の場合:**

```powershell
.\setup.ps1
```

セットアップスクリプトは以下を自動的に行います：
- git-plus バイナリのビルド
- `~/bin`（Windows: `%USERPROFILE%\bin`）への配置
- 各コマンド用のシンボリックリンク作成（Linux/macOS）またはコピー作成（Windows）
  - `git-newbranch`, `git-reset-tag`, `git-amend` などのコマンド
- PATH環境変数への追加

これにより、`git newbranch`、`git step` などの形式でコマンドを呼び出せます。

**手動でビルドする場合:**

**Linux / macOS の場合:**

```bash
# ~/bin にビルド（~/bin が存在しない場合は作成）
mkdir -p ~/bin

# git-plus をビルド
go build -o ~/bin/git-plus .

# 各コマンド用のシンボリックリンクを作成
cd ~/bin
ln -s git-plus git-newbranch
ln -s git-plus git-rename-branch
ln -s git-plus git-reset-tag
ln -s git-plus git-amend
ln -s git-plus git-squash
ln -s git-plus git-track
ln -s git-plus git-delete-local-branches
ln -s git-plus git-undo-last-commit
ln -s git-plus git-tag-diff
ln -s git-plus git-tag-diff-all
ln -s git-plus git-tag-checkout
ln -s git-plus git-stash-cleanup
ln -s git-plus git-stash-select
ln -s git-plus git-recent
ln -s git-plus git-step
ln -s git-plus git-sync
ln -s git-plus git-pr-create-merge
ln -s git-plus git-pr-merge
ln -s git-plus git-pr-list
ln -s git-plus git-pause
ln -s git-plus git-resume
ln -s git-plus git-create-repository
ln -s git-plus git-new-tag
ln -s git-plus git-browse
ln -s git-plus git-pr-checkout
ln -s git-plus git-clone-org
ln -s git-plus git-batch-clone
ln -s git-plus git-issue-list
ln -s git-plus git-issue-create
ln -s git-plus git-issue-edit
ln -s git-plus git-issue-bulk-close
ln -s git-plus git-release-notes
ln -s git-plus git-repo-others
ln -s git-plus git-pr-browse
ln -s git-plus git-pr-issue-link
ln -s git-plus git-worktree-new
ln -s git-plus git-worktree-switch
ln -s git-plus git-worktree-delete

# PATHに追加（まだ追加していない場合）
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
# zsh を使用している場合は ~/.zshrc に追加
```

**Windows (PowerShell) の場合:**

```powershell
# ユーザーディレクトリ配下に bin フォルダを作成
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# git-plus をビルド
go build -o "$env:USERPROFILE\bin\git-plus.exe" .

# 各コマンド用のコピーを作成
$binPath = "$env:USERPROFILE\bin"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-newbranch.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-rename-branch.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-reset-tag.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-amend.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-squash.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-track.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-delete-local-branches.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-undo-last-commit.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-tag-diff.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-tag-diff-all.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-tag-checkout.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-stash-cleanup.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-stash-select.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-recent.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-step.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-sync.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-create-merge.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-merge.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-list.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-merge-pr.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pause.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-resume.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-create-repository.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-new-tag.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-browse.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-checkout.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-clone-org.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-batch-clone.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-issue-list.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-issue-create.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-issue-edit.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-issue-bulk-close.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-release-notes.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-repo-others.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-browse.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-issue-link.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-worktree-new.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-worktree-switch.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-worktree-delete.exe"

# PATHに追加（まだ追加していない場合）
# システム環境変数に追加する場合は管理者権限で実行
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$env:USERPROFILE\bin", "User")
}
# 現在のセッションで即座に利用するには、PowerShellを再起動するか以下を実行
$env:Path = [Environment]::GetEnvironmentVariable("Path", "User")
```

**3. 動作確認**

```bash
git newbranch -h
git step -h
git recent -h
```

**更新方法:**

```bash
cd git-plus
git pull
# 上記のビルドコマンドを再実行
```

### ローカルで開発・テストする方法

リポジトリをクローンしている場合は、グローバルにインストールせずにローカルでそのまま実行できます。

```bash
# プロジェクトディレクトリ内でビルド
go build -o ./bin/git-plus .

# 相対パスで実行
./bin/git-plus newbranch feature/awesome
./bin/git-plus reset-tag v1.2.3
```

開発中に動作を素早く試したい場合は `go run` も利用できます。

```bash
go run . newbranch feature/awesome
go run . reset-tag v1.2.3
go run . amend --no-edit
go run . squash 3
go run . track
go run . delete-local-branches
go run . undo-last-commit
go run . tag-diff V4.2.00.00 V4.3.00.00
go run . tag-diff-all --prefix=V4
go run . stash-cleanup
go run . stash-select
go run . recent
go run . step
go run . sync
go run . pr-create-merge
go run . pause main
go run . resume
go run . new-tag feature
go run . pr-checkout
go run . pr-checkout 123
```

Windows で PowerShell を利用している場合は、`./bin/git-plus` の代わりに `.\bin\git-plus.exe` のようにパスを指定してください。

## アンインストール

**Linux / macOS:**

```bash
# バイナリとシンボリックリンクを削除
rm ~/bin/git-plus
rm ~/bin/git-*

# リポジトリも削除する場合
rm -rf ~/path/to/git-plus
```

**Windows (PowerShell):**

```powershell
# バイナリとコピーを削除
Remove-Item "$env:USERPROFILE\bin\git-plus.exe"
Remove-Item "$env:USERPROFILE\bin\git-*.exe"

# リポジトリも削除する場合
Remove-Item -Recurse -Force "C:\path\to\git-plus"
```

## プロジェクト構成

### ディレクトリ構造

```
.
├── cmd/                    # Cobraコマンド定義
│   ├── root.go            # ルートコマンド
│   ├── branch/            # ブランチ操作コマンド
│   │   ├── back.go
│   │   ├── delete_local_branches.go
│   │   ├── newbranch.go
│   │   ├── recent.go
│   │   └── sync.go
│   ├── tag/               # タグ操作コマンド
│   │   ├── new_tag.go
│   │   ├── reset_tag.go
│   │   ├── tag_checkout.go
│   │   ├── tag_diff.go
│   │   └── tag_diff_all.go
│   ├── commit/            # コミット操作コマンド
│   │   ├── amend.go
│   │   ├── squash.go
│   │   ├── track.go
│   │   └── undo_last_commit.go
│   ├── stash/             # スタッシュ操作コマンド
│   │   ├── pause.go
│   │   ├── resume.go
│   │   ├── stash_cleanup.go
│   │   └── stash_select.go
│   ├── pr/                # プルリクエストコマンド
│   │   ├── pr_browse.go
│   │   ├── pr_checkout.go
│   │   ├── pr_create_merge.go
│   │   ├── pr_issue_link.go
│   │   ├── pr_list.go
│   │   └── pr_merge.go
│   ├── repo/              # リポジトリ管理コマンド
│   │   ├── batch_clone.go
│   │   ├── browse.go
│   │   ├── clone_org.go
│   │   ├── create_repository.go
│   │   └── repo_others.go
│   ├── issue/             # Issue管理コマンド
│   │   ├── issue_bulk_close.go
│   │   ├── issue_create.go
│   │   ├── issue_edit.go
│   │   └── issue_list.go
│   ├── release/           # リリース管理コマンド
│   │   └── release_notes.go
│   ├── stats/             # 統計・分析コマンド
│   │   └── step.go
│   └── worktree/          # ワークツリー操作コマンド
│       ├── worktree_delete.go
│       ├── worktree_new.go
│       └── worktree_switch.go
├── internal/              # 内部共通パッケージ
│   ├── gitcmd/           # Gitコマンド実行の共通ユーティリティ
│   ├── ui/               # UI関連のユーティリティ
│   └── pausestate/       # pause/resume状態管理
├── doc/                  # READMEや社内向けのコマンドリファレンス
│   └── commands/         # カテゴリ別ドキュメント
├── docs/                  # 公開リポジトリに同期されるドキュメント
│   └── commands/         # 公開用ドキュメント
├── main.go               # エントリーポイント
├── bin/                  # ビルド済みバイナリ
└── go.mod
```

補足:
- `doc/` 配下がこのリポジトリの一次ソースであり、READMEや開発作業から直接リンクされています。
- `docs/` は public リポジトリに同期される公開ドキュメントです。公開向けの調整が必要なときのみ編集します。

### 共通パッケージ

#### `internal/gitcmd`

Gitコマンドの実行を共通化するユーティリティパッケージです。

**主な関数:**

- `Run(args ...string) ([]byte, error)`
  - Gitコマンドを実行して出力を取得します
  - 例: `output, err := gitcmd.Run("branch", "--show-current")`

- `RunWithIO(args ...string) error`
  - Gitコマンドを実行し、標準入出力を親プロセスにリダイレクトします
  - ユーザー入力が必要な場合や、リアルタイム出力が必要な場合に使用
  - 例: `err := gitcmd.RunWithIO("rebase", "origin/main")`

- `RunQuiet(args ...string) error`
  - Gitコマンドを静かに実行します（出力は破棄）
  - 成功/失敗のみを確認したい場合に使用
  - 例: `err := gitcmd.RunQuiet("show-ref", "--verify", "--quiet", "refs/heads/main")`

- `IsExitError(err error, code int) bool`
  - エラーが特定の終了コードかどうかをチェックします
  - 例: `if gitcmd.IsExitError(err, 1) { /* ブランチが存在しない */ }`

**使用例:**

```go
package main

import (
    "fmt"
    "github.com/tonbiattack/git-plus/internal/gitcmd"
)

func main() {
    // ブランチ名を取得
    output, err := gitcmd.Run("branch", "--show-current")
    if err != nil {
        fmt.Println("エラー:", err)
        return
    }
    fmt.Println("現在のブランチ:", string(output))

    // ブランチを切り替え（ユーザーに出力を表示）
    if err := gitcmd.RunWithIO("switch", "main"); err != nil {
        fmt.Println("切り替え失敗:", err)
        return
    }
}
```

## 開発メモ

- Go 1.22 以降でのビルドを想定しています。
- Cobra フレームワークを使用した単一バイナリ構造です。
- ルートに `go.mod` を置き、各サブコマンドは機能別にサブパッケージ（`cmd/branch/`, `cmd/tag/` など）に配置しています。
- 共通処理は `internal/` パッケージに配置しています。
- 実行ファイル名が `git-xxx` の場合、自動的に `xxx` サブコマンドとして実行されます。
- 追加のコマンドを作成する場合は、該当する機能カテゴリのサブパッケージ（例: `cmd/branch/`）に新しい `.go` ファイルを作成し、`cmd.RootCmd` に登録してください。
- 新しいカテゴリを追加する場合は、`main.go` でサブパッケージをインポートすることを忘れないでください。

## GitHub CLI のインストール

一部のコマンド（PR、Issue、リリース管理など）は GitHub CLI を必要とします。

Windows (winget):
```powershell
winget install --id GitHub.cli
```

macOS (Homebrew):
```bash
brew install gh
```

Linux (Debian/Ubuntu):
```bash
sudo apt install gh
```

**認証方法:**
```bash
gh auth login
```

対話的に以下を選択:
1. GitHub.com を選択
2. HTTPS を選択
3. ブラウザで認証を選択
