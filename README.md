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
- `git time`：コミット履歴から作業時間を自動集計し、ブランチごとやコミットごとに可視化します。

どれも `git-xxx` という名前のバイナリを用意することで、`git xxx` として呼び出せる Git 拡張サブコマンドです。

## インストール

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
go install github.com/tonbiattack/git-plus/cmd/git-time@latest
```

`@latest` で解決できない場合（モジュールプロキシの都合など）には、`@main` を指定するとリポジトリの最新コミットを直接取得できます。

`GOBIN`（または `GOPATH/bin`）が PATH に含まれていない場合は、環境に合わせて追加してください。カスタムディレクトリに配置したい場合は、`go install` の代わりに `go build -o <任意のパス>` を利用できます。

### PATH を通してカスタムコマンドとして利用する

`go install` を使わずにローカルビルドしたバイナリをサブコマンドとして登録する場合は、`~/bin` など任意のディレクトリに出力し、そのパスを環境変数 `PATH` に追加します。

```bash
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
go build -o ~/bin/git-time ./cmd/git-time
export PATH=$PATH:~/bin
git newbranch feature/awesome
```

1. `go build` でバイナリを作成し、`~/bin` に保存します。
2. `export PATH=...` で `~/bin` を検索パスに追加します。
3. 以降は `git newbranch` のように Git サブコマンドとして呼び出せます。

この設定を永続化したい場合は、`export PATH=$PATH:~/bin` の行を `~/.bashrc` や `~/.zshrc` などのシェル設定ファイルに追記してください。Fish や Windows PowerShell を利用している場合は、それぞれの方法でパスを追加してください。

同じ `go install` を再度実行すると、指定したバージョンのモジュールが改めて取得され、既存のバイナリが上書きされます。

### ローカルで動作を確認する方法

リポジトリをクローンしている場合は、`go install` を使わなくてもローカルでそのままビルド・実行できます。

```bash
git clone git@github.com:tonbiattack/git-plus.git
cd git-plus
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
go build -o ./bin/git-time ./cmd/git-time
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
go run ./cmd/git-time
```

Windows で PowerShell を利用している場合は、`./bin/git-newbranch` の代わりに `.\bin\git-newbranch.exe` のようにパスを指定してください。

同じ `go install` を再度実行すると、指定したバージョンのモジュールが改めて取得され、既存のバイナリが上書きされます。


## アンインストール

`go install` で配置したバイナリは、既定では `go env GOBIN`（未設定時は `$(go env GOPATH)/bin`）に保存されます。不要になった場合は、配置先からバイナリを削除してください。

### バイナリのみ削除（通常のアンインストール）

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
rm $(go env GOPATH)/bin/git-time
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
Remove-Item "$env:GOPATH\bin\git-time.exe"
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
```

1. `git commit --amend` を呼び出し、直前のコミットを再編集します。
2. サブコマンドに渡した追加の引数は、そのまま `git commit --amend` に引き渡されます（例: `--no-edit` や `--reset-author`）。
3. Git コマンドの終了コードを引き継ぐため、エディタを閉じるまで待機し、失敗時は同じ終了ステータスで終了します。

### git squash

```bash
git squash           # 対話的にコミット数を選択
git squash 3         # 直近3つのコミットをスカッシュ
```

1. 引数なしで実行すると、最近の10個のコミットを表示し、スカッシュするコミット数を入力で指定できます。
2. 引数でコミット数を指定すると、その数のコミットを確認表示してからスカッシュします。
3. 確認後、`git reset --soft HEAD~N` でコミットを取り消し、元のコミットメッセージを参考表示します。
4. 新しいコミットメッセージをユーザーが入力し、自動的に新しいコミットを作成します。

### git delete-local-branches

```bash
git delete-local-branches
```

1. `git branch --merged` に含まれ、`main` / `master` / `develop` 以外のブランチを抽出します。
2. 削除候補を一覧表示し、確認プロンプトで `y` / `yes` が入力されたときのみ削除します。
3. 各ブランチを `git branch -d` で削除します。未統合で削除できなかった場合はエラーを表示し、処理結果を通知します。

