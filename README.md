Git の日常操作を少しだけ楽にするためのカスタムコマンド集です。元々 Bash で書いていたスクリプトを Go で書き直し、Cobra フレームワークを使用して単一のバイナリとして配布できるようにしました。

## コマンド一覧

- `git newbranch`：指定したブランチ名を一度削除してから作り直し、トラッキングブランチとしてチェックアウトします。
- `git reset-tag`：指定したタグをローカルとリモートから削除し、最新コミットに再作成して再プッシュします。
- `git amend`：直前のコミットを `git commit --amend` で再編集します。追加のオプションはそのまま渡せます。
- `git squash`：直近の複数コミットを対話的にスカッシュします。引数なしで実行すると最近のコミットを表示して選択できます。
- `git track`：現在のブランチにトラッキングブランチを設定します。リモートブランチがなければ自動的にプッシュします。
- `git delete-local-branches`：`main` / `master` / `develop` 以外のマージ済みローカルブランチをまとめて削除します。
- `git undo-last-commit`：直近のコミットを取り消し、変更内容をステージング状態のまま残します。
- `git tag-diff`：2つのタグ間の差分を取得し、課題IDを抽出してファイルに出力します。リリースノート作成に便利です。
- `git tag-checkout`：最新のタグを取得してチェックアウトします。セマンティックバージョン順で並べられたタグから選択できます。
- `git stash-cleanup`：重複するスタッシュを検出して自動的に削除します。各重複グループの最新のものだけを残します。
- `git stash-select`：スタッシュをインタラクティブに選択して操作できます。ファイル一覧を確認しながらapply/pop/drop/showなどの操作を実行できます。
- `git recent`：最近使用したブランチを時系列で表示し、番号で選択して簡単に切り替えられます。
- `git step`：リポジトリ全体のステップ数とユーザーごとの貢献度を11の指標で集計します。追加比、削除比、更新比、コード割合など多角的な分析が可能です。
- `git sync`：現在のブランチを最新のリモートブランチ（main/master）と同期します。rebaseを使用して履歴をきれいに保ちます。
- `git pr-merge`：PRの作成からマージ、ブランチ削除、最新の変更取得までを一気に実行します。GitHub CLIを使用した自動化コマンドです。
- `git merge-pr`：プルリクエストをマージします。`gh pr merge` のラッパーで、デフォルトでマージコミットを作成し、ブランチを削除します。
- `git pause`：現在の作業を一時保存してブランチを切り替えます。変更をスタッシュして、別のブランチでの作業を開始できます。
- `git resume`：git pause で保存した作業を復元します。元のブランチに戻り、スタッシュから変更を復元します。
- `git create-repository`：GitHubリポジトリの作成からクローン、VSCode起動までを自動化します。public/private選択、説明の指定が可能です。
- `git new-tag`：セマンティックバージョニングに従って新しいタグを自動生成します。feature/bug指定でminor/patchを自動判定します。
- `git browse`：現在のリポジトリをブラウザで開きます。リポジトリの概要を素早く確認したい場合に便利です。
- `git pr-checkout`：最新または指定されたプルリクエストをチェックアウトします。現在の作業を自動保存し、git resumeで復元できます。
- `git clone-org`：GitHub組織のリポジトリを一括クローンします。最終更新日時でソートし、最新N個のみをクローン可能。既存リポジトリはスキップし、アーカイブやshallowクローンのオプションも利用可能です。
- `git back`：前のブランチやタグに戻ります。`git checkout -` のショートカットで、ブランチやタグ間の素早い移動に便利です。
- `git issue-edit`：GitHubのopenしているissueの一覧を表示し、issue番号で選択してエディタで編集します。題名（title）と本文（body）の両方を編集できます。

