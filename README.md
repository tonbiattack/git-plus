Git の日常操作を少しだけ楽にするためのカスタムコマンド集です。元々 Bash で書いていたスクリプトを Go で書き直し、単体のバイナリとして配布できるようにしました。

## コマンド一覧

- `git newbranch`：指定したブランチ名を一度削除してから作り直し、トラッキングブランチとしてチェックアウトします。
- `git reset-tag`：指定したタグをローカルとリモートから削除し、最新コミットに再作成して再プッシュします。
- `git amend`：直前のコミットを `git commit --amend` で再編集します。追加のオプションはそのまま渡せます。
- `git squash`：直近の複数コミットを対話的にスカッシュします。引数なしで実行すると最近のコミットを表示して選択できます。
- `git track`：現在のブランチにトラッキングブランチを設定します。リモートブランチがなければ自動的にプッシュします。
- `git delete-local-branches`：`main` / `master` / `develop` 以外のマージ済みローカルブランチをまとめて削除します。
- `git undo-last-commit`：直近のコミットを取り消し、変更内容をステージング状態のまま残します。
- `git tag-diff`：2つのタグ間の差分を取得し、課題IDを抽出してファイルに出力します。リリースノート作成に便利です。
- `git stash-cleanup`：重複するスタッシュを検出して自動的に削除します。各重複グループの最新のものだけを残します。
- `git recent`：最近使用したブランチを時系列で表示し、番号で選択して簡単に切り替えられます。
- `git step`：リポジトリ全体のステップ数とユーザーごとの貢献度を11の指標で集計します。追加比、削除比、更新比、コード割合など多角的な分析が可能です。
- `git sync`：現在のブランチを最新のリモートブランチ（main/master）と同期します。rebaseを使用して履歴をきれいに保ちます。
- `git pr-merge`：PRの作成からマージ、ブランチ削除、最新の変更取得までを一気に実行します。GitHub CLIを使用した自動化コマンドです。
- `git pause`：現在の作業を一時保存してブランチを切り替えます。変更をスタッシュして、別のブランチでの作業を開始できます。
- `git resume`：git pause で保存した作業を復元します。元のブランチに戻り、スタッシュから変更を復元します。
- `git create-repository`：GitHubリポジトリの作成からクローン、VSCode起動までを自動化します。public/private選択、説明の指定が可能です。
- `git new-tag`：セマンティックバージョニングに従って新しいタグを自動生成します。feature/bug指定でminor/patchを自動判定します。
- `git browse`：現在のリポジトリをブラウザで開きます。リポジトリの概要を素早く確認したい場合に便利です。
- `git pr-checkout`：最新または指定されたプルリクエストをチェックアウトします。現在の作業を自動保存し、git resumeで復元できます。
- `git clone-org`：GitHub組織のリポジトリを一括クローンします。最終更新日時でソートし、最新N個のみをクローン可能。既存リポジトリはスキップし、アーカイブやshallowクローンのオプションも利用可能です。

どれも `git-xxx` という名前のバイナリを用意することで、`git xxx` として呼び出せる Git 拡張サブコマンドです。

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
- 全20コマンドのビルド
- `~/bin`（Windows: `%USERPROFILE%\bin`）への配置
- PATH環境変数への追加

**手動でビルドする場合:**

**Linux / macOS の場合:**

```bash
# ~/bin にビルド（~/bin が存在しない場合は作成）
mkdir -p ~/bin

# 全コマンドをビルド
go build -o ~/bin/git-newbranch ./cmd/git-newbranch
go build -o ~/bin/git-reset-tag ./cmd/git-reset-tag
go build -o ~/bin/git-amend ./cmd/git-amend
go build -o ~/bin/git-squash ./cmd/git-squash
go build -o ~/bin/git-track ./cmd/git-track
go build -o ~/bin/git-delete-local-branches ./cmd/git-delete-local-branches
go build -o ~/bin/git-undo-last-commit ./cmd/git-undo-last-commit
go build -o ~/bin/git-tag-diff ./cmd/git-tag-diff
go build -o ~/bin/git-stash-cleanup ./cmd/git-stash-cleanup
go build -o ~/bin/git-recent ./cmd/git-recent
go build -o ~/bin/git-step ./cmd/git-step
go build -o ~/bin/git-sync ./cmd/git-sync
go build -o ~/bin/git-pr-merge ./cmd/git-pr-merge
go build -o ~/bin/git-pause ./cmd/git-pause
go build -o ~/bin/git-resume ./cmd/git-resume
go build -o ~/bin/git-create-repository ./cmd/git-create-repository
go build -o ~/bin/git-new-tag ./cmd/git-new-tag
go build -o ~/bin/git-browse ./cmd/git-browse
go build -o ~/bin/git-pr-checkout ./cmd/git-pr-checkout
go build -o ~/bin/git-clone-org ./cmd/git-clone-org

# PATHに追加（まだ追加していない場合）
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
# zsh を使用している場合は ~/.zshrc に追加
```