### git track

```bash
git track                    # origin/<現在のブランチ名> をトラッキング（リモートブランチがなければ自動プッシュ）
git track upstream           # upstream/<現在のブランチ名> をトラッキング
git track origin feature-123 # origin/feature-123 をトラッキング
```

1. 引数なしで実行すると、現在のブランチに対して `origin/<現在のブランチ名>` をトラッキングブランチとして設定します。
2. リモート名を指定すると、そのリモートの同名ブランチをトラッキングします（例: `upstream`）。
3. リモート名とブランチ名の両方を指定すると、そのリモートブランチをトラッキングします。
4. **指定したリモートブランチが存在しない場合は、自動的に `git push --set-upstream` を実行してリモートブランチを作成し、トラッキング設定を行います。**

`git pull` 実行時に「There is no tracking information for the current branch」というエラーが出た場合や、新しいブランチを作成後すぐに `git push` したい場合に便利です。リモートブランチがまだ存在しない場合でも、`git track` 一つでプッシュとトラッキング設定が完了します。

### git undo-last-commit

```bash
git undo-last-commit
```

1. `git reset --soft HEAD^` を実行し、直近のコミットだけを取り消します。
2. 作業ツリーとステージング内容はそのまま残るため、コミットメッセージを修正したいときや再コミットしたいときに便利です。

### git tag-diff

```bash
git tag-diff V4.2.00.00 V4.3.00.00
```

1. 2つのタグ間のコミット差分を取得します。
2. Mergeコミットは自動的に除外されます（`--no-merges`オプション使用）。
3. 出力形式は `- コミットメッセージ (作成者名, 日付)` です。
4. 出力ファイル名は自動的に `tag_diff_<旧タグ>_to_<新タグ>.txt` として生成されます。

指定したタグが存在しない場合や、タグ間に差分がない場合は適切なメッセージを表示します。

### git stash-cleanup

```bash
git stash-cleanup
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
```

1. 最近コミットがあったブランチを最大10件、時系列順（最新順）に表示します。
2. 現在のブランチは一覧から除外されます。
3. 番号を入力することで、選択したブランチに即座に切り替えられます。
4. 空入力でキャンセルできます。

引数は不要です。頻繁に複数のブランチを行き来する場合や、最近作業していたブランチ名を思い出せない場合に便利です。

### git time

```bash
git time           # デフォルト: 過去1週間の作業時間をブランチ別に表示
git time -w 1      # 過去1週間の作業時間
git time -m 1      # 過去1ヶ月の作業時間
git time -y 1      # 過去1年の作業時間
git time -w 2 -c   # 過去2週間をコミット別に表示
```

1. 指定された期間のコミット履歴を分析し、作業時間を自動集計します。
2. デフォルトではブランチごとに集計して表示します（`--commits`オプションでコミット別表示に切り替え）。
3. 連続するコミット間が2時間以内の場合、その時間を作業時間として計算します。
4. 2時間を超える場合や最後のコミットは、デフォルトで30分と見積もります。
5. **結果は自動的にファイル（`git_time_*.txt`）に保存されます。**

**オプション:**
- `-w, --weeks <数>`: 過去N週間の作業時間を集計
- `-m, --months <数>`: 過去Nヶ月の作業時間を集計
- `-y, --years <数>`: 過去N年の作業時間を集計
- `--since, -s <日時>`: 集計開始日時
- `--until, -u <日時>`: 集計終了日時（デフォルト: 現在）
- `--commits, -c`: コミット別に表示（デフォルトはブランチ別）
- `--help, -h`: ヘルプを表示

プロジェクトごとの工数把握や、どのブランチにどれくらい時間を費やしたかを可視化したい場合に便利です。出力ファイルは工数レポートとして活用できます。

## 開発メモ

- Go 1.22 以降でのビルドを想定しています。
- ルートに `go.mod` を置き、各コマンドは `cmd/<name>/main.go` に配置しています。
- 追加のコマンドを作成する場合は `cmd` 配下にディレクトリを増やし、`go build ./cmd/<name>` でビルドしてください。
 