Cobra フレームワークで実装された単一のバイナリから、`git-xxx` 形式の名前でシンボリックリンク（Linux/macOS）またはコピー（Windows）が作成され、`git xxx` として呼び出せる Git 拡張サブコマンドとして機能します。

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
  - `git-newbranch`, `git-reset-tag`, `git-amend` など25個のコマンド
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
ln -s git-plus git-reset-tag
ln -s git-plus git-amend
ln -s git-plus git-squash
ln -s git-plus git-track
ln -s git-plus git-delete-local-branches
ln -s git-plus git-undo-last-commit
ln -s git-plus git-tag-diff
ln -s git-plus git-tag-checkout
ln -s git-plus git-stash-cleanup
ln -s git-plus git-stash-select
ln -s git-plus git-recent
ln -s git-plus git-step
ln -s git-plus git-sync
ln -s git-plus git-pr-merge
ln -s git-plus git-merge-pr
ln -s git-plus git-pause
ln -s git-plus git-resume
ln -s git-plus git-create-repository
ln -s git-plus git-new-tag
ln -s git-plus git-browse
ln -s git-plus git-pr-checkout
ln -s git-plus git-clone-org
ln -s git-plus git-back
ln -s git-plus git-issue-edit

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
Copy-Item "$binPath\git-plus.exe" "$binPath\git-reset-tag.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-amend.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-squash.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-track.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-delete-local-branches.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-undo-last-commit.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-tag-diff.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-tag-checkout.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-stash-cleanup.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-stash-select.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-recent.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-step.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-sync.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-merge.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-merge-pr.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pause.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-resume.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-create-repository.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-new-tag.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-browse.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-pr-checkout.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-clone-org.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-back.exe"
Copy-Item "$binPath\git-plus.exe" "$binPath\git-issue-edit.exe"

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
go run . stash-cleanup
go run . stash-select
go run . recent
go run . step
go run . sync
go run . pr-merge
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

## 使い方

### git newbranch

```bash
git newbranch feature/awesome
git newbranch -h                 # ヘルプを表示
```

1. 同名のローカルブランチが存在しない場合は、新しいブランチを作成して切り替えます。
2. 同名のローカルブランチが存在する場合は、以下の選択肢が表示されます：
   - `[r]ecreate`: ブランチを削除して作り直し
   - `[s]witch`: 既存のブランチに切り替え
   - `[c]ancel`: 処理を中止
3. recreate を選択すると、既存ブランチを強制削除してから新しいブランチを作成します。
4. switch を選択すると、既存ブランチに `git checkout` で切り替えます。

存在しないブランチを削除しようとした場合のエラーは無視されるため、安全に再作成できます。

### git reset-tag

```bash
git reset-tag v1.2.3
git reset-tag -h                 # ヘルプを表示
```

1. 指定したタグをローカルから削除します。
2. リモート（デフォルトでは `origin`）からも削除します。
3. 最新コミットに同名のタグを再作成します。
4. 新しいタグをリモートへプッシュし直します。

既存のタグが見つからない場合は警告が表示されますが、処理自体は継続します。再作成やプッシュに失敗した場合は終了コード 1 で停止するため、CI などでも利用できます。

### git amend

```bash
git amend
git amend --no-edit
git amend -h                     # ヘルプを表示
```

1. `git commit --amend` を呼び出し、直前のコミットを再編集します。
2. サブコマンドに渡した追加の引数は、そのまま `git commit --amend` に引き渡されます（例: `--no-edit` や `--reset-author`）。
3. Git コマンドの終了コードを引き継ぐため、エディタを閉じるまで待機し、失敗時は同じ終了ステータスで終了します。

### git squash

```bash
git squash           # 対話的にコミット数を選択
git squash 3         # 直近3つのコミットをスカッシュ
git squash -h        # ヘルプを表示
```

1. 引数なしで実行すると、最近の10個のコミットを表示し、スカッシュするコミット数を入力で指定できます。
2. 引数でコミット数を指定すると、その数のコミットを確認表示してからスカッシュします。
3. 確認後、`git reset --soft HEAD~N` でコミットを取り消し、元のコミットメッセージを参考表示します。
4. 新しいコミットメッセージをユーザーが入力し、自動的に新しいコミットを作成します。

### git delete-local-branches

```bash
git delete-local-branches
git delete-local-branches -h     # ヘルプを表示
```

1. `git branch --merged` に含まれ、`main` / `master` / `develop` 以外のブランチを抽出します。
2. 削除候補を一覧表示し、確認プロンプトで `y` / `yes` が入力されたときのみ削除します。
3. 各ブランチを `git branch -d` で削除します。未統合で削除できなかった場合はエラーを表示し、処理結果を通知します。

### git track

```bash
git track                    # origin/<現在のブランチ名> をトラッキング（リモートブランチがなければ自動プッシュ）
git track upstream           # upstream/<現在のブランチ名> をトラッキング
git track origin feature-123 # origin/feature-123 をトラッキング
git track -h                 # ヘルプを表示
```

