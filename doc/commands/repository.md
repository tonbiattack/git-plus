# リポジトリ管理コマンド

リポジトリの作成、クローン、ブラウザで開く、他人のリポジトリ一覧などのリポジトリ管理に関するコマンドです。

## git create-repository

GitHubリポジトリの作成からクローン、VSCode起動までを自動化します。public/private選択、説明の指定が可能です。

**使い方:**
```bash
git create-repository <リポジトリ名>
git create-repository -h         # ヘルプを表示
```

**処理フロー:**
1. GitHubにリモートリポジトリを作成（public/private選択可能、Description指定可能）
2. 作成したリポジトリをクローン
3. クローンしたディレクトリに移動
4. VSCodeでプロジェクトを開く

**使用例:**
```bash
git create-repository my-new-project
```

**実行の流れ:**
1. コマンドを実行してリポジトリ名を指定
2. 公開設定（public/private）を選択（デフォルト: private）
3. 説明を入力（省略可）
4. 確認メッセージで `y` を入力
5. 自動的にリポジトリ作成→クローン→移動→VSCode起動を実行

**使用する主なコマンド:**
- `gh repo create`: GitHubリポジトリの作成
- `git clone`: リポジトリのクローン
- `code .`: VSCodeの起動

**注意事項:**
- GitHub CLI (`gh`) がインストールされている必要があります
- `gh auth login` でログイン済みである必要があります
- VSCode (`code` コマンド) がパスに含まれている必要があります

## git clone-org

GitHub組織のリポジトリを一括クローンします。最終更新日時でソートし、最新N個のみをクローン可能。既存リポジトリはスキップし、アーカイブやshallowクローンのオプションも利用可能です。

**使い方:**
```bash
git clone-org <organization> [オプション]
git clone-org -h                 # ヘルプを表示
```

**引数:**
- `organization`: GitHub組織名

**オプション:**
- `-n, --limit <数>`: 最新N個のリポジトリのみをクローン（デフォルト: すべて）
- `-a, --archived`: アーカイブされたリポジトリも含める（デフォルト: 除外）
- `-s, --shallow`: shallow クローンを使用（`--depth=1`）
- `-h, --help`: ヘルプを表示

**処理フロー:**
1. GitHub CLI を使用してリポジトリ一覧を取得
2. 最終更新日時でソート（最新順）
3. 組織名のディレクトリを作成
4. 各リポジトリを順次クローン
   - 既存のリポジトリはスキップ
   - アーカイブされたリポジトリは `--archived` オプションがない限りスキップ
   - `--limit N` が指定されている場合は上位N個のみをクローン
5. 結果を表示

**使用例:**
```bash
# myorg 組織の全リポジトリをクローン
git clone-org myorg

# 最新5個のリポジトリのみをクローン
git clone-org myorg --limit 5

# 最新10個のリポジトリのみをクローン（省略形）
git clone-org myorg -n 10

# アーカイブも含める
git clone-org myorg --archived
git clone-org myorg -a              # 短縮形

# shallow クローンを使用
git clone-org myorg --shallow
git clone-org myorg -s              # 短縮形

# 最新3個をshallowクローン
git clone-org myorg --limit 3 --shallow
git clone-org myorg -n 3 -s         # 短縮形
```

**実行の流れ (--limit 指定時):**
```
組織名: myorg
オプション: 最新 5 個のリポジトリのみをクローン
オプション: shallow クローン (--depth=1)

[1/3] リポジトリ一覧を取得しています...
✓ 15個のリポジトリを取得しました

注意: 3個のアーカイブされたリポジトリをスキップします。
アーカイブされたリポジトリも含める場合は --archived オプションを使用してください。

最新 5 個のリポジトリに制限します。

5個のリポジトリをクローンしますか？
続行しますか？ (Y/n): y

[2/3] クローン先ディレクトリを作成しています...
✓ ディレクトリを作成しました: ./myorg

[3/3] リポジトリをクローンしています...

[1/5] repo-latest
  📥 クローン中...
  ✅ 完了

[2/5] repo-recent
  ⏩ スキップ: すでに存在します

...

✓ すべての処理が完了しました！
📊 結果: 4個クローン, 1個スキップ
```

**主な機能:**
- **最終更新日時でソート**: リポジトリを最終更新日時（pushedAt）でソートし、最新順にクローン
- **件数制限**: `--limit N` で最新N個のリポジトリのみをクローン可能
- **スマートな警告**: 50個以上のリポジトリをクローンする場合、自動的に警告を表示し、`--limit` オプションの使用を提案
- **重複チェック**: すでに同じフォルダに同じ名前のリポジトリがある場合は自動的にスキップ
- **アーカイブフィルタリング**: デフォルトでアーカイブされたリポジトリを除外（`--archived` で含める）
- **Shallow クローン**: `--shallow` オプションで高速なクローンが可能
- **進捗表示**: クローン中のリポジトリと結果をリアルタイムで表示
- **エラーハンドリング**: クローンに失敗した場合でも続行し、最後に結果を表示

**注意事項:**
- GitHub CLI (`gh`) がインストールされている必要があります
- `gh auth login` でログイン済みである必要があります
- HTTPS URLを使用するため、SSH認証の設定は不要です
- リポジトリ数が多い場合は時間がかかることがあります
- リポジトリは `./組織名/` ディレクトリ配下にクローンされます

