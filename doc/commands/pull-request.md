# プルリクエストコマンド

プルリクエストの作成、マージ、チェックアウト、一覧表示などのプルリクエスト管理に関するコマンドです。

## git pr-create-merge

PRの作成からマージ、ブランチ削除、最新の変更取得までを一気に実行します。GitHub CLIを使用した自動化コマンドです。

```bash
git pr-create-merge [ベースブランチ名]
git pr-create-merge -h              # ヘルプを表示
```

GitHub CLIを使用して以下の処理を自動化します:

1. タイトル・本文なしでPRを作成（`--fill`オプション使用）
2. PRをマージしてブランチを削除（`--merge --delete-branch --auto`）
3. ベースブランチに切り替え（`git switch`）
4. 最新の変更を取得（`git pull`）

**引数:**
- `ベースブランチ名` (省略可): マージ先のブランチ名（省略時は対話的に入力、デフォルト: main）

**使用例:**

```bash
# feature-branchで作業中
git add .
git commit -m "Add new feature"
git push

# 方法1: ベースブランチを引数で指定
git pr-create-merge main

# 方法2: 対話的に入力
git pr-create-merge
# マージ先のベースブランチを入力してください (デフォルト: main): develop

# 確認プロンプト
# PRを作成してマージしますか？ (y/N): y

# 以下が自動実行される:
# [1/5] PRを作成しています...
# ✓ PRを作成しました
# [2/5] PRをマージしてブランチを削除しています...
# ✓ PRをマージしてブランチを削除しました
# [3/5] ブランチ 'main' に切り替えています...
# ✓ ブランチ 'main' に切り替えました
# [4/5] 最新の変更を取得しています...
# ✓ 最新の変更を取得しました
# ✓ すべての処理が完了しました！
```

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- リモートリポジトリへのプッシュ権限があること

**注意事項:**
- `--fill`オプションを使用するため、PRのタイトルと本文は最新のコミットメッセージから自動生成されます
- `--auto`オプションを使用するため、ステータスチェックが通過すると自動的にマージされます
- マージ後、リモートのブランチは自動的に削除されます

## git pr-list

プルリクエスト一覧を表示します。`gh pr list` のラッパーで、状態やラベルでフィルタリングできます。

```bash
git pr-list [オプション]
git pr-list -h               # ヘルプを表示
```

**主な機能:**
- **PR一覧の表示**: リポジトリのプルリクエストを一覧表示
- **フィルタリング**: 状態、作成者、アサイン先、ラベル、ブランチなどでフィルタリング可能
- **柔軟な出力形式**: テーブル形式、JSON、カスタムテンプレートなどで出力
- **すべてのオプションをサポート**: `gh pr list` のすべてのオプションがそのまま使用可能

**使用例:**

```bash
# 基本的な使い方
git pr-list                      # オープンなPRの一覧を表示

# 状態でフィルタリング
git pr-list --state open         # オープンなPRのみ表示（デフォルト）
git pr-list --state closed       # クローズされたPRのみ表示
git pr-list --state merged       # マージされたPRのみ表示
git pr-list --state all          # すべてのPRを表示

# 作成者やアサイン先でフィルタリング
git pr-list --author @me         # 自分が作成したPRを表示
git pr-list --assignee @me       # 自分にアサインされたPRを表示
git pr-list --author octocat     # 特定ユーザーが作成したPRを表示

# ラベルやブランチでフィルタリング
git pr-list --label bug          # "bug" ラベルが付いたPRを表示
git pr-list --label "help wanted" # "help wanted" ラベルが付いたPRを表示
git pr-list --base main          # mainブランチへのPRを表示
git pr-list --head feature-123   # feature-123ブランチからのPRを表示

# 表示件数の制限
git pr-list --limit 10           # 最新10件のPRを表示
git pr-list --limit 50           # 最新50件のPRを表示

# JSON形式で出力
git pr-list --json number,title,state,author
git pr-list --json number,title,url --jq '.[].url'

# ブラウザで開く
git pr-list --web                # ブラウザでPR一覧を開く

# 複数のオプションを組み合わせ
git pr-list --state merged --author @me --limit 20
```

**サポートされるオプション:**