1. 引数なしで実行すると、現在のブランチに対して `origin/<現在のブランチ名>` をトラッキングブランチとして設定します。
2. リモート名を指定すると、そのリモートの同名ブランチをトラッキングします（例: `upstream`）。
3. リモート名とブランチ名の両方を指定すると、そのリモートブランチをトラッキングします。
4. **指定したリモートブランチが存在しない場合は、自動的に `git push --set-upstream` を実行してリモートブランチを作成し、トラッキング設定を行います。**

`git pull` 実行時に「There is no tracking information for the current branch」というエラーが出た場合や、新しいブランチを作成後すぐに `git push` したい場合に便利です。リモートブランチがまだ存在しない場合でも、`git track` 一つでプッシュとトラッキング設定が完了します。

### git undo-last-commit

```bash
git undo-last-commit
git undo-last-commit -h          # ヘルプを表示
```

1. `git reset --soft HEAD^` を実行し、直近のコミットだけを取り消します。
2. 作業ツリーとステージング内容はそのまま残るため、コミットメッセージを修正したいときや再コミットしたいときに便利です。

### git tag-diff

```bash
git tag-diff V4.2.00.00 V4.3.00.00
git tag-diff -h                    # ヘルプを表示
```

1. 2つのタグ間のコミット差分を取得します。
2. Mergeコミットは自動的に除外されます（`--no-merges`オプション使用）。
3. 出力形式は `- コミットメッセージ (作成者名, 日付)` です。
4. 出力ファイル名は自動的に `tag_diff_<旧タグ>_to_<新タグ>.txt` として生成されます。

指定したタグが存在しない場合や、タグ間に差分がない場合は適切なメッセージを表示します。

### git tag-checkout

```bash
git tag-checkout                    # 最新10個のタグから選択
git tag-checkout -n 5               # 最新5個のタグから選択
git tag-checkout -y                 # 最新タグに自動チェックアウト
git tag-checkout --limit 20         # 最新20個のタグから選択
git tag-checkout --latest           # 最新タグを表示するのみ
git tag-checkout -h                 # ヘルプを表示
```

1. セマンティックバージョン順（`--sort=-v:refname`）で最新のタグを取得します。
2. デフォルトで最新10個のタグを表示します（`-n` または `--limit` オプションで変更可能）。
3. 対話的にタグを選択してチェックアウトできます。
4. `-y` オプションを使用すると、確認なしで最新タグにチェックアウトします。
5. `--latest` オプションを使用すると、最新タグを表示するのみで終了します。

**オプション:**
- `-n, --limit <数>`: 表示するタグの数（デフォルト: 10）
- `-y, --yes`: 確認なしで最新タグにチェックアウト
- `--latest`: 最新タグのみを表示して終了
- `-h`: ヘルプを表示

**主な機能:**
- **セマンティックバージョン順ソート**: `git tag --sort=-v:refname` を使用して、セマンティックバージョンに従って新しいもの → 古いものの順に並べます。
- **対話的な選択**: タグ一覧から番号を選択してチェックアウトできます。
- **高速チェックアウト**: `-y` オプションで最新タグに即座にチェックアウトできます。
- **最新タグの確認**: `--latest` オプションで最新タグを確認するのみの用途にも使えます。

引数は不要です。リリースタグやバージョンタグが多数存在する場合に、最新のタグに素早く切り替えたいときに便利です。

### git stash-cleanup

```bash
git stash-cleanup
git stash-cleanup -h             # ヘルプを表示
```

1. 全てのスタッシュを分析し、ファイル構成と内容が完全に同一のスタッシュを検出します。
2. 重複するスタッシュをグループ化して表示します。
3. 削除確認のプロンプトを表示します（`y` / `yes` で実行）。
4. 各重複グループから最新のスタッシュ（インデックスが最小）のみを残し、古い重複を削除します。
5. 削除結果と残りのスタッシュ数を表示します。

引数は不要です。このコマンドは全スタッシュを自動的にスキャンして重複を検出します。誤って同じ変更を複数回スタッシュした場合や、スタッシュが溜まりすぎた場合の整理に便利です。

### git stash-select

```bash
git stash-select
git stash-select -h              # ヘルプを表示
```

スタッシュされている変更をインタラクティブに選択して操作できます。各スタッシュのファイル一覧を確認しながら、apply（適用）、pop（適用して削除）、drop（削除）、show（差分表示）などの操作を実行できます。

**実行の流れ:**

1. 全スタッシュの一覧を表示（番号、ブランチ、メッセージ、ファイル数）
2. 番号を入力してスタッシュを選択
3. 選択したスタッシュの詳細情報とファイル一覧を表示
4. 操作を選択：
   - `[a]pply`: スタッシュを適用（スタッシュは残す）
   - `[p]op`: スタッシュを適用して削除
   - `[d]rop`: スタッシュを削除
   - `[s]how`: 差分を表示
   - `[c]ancel`: キャンセル