**Windows (PowerShell) の場合:**

```powershell
# ユーザーディレクトリ配下に bin フォルダを作成
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# 全コマンドをビルド
go build -o "$env:USERPROFILE\bin\git-newbranch.exe" .\cmd\git-newbranch
go build -o "$env:USERPROFILE\bin\git-reset-tag.exe" .\cmd\git-reset-tag
go build -o "$env:USERPROFILE\bin\git-amend.exe" .\cmd\git-amend
go build -o "$env:USERPROFILE\bin\git-squash.exe" .\cmd\git-squash
go build -o "$env:USERPROFILE\bin\git-track.exe" .\cmd\git-track
go build -o "$env:USERPROFILE\bin\git-delete-local-branches.exe" .\cmd\git-delete-local-branches
go build -o "$env:USERPROFILE\bin\git-undo-last-commit.exe" .\cmd\git-undo-last-commit
go build -o "$env:USERPROFILE\bin\git-tag-diff.exe" .\cmd\git-tag-diff
go build -o "$env:USERPROFILE\bin\git-stash-cleanup.exe" .\cmd\git-stash-cleanup
go build -o "$env:USERPROFILE\bin\git-recent.exe" .\cmd\git-recent
go build -o "$env:USERPROFILE\bin\git-step.exe" .\cmd\git-step
go build -o "$env:USERPROFILE\bin\git-sync.exe" .\cmd\git-sync
go build -o "$env:USERPROFILE\bin\git-pr-merge.exe" .\cmd\git-pr-merge
go build -o "$env:USERPROFILE\bin\git-pause.exe" .\cmd\git-pause
go build -o "$env:USERPROFILE\bin\git-resume.exe" .\cmd\git-resume
go build -o "$env:USERPROFILE\bin\git-create-repository.exe" .\cmd\git-create-repository
go build -o "$env:USERPROFILE\bin\git-new-tag.exe" .\cmd\git-new-tag
go build -o "$env:USERPROFILE\bin\git-browse.exe" .\cmd\git-browse
go build -o "$env:USERPROFILE\bin\git-pr-checkout.exe" .\cmd\git-pr-checkout
go build -o "$env:USERPROFILE\bin\git-clone-org.exe" .\cmd\git-clone-org

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
```

**更新方法:**

```bash
cd git-plus
git pull
# 上記のビルドコマンドを再実行
```

### 代替: go install を使用（ネットワーク環境による）

Go 1.22 以降がインストールされていれば、以下のコマンドだけで導入できます。

```bash
go install github.com/tonbiattack/git-plus/cmd/git-newbranch@latest
go install github.com/tonbiattack/git-plus/cmd/git-reset-tag@latest
go install github.com/tonbiattack/git-plus/cmd/git-amend@latest
go install github.com/tonbiattack/git-plus/cmd/git-squash@latest
go install github.com/tonbiattack/git-plus/cmd/git-track@latest
go install github.com/tonbiattack/git-plus/cmd/git-delete-local-branches@latest
go install github.com/tonbiattack/git-plus/cmd/git-undo-last-commit@latest
go install github.com/tonbiattack/git-plus/cmd/git-tag-diff@latest
go install github.com/tonbiattack/git-plus/cmd/git-stash-cleanup@latest
go install github.com/tonbiattack/git-plus/cmd/git-recent@latest
go install github.com/tonbiattack/git-plus/cmd/git-step@latest
go install github.com/tonbiattack/git-plus/cmd/git-sync@latest
go install github.com/tonbiattack/git-plus/cmd/git-pr-merge@latest
go install github.com/tonbiattack/git-plus/cmd/git-pause@latest
go install github.com/tonbiattack/git-plus/cmd/git-resume@latest
go install github.com/tonbiattack/git-plus/cmd/git-new-tag@latest
go install github.com/tonbiattack/git-plus/cmd/git-pr-checkout@latest
```

`@latest` で解決できない場合（モジュールプロキシの都合など）には、`@main` を指定するとリポジトリの最新コミットを直接取得できます。

`GOBIN`（または `GOPATH/bin`）が PATH に含まれていない場合は、環境に合わせて追加してください。

### ローカルで開発・テストする方法

リポジトリをクローンしている場合は、グローバルにインストールせずにローカルでそのまま実行できます。

