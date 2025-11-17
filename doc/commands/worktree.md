# ワークツリー操作コマンド

git worktreeを使った並行開発を支援するコマンドです。複数のタスクを同時に作業する際に便利です。

## git worktree-new

新しいブランチをworktreeとして別ディレクトリに作成し、VSCodeを開きます。

```bash
git worktree-new feature/new-login
git worktree-new bugfix/issue-123 --no-code
git worktree-new feature/api-v2 --base develop
git worktree-new -h                              # ヘルプを表示
```

**動作:**
1. 指定されたブランチ名から自動的にディレクトリ名を生成します。
   - `feature/xxx` → `../repo-feature-xxx`
   - `bugfix/abc` → `../repo-bugfix-abc`
2. ブランチが既に存在する場合は、そのブランチを使用してworktreeを作成します。
3. ブランチが存在しない場合は、新しいブランチを作成してworktreeを作成します。
4. デフォルトでVSCodeを開きます（`--no-code`で無効化可能）。

**オプション:**
- `--no-code`: VSCodeを開かない
- `--base <branch>`: ベースブランチを指定（デフォルトは現在のブランチ）
- `-h, --help`: ヘルプを表示

## git worktree-switch

既存のworktree一覧を表示し、インタラクティブに選択してVSCodeを開きます。

```bash
git worktree-switch
git worktree-switch --no-code
git worktree-switch -h                           # ヘルプを表示
```

**動作:**
1. 既存のworktree一覧を表示します（現在のworktreeは除外）。
2. 各worktreeのパス、ブランチ名、コミットハッシュを表示します。
3. 番号を入力することで、選択したworktreeのディレクトリでVSCodeを開きます。
4. 空入力でキャンセルできます。

**オプション:**
- `--no-code`: VSCodeを開かない
- `-h, --help`: ヘルプを表示

## git worktree-delete

既存のworktree一覧を表示し、インタラクティブに選択して削除します。

```bash
git worktree-delete
git worktree-delete --force
git worktree-delete -h                           # ヘルプを表示
```

**動作:**
1. 削除可能なworktree一覧を表示します（現在のworktreeは除外）。
2. 各worktreeのパス、ブランチ名、コミットハッシュを表示します。
3. 番号を入力し、確認プロンプトで`y`または`yes`を入力すると削除します。
4. 削除後もブランチ自体は保持されます。ブランチも削除する場合は`git branch -d <branch>`を実行してください。

**オプション:**
- `--force`: 未コミットの変更があっても強制削除
- `-h, --help`: ヘルプを表示

**注意事項:**
- 未コミットの変更がある場合は通常の削除が失敗します。`--force`を使用してください。
- 削除されるのはworktreeのディレクトリのみで、ブランチは保持されます。

## 使用例

### 複数の機能を並行開発する

```bash
# メインリポジトリで新機能Aのworktreeを作成
git worktree-new feature/user-auth

# VSCodeが開き、別ディレクトリで作業開始
# ../repo-feature-user-auth/ で作業

# 別の新機能Bも並行して作業したい場合
git worktree-new feature/payment-integration

# ../repo-feature-payment-integration/ で作業

# 作業完了後、worktreeを削除
git worktree-delete
```

### 既存のworktree間を切り替える

```bash
# 現在のworktree以外を一覧表示して選択
git worktree-switch

# 表示例:
# Worktree 一覧:
#
# 1. /path/to/repo-feature-user-auth
#    ブランチ: feature/user-auth
#    コミット: abc1234
#
# 2. /path/to/repo-feature-payment
#    ブランチ: feature/payment
#    コミット: def5678
#
# 選択してください (番号を入力、Enterでキャンセル): 1
```