**使用例:**

```bash
git stash-select

# 実行結果例:
# スタッシュ一覧 (2 個):
#
# 1. stash@{0}
#    ブランチ: feature/login
#    メッセージ: WIP on feature/login: Add login form
#    ファイル数: 3
#
# 2. stash@{1}
#    ブランチ: main
#    メッセージ: WIP on main: Update README
#    ファイル数: 1
#
# 選択してください (番号を入力、Enterでキャンセル): 1
#
# 選択されたスタッシュ: stash@{0}
# メッセージ: WIP on feature/login: Add login form
# ブランチ: feature/login
#
# 変更されたファイル:
#   - src/components/LoginForm.tsx
#   - src/api/auth.ts
#   - src/types/user.ts
#
# 操作を選択してください:
#   [a]pply  - スタッシュを適用（スタッシュは残す）
#   [p]op    - スタッシュを適用して削除
#   [d]rop   - スタッシュを削除
#   [s]how   - 差分を表示
#   [c]ancel - キャンセル
#
# 選択 (a/p/d/s/c): a
```

**主な機能:**

- **視覚的な一覧表示**: 各スタッシュのブランチ名、メッセージ、ファイル数を一目で確認できます。
- **ファイル一覧の表示**: 選択したスタッシュに含まれるファイルを確認してから操作できます。
- **安全な操作**: 各操作の意味を明示し、誤操作を防ぎます。
- **柔軟な操作**: apply（残す）、pop（削除）、drop（削除のみ）、show（表示のみ）から選択できます。

引数は不要です。`git stash list` で一覧を見て、`git stash show stash@{0}` で内容を確認して、`git stash apply stash@{0}` で適用する...という手順を1つのコマンドで完結できます。スタッシュが複数ある場合や、どのスタッシュを適用すべきか確認したい場合に便利です。

### git recent

```bash
git recent
git recent -h                    # ヘルプを表示
```

1. 最近コミットがあったブランチを最大10件、時系列順（最新順）に表示します。
2. 現在のブランチは一覧から除外されます。
3. 番号を入力することで、選択したブランチに即座に切り替えられます。
4. 空入力でキャンセルできます。

引数は不要です。頻繁に複数のブランチを行き来する場合や、最近作業していたブランチ名を思い出せない場合に便利です。

### git step

```bash
git step                    # 全期間のステップ数を表示
git step -w 1               # 過去1週間
git step -m 1               # 過去1ヶ月
git step -y 1               # 過去1年
git step --since 2024-01-01 # 指定日以降
git step --include-initial  # 初回コミットを含める
git step -h                 # ヘルプを表示
```

1. リポジトリ全体のステップ数（行数）とユーザーごとの貢献度を集計します。
2. デフォルトで初回コミットは除外されます（大量の行数が追加されることが多いため）。
3. コード割合が多い順に表示されます。
4. 各作成者の11の指標を表示します：
   - **追加**: 過去に追加した行数の累計
   - **削除**: 過去に削除した行数の累計
   - **更新**: 追加と削除の合計（変更に関わった総行数）
   - **現在**: 現在のコードベースに残っている行数（git blameベース）
   - **コミ**: コミット数
   - **平均**: 平均コミットサイズ（更新行数 / コミット数）
   - **追加比**: 全体の追加行数に占める割合
   - **削除比**: 全体の削除行数に占める割合
   - **更新比**: 全体の更新行数に占める割合
   - **コード割合**: 期間指定時は期間内のコード行数の合計に対する割合（合計100%）、期間指定なしは現在のリポジトリ総行数に占める割合（合計100%）
5. 現在のリポジトリ総行数も表示されます。
6. **結果は自動的にテキストファイル（`git_step_*.txt`）とCSVファイル（`git_step_*.csv`）に保存されます。**

**オプション:**
- `-w, --weeks <数>`: 過去N週間を集計
- `-m, --months <数>`: 過去Nヶ月を集計
- `-y, --years <数>`: 過去N年を集計
- `--since, -s <日時>`: 集計開始日時
- `--until, -u <日時>`: 集計終了日時
- `--include-initial`: 初回コミットを含める（デフォルトは除外）
- `-h`: ヘルプを表示