| オプション | 説明 |
|----------|------|
| `--state <state>` | PR の状態でフィルタ（`open`, `closed`, `merged`, `all`） |
| `--author <user>` | 作成者でフィルタ（`@me` で自分のPR） |
| `--assignee <user>` | アサイン先でフィルタ（`@me` で自分がアサインされたPR） |
| `--label <name>` | ラベルでフィルタ |
| `--limit <int>` | 表示件数を制限（デフォルト: 30） |
| `--base <branch>` | ベースブランチでフィルタ |
| `--head <branch>` | ヘッドブランチでフィルタ |
| `--json <fields>` | JSON形式で出力 |
| `--jq <expression>` | jq式でフィルタ |
| `--template <string>` | Go template形式で出力 |
| `--web` | ブラウザで開く |

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- GitHubリポジトリのPRにアクセスできること

**注意事項:**
- すべての `gh pr list` のオプションがそのまま使用できます
- デフォルトでは最新30件のオープンなPRが表示されます
- `--limit` オプションで表示件数を変更できます

## git pr-merge

プルリクエストをマージします。`gh pr merge` のラッパーで、デフォルトでマージコミットを作成し、ブランチを削除します。

```bash
git pr-merge [PR番号] [オプション]
git pr-merge -h              # ヘルプを表示
```

**デフォルトの動作:**
- **マージコミットで直接実行**: 対話なしでマージコミットを作成（`--merge` が自動適用）
- **ブランチ自動削除**: マージ後にブランチを自動削除（`--delete-branch` が自動適用）

**主な機能:**
- **マージ方法の選択**: merge commit（デフォルト）、squash、rebase から選択可能
- **すべてのオプションをサポート**: gh pr merge のすべてのオプションがそのまま使用できます

**使用例:**

```bash
# カレントブランチのPRをマージコミットで直接マージ（ブランチも削除）
git pr-merge

# PR番号を指定してマージコミットで直接マージ（ブランチも削除）
git pr-merge 89

# スカッシュマージで直接マージ（ブランチも削除）
git pr-merge --squash

# リベースマージで直接マージ（ブランチも削除）
git pr-merge --rebase

# 複数のオプションを組み合わせ
git pr-merge 89 --squash --auto

# 自動マージ（ステータスチェック通過後に自動マージ）
git pr-merge --auto
```

**サポートされるオプション:**
- `--merge`: マージコミットを作成（デフォルト）
- `--squash`: スカッシュマージ（デフォルトを上書き）
- `--rebase`: リベースマージ（デフォルトを上書き）
- `--delete-branch`: マージ後にブランチを削除（デフォルト）
- `--auto`: ステータスチェック通過後に自動マージ
- `--body <text>`: マージコミットのボディ
- `--subject <text>`: マージコミットのサブジェクト

その他のオプションについては `gh pr merge --help` を参照してください。

**git pr-create-merge との違い:**

| コマンド | 用途 |
|---------|------|
| `git pr-create-merge` | PR作成→マージ→ブランチ切り替え→pull の一連の流れを自動化 |
| `git pr-merge` | 既存のPRをマージコミットで直接マージ（`gh pr merge` のラッパー） |

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- マージ権限があること

**注意事項:**
- 引数なしで実行すると、カレントブランチに関連するPRをマージコミットで直接マージします
- デフォルトでブランチが削除されるため、マージ後はローカル・リモート両方でブランチが削除されます

## git pr-checkout

最新または指定されたプルリクエストをチェックアウトします。現在の作業を自動保存し、git resumeで復元できます。

```bash
git pr-checkout              # 最新のPRをチェックアウト
git pr-checkout 123          # PR #123 をチェックアウト
git pr-checkout -h           # ヘルプを表示
```

**主な機能:**
- **最新PR自動取得**: 引数なしで実行すると、最新のオープンなPRを自動的にチェックアウトします。
- **指定PR取得**: PR番号を指定すると、そのPRをチェックアウトします。
- **作業の自動保存**: 現在の変更を自動的にスタッシュし、git pause と同じ仕組みで状態を保存します。
- **簡単に元に戻せる**: git resume で元のブランチと作業内容を復元できます。

**使用例:**

```bash
# 最新のPRをチェックアウト
git pr-checkout
# 最新のPRを取得中...
# 最新のPR #123 をチェックアウトします
# 変更を保存中...
# ✓ 変更を保存しました: abc123...
# PR #123 をチェックアウト中...
# ✓ PR #123 のブランチ 'feature/awesome' にチェックアウトしました
#
# 元のブランチに戻るには: git resume

# 特定のPRをチェックアウト
git pr-checkout 456

# PRの確認・テストが完了したら元のブランチに戻る
git resume
```