## git batch-clone

ファイルに記載されたリポジトリURLを一括でクローンします。テキストファイルにリポジトリURLをリストアップしておくことで、複数のリポジトリを効率的にクローンできます。

**使い方:**
```bash
git batch-clone <file> [オプション]
git batch-clone -h                   # ヘルプを表示
```

**引数:**
- `file`: リポジトリURLを記載したテキストファイルのパス

**オプション:**
- `-d, --dir <ディレクトリ>`: クローン先ディレクトリ名（省略時はファイル名が使用されます）
- `-s, --shallow`: shallow クローンを使用（`--depth=1`）
- `-h, --help`: ヘルプを表示

**ファイルフォーマット:**
```
# マイプロジェクト
https://github.com/user/repo1
https://github.com/user/repo2

# アーカイブ
https://github.com/user/old-repo
git@github.com:user/private-repo.git
```

- 1行に1つのリポジトリURL（HTTPSまたはSSH形式）
- 空行は無視されます
- `#` で始まる行はコメントとして無視されます

**処理フロー:**
1. ファイルからリポジトリURLを読み込み
2. クローン先ディレクトリを作成
3. 各リポジトリを順次クローン
   - 既存のリポジトリはスキップ
   - エラーが発生した場合は続行
4. 結果を表示

**使用例:**
```bash
# repos.txt のリポジトリを "repos" フォルダにクローン
git batch-clone repos.txt

# "myprojects" フォルダにクローン
git batch-clone repos.txt --dir myprojects
git batch-clone repos.txt -d myprojects       # 短縮形

# shallow クローンを使用
git batch-clone repos.txt --shallow
git batch-clone repos.txt -s                  # 短縮形

# カスタムフォルダ + shallow クローン
git batch-clone repos.txt -d proj -s
```

**実行の流れ:**
```
入力ファイル: repos.txt
クローン先ディレクトリ: repos
オプション: shallow クローン (--depth=1)

[1/3] リポジトリURLを読み込んでいます...
✓ 3個のリポジトリURLを読み込みました

クローン対象リポジトリ:
  1. user/repo1 (https://github.com/user/repo1)
  2. user/repo2 (https://github.com/user/repo2)
  3. user/old-repo (https://github.com/user/old-repo)

3個のリポジトリをクローンしますか？
続行しますか？ (Y/n): y

[2/3] クローン先ディレクトリを作成しています...
✓ ディレクトリを作成しました: repos

[3/3] リポジトリをクローンしています...

[1/3] user/repo1
  📥 クローン中...
  ✅ 完了

[2/3] user/repo2
  ⏩ スキップ: すでに存在します

[3/3] user/old-repo
  📥 クローン中...
  ✅ 完了

✓ すべての処理が完了しました！
📊 結果: 2個クローン, 1個スキップ, 0個失敗
```

**主な機能:**
- **柔軟なURLフォーマット**: HTTPSとSSH形式の両方に対応
- **コメント対応**: `#` で始まる行をコメントとして使用可能
- **クローン先カスタマイズ**: デフォルトはファイル名、`--dir` で任意のディレクトリ名を指定可能
- **重複チェック**: すでに存在するリポジトリは自動的にスキップ
- **Shallow クローン**: `--shallow` オプションで高速なクローンが可能
- **進捗表示**: クローン中のリポジトリと結果をリアルタイムで表示
- **エラーハンドリング**: クローンに失敗した場合でも続行し、最後に結果を表示

**注意事項:**
- HTTPSとSSH形式の両方のURLに対応していますが、SSH形式を使用する場合は適切なSSH認証の設定が必要です
- 無効なURL形式の行は警告を表示してスキップされます
- リポジトリは `<クローン先ディレクトリ>/<リポジトリ名>/` にクローンされます

## git browse

現在のリポジトリをブラウザで開きます。リポジトリの概要を素早く確認したい場合に便利です。

```bash
git browse
git browse -h                    # ヘルプを表示
```

## git repo-others

ローカルにクローン済みの他人のGitHubリポジトリを一覧表示します。フォークも含め、README プレビューを表示し、番号選択でブラウザで開くことができます。

```bash
git repo-others                      # カレントディレクトリから検索
git repo-others --path ~/projects    # 指定したディレクトリから検索
git repo-others -p ~/projects        # 短縮形
git repo-others --all                # 自分のリポジトリも含めてすべて表示
git repo-others -a                   # 短縮形
git repo-others -h                   # ヘルプを表示
```

**オプション:**
- `-p, --path <ディレクトリ>`: 検索するディレクトリ（デフォルト: カレントディレクトリ）
- `-a, --all`: 自分のリポジトリも含めてすべて表示（デフォルト: 他人のリポジトリのみ）
- `-h, --help`: ヘルプを表示

**主な機能:**
- ローカルにクローン済みの他人のリポジトリを検索
- READMEの最初の数行をプレビュー表示
- 番号を入力してブラウザで開く
- フォークしたリポジトリも含めて表示
- 検索ディレクトリのカスタマイズ可能
- 自分のリポジトリを含めるオプション

**注意事項:**
- GitHub CLI (`gh`) がインストールされている必要があります
- `gh auth login` でログイン済みである必要があります