**主な機能:**
- **11の包括的な指標**: 追加、削除、更新、現在、コミット数、平均コミットサイズ、追加比、削除比、更新比、コード割合を一度に確認できます。
- **CSV出力対応**: すべての指標を省略なしでCSV形式で出力するため、ExcelやGoogleスプレッドシートで分析できます。
- **貢献度の多角的な可視化**:
  - **追加比**、**削除比**、**更新比**で、チーム全体の作業量の中での相対的な貢献度が分かります。
  - **平均コミットサイズ**で、大きな変更を好むか小さく頻繁にコミットするかが分かります。
  - **コミット数**で、活動頻度が分かります。
  - **現在**と**コード割合**で、現在のコードベースへの実際の貢献度が分かります。
- **期間指定**: 特定の期間に限定した集計が可能です。期間指定時は、その期間内のコード行数を100%としてコード割合を計算します。
- **初回コミット除外**: デフォルトで初回コミットを除外することで、より実態に即した統計が得られます。

**注意事項:**
- バイナリファイルの行数は集計から除外されます。
- **現在**は`git blame`を使用して、現在のコードベースに実際に残っている各ユーザーの行数を正確に集計します。
- **コード割合**は、期間指定なしの場合は現在のリポジトリ総行数に対する各ユーザーの現在行数の割合、期間指定ありの場合は期間内のコード行数の合計に対する割合です（いずれも合計100%）。
- **追加比**、**削除比**、**更新比**は、それぞれ全体の追加行数・削除行数・更新行数に対する割合です。
- **平均コミットサイズ**が大きい場合、大規模な変更を一度にコミットする傾向があります。

リポジトリ全体の規模把握や、チームメンバーの貢献度を多角的に分析したい場合に便利です。

### git sync

```bash
git sync                    # 現在のブランチをリモートのデフォルトブランチと同期
git sync feature-branch     # 指定したブランチをリモートのデフォルトブランチと同期
git sync --continue         # コンフリクト解消後に同期を続行
git sync --abort            # 同期を中止して元の状態に戻す
git sync -h                 # ヘルプを表示
```

現在のブランチをリモートのデフォルトブランチ（main/master）の最新状態と同期します。内部的にはrebaseを使用するため、きれいな履歴を保ちながら最新の変更を取り込めます。

**主な機能:**
- **自動ブランチ検出**: リモートのデフォルトブランチ（origin/main または origin/master）を自動検出します。
- **rebaseベースの同期**: マージコミットを作らずに、きれいな履歴を維持します。
- **コンフリクト処理**: コンフリクトが発生した場合は、解消後に`git sync --continue`で続行できます。
- **安全な中止**: `git sync --abort`で同期をキャンセルし、元の状態に戻せます。
- **進行中のrebase検出**: すでにrebase中の場合は適切なメッセージを表示します。

**使用例:**
1. feature-branch で作業中、main の最新変更を取り込みたい場合:
   ```bash
   git switch feature-branch
   git sync
   # コンフリクトが発生した場合は解消後に:
   git sync --continue
   ```

2. 別のブランチを同期したい場合:
   ```bash
   git sync develop
   ```

**注意事項:**
- リモートへプッシュ済みのコミットをrebaseすると、履歴が書き換わるため、チームで共有しているブランチでは注意が必要です。
- コンフリクトが発生した場合は、ファイルを編集してコンフリクトを解消し、`git add`した後に`git sync --continue`を実行してください。

### git pause

```bash
git pause main              # 現在の作業を保存してmainに切り替え
git pause develop           # 現在の作業を保存してdevelopに切り替え
git pause -h                # ヘルプを表示
```

現在の作業を一時保存してブランチを切り替えます。変更をスタッシュして、別のブランチでの作業を開始できます。

**主な機能:**
- **変更の自動保存**: コミットされていない変更を自動的にスタッシュに保存します。
- **状態管理**: どのブランチからどのブランチに切り替えたかを記録します（`~/.git-plus/pause-state.json`）。
- **安全な上書き確認**: 既に pause 状態の場合は上書き確認を行います。
- **変更なしの最適化**: 変更がない場合はスタッシュせずにブランチ切り替えのみ実行します。

**使用例:**

```bash
# feature-branchで作業中、急にmainで作業が必要になった場合
git pause main

# 以下が自動実行される:
# 変更を保存中...
# ✓ 変更を保存しました: stash@{0}
# ブランチを切り替え中: feature-branch → main
# ✓ main に切り替えました
#
# 元のブランチに戻るには: git resume
```