**動作:**
1. GitHub CLIを使用してPR情報を取得（引数なしの場合は最新のPRを取得）
2. 現在のブランチと変更を確認
3. 変更があればスタッシュに保存
4. 状態を保存（`~/.git-plus/pause-state.json`）
5. `gh pr checkout` でPRブランチをチェックアウト
6. git resume で元のブランチと変更を復元可能に

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- GitHubリポジトリのPRにアクセスできること

**注意事項:**
- 既に pause 状態の場合は上書き確認が表示されます
- チェックアウト後は `git resume` で元のブランチに戻ることができます
- PRブランチで作業した内容は通常通りコミット・プッシュできます

## git pr-browse

プルリクエストをブラウザで開きます。`gh pr view --web` のラッパーです。

```bash
git pr-browse [PR番号] [オプション]
git pr-browse -h             # ヘルプを表示
```

**使用例:**

```bash
# カレントブランチのPRをブラウザで開く
git pr-browse

# PR番号を指定してブラウザで開く
git pr-browse 123
```

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること

## git pr-issue-link

PRを作成する際にGitHubのIssueと紐づけます。PRの説明欄に「Closes #番号」を自動的に追加することで、PRがマージされた際に関連するIssueが自動的にクローズされます。

```bash
git pr-issue-link [オプション]
git pr-issue-link -h                    # ヘルプを表示
```

**主な機能:**
- **Issue一覧表示**: オープンなIssueの一覧を表示し、対話的に選択可能
- **複数Issue対応**: 複数のIssueを同時に紐づけることが可能
- **自動クローズ**: PRマージ時に紐づけられたIssueが自動的にクローズ
- **柔軟な指定方法**: 対話的選択またはコマンドラインオプションで指定

**使用例:**

```bash
# 対話的にIssueを選択してPR作成
git pr-issue-link

# ベースブランチを指定
git pr-issue-link --base main
git pr-issue-link -b develop

# Issue番号を直接指定
git pr-issue-link --issue 123
git pr-issue-link -i 42

# 複数のIssueを指定（カンマ区切り）
git pr-issue-link --issue 123,456,789

# タイトルを指定
git pr-issue-link --title "Fix authentication bug"
git pr-issue-link -t "Add new feature"

# 本文を指定（Closes #番号は自動追加）
git pr-issue-link --body "This PR fixes the login issue"

# 複数オプションを組み合わせ
git pr-issue-link -b main -i 42 -t "Fix #42: Authentication bug"
```

**対話的な操作例:**

```bash
git pr-issue-link

# 現在のブランチ: feature/fix-auth
# マージ先のベースブランチを入力してください (デフォルト: main):
# ベースブランチ: main

# オープンなIssue一覧 (3 個):

# 1. #42: Authentication fails on mobile
#    Users report login issues...

# 2. #38: Performance improvement needed
#    The API response time...

# 3. #35: Add dark mode support
#    Users have requested...

# 紐づけるIssueを選択してください:
#   番号: 単一のIssueを選択
#   1,3,5: 複数のIssueを選択（カンマ区切り）
#   all: すべてのIssueを選択
#   none: Issueを紐づけない

# 入力: 1

# PRのタイトルを入力してください (空の場合はコミットから自動生成): Fix authentication on mobile

# ========================================
# ベースブランチ: main
# ヘッドブランチ: feature/fix-auth
# タイトル: Fix authentication on mobile
# 紐づけるIssue: #42

# --- PR本文 ---
# Closes #42
# ========================================

# PRを作成しますか？ (Y/n): y

# PRを作成しています...

# ✓ PRを作成しました
# URL: https://github.com/user/repo/pull/89

# 紐づけられたIssue:
#   - Issue #42 (PRマージ時に自動クローズされます)
```

**オプション:**

| オプション | 短縮形 | 説明 |
|----------|--------|------|
| `--base <branch>` | `-b` | マージ先のベースブランチ |
| `--issue <numbers>` | `-i` | 紐づけるIssue番号（カンマ区切りで複数指定可能） |
| `--title <text>` | `-t` | PRのタイトル |
| `--body <text>` | - | PRの本文（Closes #番号は自動追加） |

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- リモートリポジトリへのプッシュ権限があること

**注意事項:**
- PRの本文には自動的に「Closes #番号」が追加されます
- GitHubの仕様により、PRがデフォルトブランチにマージされると紐づけられたIssueが自動的にクローズされます
- 複数のIssueを紐づける場合、すべてのIssueが自動クローズされます
- タイトルを指定しない場合は、`--fill`オプションと同様にコミットメッセージから自動生成されます