```bash
# プロジェクトディレクトリ内でビルド
go build -o ./bin/git-newbranch ./cmd/git-newbranch
go build -o ./bin/git-reset-tag ./cmd/git-reset-tag
go build -o ./bin/git-amend ./cmd/git-amend
go build -o ./bin/git-squash ./cmd/git-squash
go build -o ./bin/git-track ./cmd/git-track
go build -o ./bin/git-delete-local-branches ./cmd/git-delete-local-branches
go build -o ./bin/git-undo-last-commit ./cmd/git-undo-last-commit
go build -o ./bin/git-tag-diff ./cmd/git-tag-diff
go build -o ./bin/git-stash-cleanup ./cmd/git-stash-cleanup
go build -o ./bin/git-recent ./cmd/git-recent
go build -o ./bin/git-step ./cmd/git-step
go build -o ./bin/git-sync ./cmd/git-sync
go build -o ./bin/git-pr-merge ./cmd/git-pr-merge
go build -o ./bin/git-pause ./cmd/git-pause
go build -o ./bin/git-resume ./cmd/git-resume
go build -o ./bin/git-new-tag ./cmd/git-new-tag
go build -o ./bin/git-pr-checkout ./cmd/git-pr-checkout
go build -o ./bin/git-clone-org ./cmd/git-clone-org

# 相対パスで実行
./bin/git-newbranch feature/awesome
./bin/git-reset-tag v1.2.3
```

開発中に動作を素早く試したい場合は `go run` も利用できます。

```bash
go run ./cmd/git-newbranch feature/awesome
go run ./cmd/git-reset-tag v1.2.3
go run ./cmd/git-amend --no-edit
go run ./cmd/git-squash 3
go run ./cmd/git-track
go run ./cmd/git-delete-local-branches
go run ./cmd/git-undo-last-commit
go run ./cmd/git-tag-diff V4.2.00.00 V4.3.00.00
go run ./cmd/git-stash-cleanup
go run ./cmd/git-recent
go run ./cmd/git-step
go run ./cmd/git-sync
go run ./cmd/git-pr-merge
go run ./cmd/git-pause main
go run ./cmd/git-resume
go run ./cmd/git-new-tag feature
go run ./cmd/git-pr-checkout
go run ./cmd/git-pr-checkout 123
```

Windows で PowerShell を利用している場合は、`./bin/git-newbranch` の代わりに `.\bin\git-newbranch.exe` のようにパスを指定してください。


## アンインストール

### クローン版のアンインストール

**Linux / macOS:**

```bash
# バイナリを削除
rm ~/bin/git-newbranch
rm ~/bin/git-reset-tag
rm ~/bin/git-amend
rm ~/bin/git-squash
rm ~/bin/git-track
rm ~/bin/git-delete-local-branches
rm ~/bin/git-undo-last-commit
rm ~/bin/git-tag-diff
rm ~/bin/git-stash-cleanup
rm ~/bin/git-recent
rm ~/bin/git-step
rm ~/bin/git-sync
rm ~/bin/git-pr-merge
rm ~/bin/git-pause
rm ~/bin/git-resume
rm ~/bin/git-new-tag
rm ~/bin/git-pr-checkout
rm ~/bin/git-clone-org

# リポジトリも削除する場合
rm -rf ~/path/to/git-plus
```

**Windows (PowerShell):**

```powershell
# バイナリを削除
Remove-Item "$env:USERPROFILE\bin\git-newbranch.exe"
Remove-Item "$env:USERPROFILE\bin\git-reset-tag.exe"
Remove-Item "$env:USERPROFILE\bin\git-amend.exe"
Remove-Item "$env:USERPROFILE\bin\git-squash.exe"
Remove-Item "$env:USERPROFILE\bin\git-track.exe"
Remove-Item "$env:USERPROFILE\bin\git-delete-local-branches.exe"
Remove-Item "$env:USERPROFILE\bin\git-undo-last-commit.exe"
Remove-Item "$env:USERPROFILE\bin\git-tag-diff.exe"
Remove-Item "$env:USERPROFILE\bin\git-stash-cleanup.exe"
Remove-Item "$env:USERPROFILE\bin\git-recent.exe"
Remove-Item "$env:USERPROFILE\bin\git-step.exe"
Remove-Item "$env:USERPROFILE\bin\git-sync.exe"
Remove-Item "$env:USERPROFILE\bin\git-pr-merge.exe"
Remove-Item "$env:USERPROFILE\bin\git-pause.exe"
Remove-Item "$env:USERPROFILE\bin\git-resume.exe"
Remove-Item "$env:USERPROFILE\bin\git-new-tag.exe"
Remove-Item "$env:USERPROFILE\bin\git-pr-checkout.exe"
Remove-Item "$env:USERPROFILE\bin\git-clone-org.exe"

# リポジトリも削除する場合
Remove-Item -Recurse -Force "C:\path\to\git-plus"
```

### go install 版のアンインストール