**動作:**
1. 現在のブランチ名を記録
2. 変更があればスタッシュに保存（メッセージ: `git-pause: from <現在のブランチ>`）
3. 状態を `~/.git-plus/pause-state.json` に保存
4. 指定されたブランチに切り替え

**注意事項:**
- 既に pause 状態の場合は上書き確認が表示されます
- 変更がない場合はスタッシュせずにブランチ切り替えのみ実行されます

### git resume

```bash
git resume                  # git pause で保存した作業を復元
git resume -h               # ヘルプを表示
```

git pause で保存した作業を復元します。元のブランチに戻り、スタッシュから変更を復元します。

**主な機能:**
- **元のブランチに自動復帰**: pause 時のブランチに自動的に切り替わります。
- **スタッシュの自動復元**: 保存されていた変更を自動的に復元します。
- **状態のクリーンアップ**: 復元後、状態ファイルを自動的に削除します。
- **エラーハンドリング**: スタッシュの復元に失敗した場合でも、適切なメッセージを表示します。

**使用例:**

```bash
# mainでの作業が終わり、元のfeature-branchに戻る場合
git resume

# 以下が自動実行される:
# 元のブランチに戻ります: main → feature-branch
# ブランチを切り替え中: main → feature-branch
# ✓ feature-branch に切り替えました
# 変更を復元中...
# ✓ 変更を復元しました
#
# ✓ 作業の復元が完了しました
```

**動作:**
1. 状態ファイル（`~/.git-plus/pause-state.json`）を読み込み
2. 元のブランチに切り替え
3. スタッシュから変更を復元（`git stash pop`）
4. 状態ファイルを削除

**注意事項:**
- pause 状態がない場合はエラーメッセージを表示します
- スタッシュの復元に失敗した場合は警告を表示し、手動での復元を促します

### git new-tag

```bash
git new-tag feature          # 機能追加（minor）
git new-tag bug              # バグ修正（patch）
git new-tag major            # 破壊的変更
git new-tag f --push         # 省略形 + プッシュ
git new-tag -h               # ヘルプを表示
```

セマンティックバージョニングに従って新しいタグを自動生成します。現在の最新タグから自動的に次のバージョンを計算します。

**主な機能:**
- **自動バージョン計算**: 最新タグ（v1.2.3）から次のバージョンを自動計算
- **直感的なタイプ指定**: feature/bug で minor/patch を自動判定
- **省略形サポート**: f（feature）、b（bug）、m（major）など
- **確認プロンプト**: 誤ったタグ作成を防ぐ
- **自動プッシュ**: `--push` オプションでリモートへ自動プッシュ
- **対話的モード**: 引数なしで実行時に選択肢を表示

**バージョンタイプ:**

| タイプ | 説明 | 例 (v1.2.3 →) |
|-------|------|---------------|
| `major`, `m`, `breaking` | メジャーバージョンアップ | v2.0.0 |
| `minor`, `n`, `feature`, `f` | マイナーバージョンアップ | v1.3.0 |
| `patch`, `p`, `bug`, `b`, `fix` | パッチバージョンアップ | v1.2.4 |

**オプション:**
- `-m, --message <msg>`: タグメッセージを指定（アノテーテッドタグを作成）
- `--push`: 作成後に自動的にリモートへプッシュ
- `--dry-run`: 実際には作成せず、次のバージョンだけを表示

**使用例:**

```bash
# 機能追加のタグを作成
git new-tag feature
# 現在のタグ: v1.2.3
# 新しいタグ: v1.3.0 (MINOR)
# タグを作成しますか？ (y/N): y
# ✓ タグを作成しました: v1.3.0

# バグ修正のタグを作成してプッシュ
git new-tag bug --push
# ✓ タグを作成しました: v1.2.4
# ✓ リモートにプッシュしました: v1.2.4

# 省略形を使用
git new-tag f              # feature と同じ
git new-tag b              # bug と同じ

# メッセージ付きで作成
git new-tag feature -m "Add awesome feature"

# ドライラン（確認のみ）
git new-tag major --dry-run
# 現在のタグ: v1.2.3
# 次のバージョン: v2.0.0 (MAJOR)
# (--dry-run のため、タグは作成されません)

# 対話的モード
git new-tag
# 新しいタグのタイプを選択してください:
#   [1] major   - v2.0.0 (破壊的変更)
#   [2] minor   - v1.3.0 (機能追加)
#   [3] patch   - v1.2.4 (バグ修正)
# 選択 (1-3): 2
```

