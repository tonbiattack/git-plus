# リリース管理コマンド

GitHubリリースノートの自動生成に関するコマンドです。

## git release-notes

既存のタグからGitHubリリースノートを自動生成します。GitHub CLIを使用して、タグ間の変更内容を自動的に解析してリリースを作成します。

**使い方:**
```bash
git release-notes [オプション]
git release-notes -h             # ヘルプを表示
```

**処理フロー:**
1. 既存のタグを一覧表示（または指定されたタグを使用）
2. タグを選択
3. GitHub CLIを使用してリリースノートを自動生成
4. GitHubリリースとして公開

**オプション:**
- `-t, --tag <タグ名>`: リリースを作成するタグを指定
- `-l, --latest`: 最新タグからリリースを作成
- `-d, --draft`: ドラフトとして作成
- `-p, --prerelease`: プレリリースとして作成
- `-h, --help`: ヘルプを表示

**使用例:**
```bash
# 対話的にタグを選択
git release-notes

# 指定したタグからリリース作成
git release-notes --tag v1.2.3
git release-notes -t v1.2.3          # 短縮形

# 最新タグからリリース作成
git release-notes --latest
git release-notes -l                 # 短縮形

# ドラフトとして作成
git release-notes --draft
git release-notes -d                 # 短縮形

# プレリリースとして作成
git release-notes --prerelease
git release-notes -p                 # 短縮形

# 複数のオプションを組み合わせ
git release-notes --tag v1.2.3 --draft --prerelease
git release-notes -t v1.2.3 -d -p    # 短縮形
```

**実行の流れ（対話的モード）:**
```
最近のタグ一覧 (10 個):

1. v1.3.0
2. v1.2.5
3. v1.2.4
4. v1.2.3
5. v1.2.2
...

リリースノートを作成するタグを選択してください (番号を入力、Enterでキャンセル): 1

選択されたタグ: v1.3.0

タグ: v1.3.0

リリースノートを作成しますか？ (Y/n): y

✓ リリースノートを作成しました
詳細を確認するには: gh release view v1.3.0 --web
```

**主な機能:**
- **自動リリースノート生成**: GitHub CLIの`--generate-notes`オプションを使用して、タグ間の変更内容から自動的にリリースノートを生成します。
- **対話的なタグ選択**: タグを指定しない場合、最近のタグから選択できます。
- **ドラフトモード**: `--draft`オプションでドラフトとして作成し、公開前にレビューできます。
- **プレリリースモード**: `--prerelease`オプションでプレリリースとしてマークできます。
- **最新タグ自動選択**: `--latest`オプションで最新タグを自動的に使用できます。

**注意事項:**
- GitHub CLI (`gh`) がインストールされている必要があります
- `gh auth login`でログイン済みである必要があります
- このコマンドは既存のタグに対してリリースを作成します
- 新しいタグを作成する場合は、事前に`git new-tag`コマンドを使用してください
- リリースノートは、前のタグとの差分から自動的に生成されます
- 生成されたリリースノートには、PRのタイトルとマージされた変更が含まれます

**GitHub CLI のインストール:**

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

**git new-tag との連携:**

新しいバージョンをリリースする場合の推奨フロー:
```bash
# 1. 新しいタグを作成してプッシュ
git new-tag feature --push

# 2. リリースノートを作成
git release-notes --latest

# または一度に実行（最新タグを自動使用）
git new-tag feature --push && git release-notes --latest
```