`go install` で配置したバイナリは、既定では `go env GOBIN`（未設定時は `$(go env GOPATH)/bin`）に保存されます。

**Linux / macOS:**

```bash
rm $(go env GOPATH)/bin/git-newbranch
rm $(go env GOPATH)/bin/git-reset-tag
rm $(go env GOPATH)/bin/git-amend
rm $(go env GOPATH)/bin/git-squash
rm $(go env GOPATH)/bin/git-track
rm $(go env GOPATH)/bin/git-delete-local-branches
rm $(go env GOPATH)/bin/git-undo-last-commit
rm $(go env GOPATH)/bin/git-tag-diff
rm $(go env GOPATH)/bin/git-stash-cleanup
rm $(go env GOPATH)/bin/git-recent
rm $(go env GOPATH)/bin/git-step
rm $(go env GOPATH)/bin/git-sync
rm $(go env GOPATH)/bin/git-pr-merge
rm $(go env GOPATH)/bin/git-pause
rm $(go env GOPATH)/bin/git-resume
rm $(go env GOPATH)/bin/git-new-tag
rm $(go env GOPATH)/bin/git-pr-checkout
rm $(go env GOPATH)/bin/git-clone-org
```

**Windows (PowerShell):**

```powershell
Remove-Item "$env:GOPATH\bin\git-newbranch.exe"
Remove-Item "$env:GOPATH\bin\git-reset-tag.exe"
Remove-Item "$env:GOPATH\bin\git-amend.exe"
Remove-Item "$env:GOPATH\bin\git-squash.exe"
Remove-Item "$env:GOPATH\bin\git-track.exe"
Remove-Item "$env:GOPATH\bin\git-delete-local-branches.exe"
Remove-Item "$env:GOPATH\bin\git-undo-last-commit.exe"
Remove-Item "$env:GOPATH\bin\git-tag-diff.exe"
Remove-Item "$env:GOPATH\bin\git-stash-cleanup.exe"
Remove-Item "$env:GOPATH\bin\git-recent.exe"
Remove-Item "$env:GOPATH\bin\git-step.exe"
Remove-Item "$env:GOPATH\bin\git-sync.exe"
Remove-Item "$env:GOPATH\bin\git-pr-merge.exe"
Remove-Item "$env:GOPATH\bin\git-pause.exe"
Remove-Item "$env:GOPATH\bin\git-resume.exe"
Remove-Item "$env:GOPATH\bin\git-new-tag.exe"
Remove-Item "$env:GOPATH\bin\git-pr-checkout.exe"
Remove-Item "$env:GOPATH\bin\git-clone-org.exe"
```

### go install で更新されない場合の対処法

`go install` を再実行しても最新版に更新されない場合は、Go のモジュールキャッシュが原因の可能性があります。以下の方法で解決できます。

#### 1. 対象パッケージのキャッシュのみ削除（推奨）

**Linux / macOS:**

```bash
rm -rf $(go env GOMODCACHE)/github.com/tonbiattack/git-plus*
```

**Windows (PowerShell):**

```powershell
Remove-Item -Recurse -Force "$env:GOMODCACHE\github.com\tonbiattack\git-plus*"
```

削除後、再度 `go install` を実行してください。

#### 2. すべてのモジュールキャッシュを削除（影響範囲が大きい）

```bash
go clean -modcache
```

このコマンドは `$GOMODCACHE`（通常は `$GOPATH/pkg/mod`）配下のすべてのキャッシュを削除します。他のパッケージにも影響するため、必要な場合のみ実行してください。

### まとめ

| 目的 | コマンド例 |
|-----|----------|
| バイナリのみ削除（通常のアンインストール） | `rm $(go env GOPATH)/bin/git-*` |
| 対象パッケージのキャッシュ削除（更新されない時） | `rm -rf $(go env GOMODCACHE)/github.com/tonbiattack/git-plus*` |
| すべてのキャッシュ削除 | `go clean -modcache` |

💡 **補足**: `go install` にはアンインストールコマンドは存在しません。バイナリを直接削除するだけでアンインストールできます。

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

## プロジェクト構成

### ディレクトリ構造

```
.
├── cmd/               # 各コマンドのエントリーポイント
│   ├── git-amend/
│   ├── git-newbranch/
│   ├── git-sync/
│   └── ...
├── internal/          # 内部共通パッケージ
│   └── gitcmd/       # Gitコマンド実行の共通ユーティリティ
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

## 開発メモ

- Go 1.22 以降でのビルドを想定しています。
- ルートに `go.mod` を置き、各コマンドは `cmd/<name>/main.go` に配置しています。
- 共通処理は `internal/` パッケージに配置しています。
- 追加のコマンドを作成する場合は `cmd` 配下にディレクトリを増やし、`go build ./cmd/<name>` でビルドしてください。
 