**動作:**
1. 最新のタグを取得（`git describe --tags --abbrev=0`）
2. バージョン番号を解析（v1.2.3 → MAJOR=1, MINOR=2, PATCH=3）
3. 指定されたタイプに応じて新しいバージョンを計算
4. 確認プロンプトを表示
5. タグを作成（メッセージありの場合はアノテーテッドタグ）
6. `--push` オプションがある場合はリモートへプッシュ

**注意事項:**
- タグが存在しない場合はエラーになります。最初のタグは手動で作成してください（例: `git tag v0.1.0`）
- セマンティックバージョニング（v1.2.3形式）に従ったタグが必要です
- リモートへのプッシュは `--push` オプションを指定した場合のみ実行されます

### git pr-checkout

```bash
git pr-checkout              # 最新のPRをチェックアウト
git pr-checkout 123          # PR #123 をチェックアウト
git pr-checkout -h           # ヘルプを表示
```

最新または指定されたプルリクエストをチェックアウトします。現在の作業を自動的に保存（git pause と同様）するため、後で git resume で戻ることができます。

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

### git pr-merge

```bash
git pr-merge [ベースブランチ名]
git pr-merge -h              # ヘルプを表示
```

PRの作成からマージ、ブランチ削除、最新の変更取得までを一気に実行します。GitHub CLIを使用して以下の処理を自動化します:

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
git pr-merge main

# 方法2: 対話的に入力
git pr-merge
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

### git merge-pr

```bash
git merge-pr [PR番号] [オプション]
git merge-pr -h              # ヘルプを表示
```

プルリクエストをマージします。GitHub CLI の `gh pr merge` をラップして、Git コマンドとして実行できるようにします。

**デフォルトの動作:**
- **マージコミットで直接実行**: 対話なしでマージコミットを作成（`--merge` が自動適用）
- **ブランチ自動削除**: マージ後にブランチを自動削除（`--delete-branch` が自動適用）

**主な機能:**
- **マージ方法の選択**: merge commit（デフォルト）、squash、rebase から選択可能
- **すべてのオプションをサポート**: gh pr merge のすべてのオプションがそのまま使用できます

**使用例:**

```bash
# カレントブランチのPRをマージコミットで直接マージ（ブランチも削除）
git merge-pr

# PR番号を指定してマージコミットで直接マージ（ブランチも削除）
git merge-pr 89

# スカッシュマージで直接マージ（ブランチも削除）
git merge-pr --squash

# リベースマージで直接マージ（ブランチも削除）
git merge-pr --rebase

# 複数のオプションを組み合わせ
git merge-pr 89 --squash --auto

# 自動マージ（ステータスチェック通過後に自動マージ）
git merge-pr --auto
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

**git pr-merge との違い:**

| コマンド | 用途 |
|---------|------|
| `git pr-merge` | PR作成→マージ→ブランチ切り替え→pull の一連の流れを自動化 |
| `git merge-pr` | 既存のPRをマージコミットで直接マージ（`gh pr merge` のラッパー） |

**前提条件:**
- GitHub CLI (gh) がインストールされていること
- `gh auth login`でログイン済みであること
- マージ権限があること

**注意事項:**
- 引数なしで実行すると、カレントブランチに関連するPRをマージコミットで直接マージします
- デフォルトでブランチが削除されるため、マージ後はローカル・リモート両方でブランチが削除されます

## プロジェクト構成

### ディレクトリ構造

```
.
├── cmd/               # Cobraコマンド定義
│   ├── root.go       # ルートコマンド
│   ├── amend.go      # amendサブコマンド
│   ├── newbranch.go  # newbranchサブコマンド
│   ├── sync.go       # syncサブコマンド
│   └── ...           # その他20のサブコマンド
├── internal/          # 内部共通パッケージ
│   ├── gitcmd/       # Gitコマンド実行の共通ユーティリティ
│   ├── ui/           # UI関連のユーティリティ
│   └── pausestate/   # pause/resume状態管理
├── main.go           # エントリーポイント
├── bin/              # ビルド済みバイナリ
└── go.mod
```

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

## コマンド詳細

### git create-repository

GitHubリポジトリの作成からクローン、VSCode起動までを自動化するコマンドです。

**使い方:**
```bash
git create-repository <リポジトリ名>
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

### git clone-org

GitHub組織のリポジトリを一括クローンするコマンドです。

**使い方:**
```bash
git clone-org <organization> [オプション]
```

**引数:**
- `organization`: GitHub組織名

**オプション:**
- `--limit N, -n N`: 最新N個のリポジトリのみをクローン（デフォルト: すべて）
- `--archived`: アーカイブされたリポジトリも含める（デフォルト: 除外）
- `--shallow`: shallow クローンを使用（`--depth=1`）
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

# shallow クローンを使用
git clone-org myorg --shallow

# 最新3個をshallowクローン
git clone-org myorg --limit 3 --shallow
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

**実行の流れ (全リポジトリクローン時、50個以上の場合):**
```
組織名: myorg

[1/3] リポジトリ一覧を取得しています...
✓ 75個のリポジトリを取得しました

⚠️  警告: 75個のリポジトリをクローンします。
   多数のリポジトリをクローンする場合は時間がかかります。
   最新のリポジトリのみが必要な場合は --limit オプションを検討してください。
   例: git clone-org myorg --limit 10

75個のリポジトリをクローンしますか？
続行しますか？ (Y/n): y

[2/3] クローン先ディレクトリを作成しています...
✓ ディレクトリを作成しました: ./myorg

[3/3] リポジトリをクローンしています...
...
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

認証後、HTTPS経由でリポジトリをクローンします。SSH認証の設定は不要です。

### git issue-edit

GitHubのopenしているissueの一覧を表示し、選択したissueをエディタで編集するコマンドです。

**使い方:**
```bash
git issue-edit
```

**処理フロー:**
1. GitHubのopenしているissueを一覧表示
2. issue番号を入力してissueを選択
3. 選択したissueの題名と本文を一時ファイルに書き出し
4. ユーザーが設定しているエディタで編集
5. 編集内容でissueを更新

**使用例:**
```bash
git issue-edit
```

**実行の流れ:**
```
Open Issue一覧 (3 個):

#42: ログイン機能の不具合
   ログインボタンを押しても反応がない...

#43: ダークモード対応
   ダークモードのデザインを実装してほしい...

#44: パフォーマンス改善
   初回読み込みが遅いため改善が必要...

編集するissueを選択してください (issue番号を入力、Enterでキャンセル): 43

選択されたissue: #43
タイトル: ダークモード対応
URL: https://github.com/username/repo/issues/43

エディタで編集中... (code --wait)
✓ issueを更新しました
```

**エディタの設定:**

エディタは以下の優先順位で自動検出されます：
1. `git config core.editor` の設定
2. 環境変数 `VISUAL`
3. 環境変数 `EDITOR`
4. デフォルト（vi）

VSCodeを使用する場合は、以下のように設定します：
```bash
git config --global core.editor "code --wait"
```

**一時ファイルの形式:**

エディタで開かれるファイルは以下のような形式です：
```markdown
# Issue #43
# URL: https://github.com/username/repo/issues/43
#
# 以下のissueの題名と本文を編集してください。
# '#' で始まる行はコメントとして無視されます。
# 'Title:' の後に題名を記載し、'---' の区切り線の後に本文を記載してください。
# ファイルを保存して閉じると、issueが更新されます。
# ========================================

Title: ダークモード対応

---

現在の本文がここに表示されます。
この部分を編集して保存すると、issueが更新されます。
```

**主な機能:**
- **一覧表示**: openしているissueをissue番号付きで表示
- **プレビュー**: 各issueの本文の最初の50文字をプレビュー表示
- **タイトル編集**: issueのタイトルも編集可能
- **本文編集**: issueの本文も編集可能
- **エディタ編集**: ユーザーが設定しているエディタで編集可能
- **コメント行除外**: `#` で始まる行はコメントとして無視
- **変更検出**: タイトルと本文の両方で変更を検出し、変更がない場合は更新をスキップ

**注意事項:**
- GitHub CLI (`gh`) がインストールされている必要があります
- `gh auth login` でログイン済みである必要があります
- タイトルと本文の両方を編集できます
- タイトルは `Title:` の後に、本文は `---` の区切り線の後に記載します

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

## 開発メモ

- Go 1.22 以降でのビルドを想定しています。
- Cobra フレームワークを使用した単一バイナリ構造です。
- ルートに `go.mod` を置き、各サブコマンドは `cmd/<name>.go` に配置しています。
- 共通処理は `internal/` パッケージに配置しています。
- 実行ファイル名が `git-xxx` の場合、自動的に `xxx` サブコマンドとして実行されます。
- 追加のコマンドを作成する場合は `cmd` 配下に新しい `.go` ファイルを作成し、`rootCmd` に登録してください。
 