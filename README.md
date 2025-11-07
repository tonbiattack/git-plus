Git の日常操作を少しだけ楽にするためのカスタムコマンド集です。元々 Bash で書いていたスクリプトを Go で書き直し、単体のバイナリとして配布できるようにしました。

## コマンド一覧

- `git newbranch`：指定したブランチ名を一度削除してから作り直し、トラッキングブランチとしてチェックアウトします。
- `git reset-tag`：指定したタグをローカルとリモートから削除し、最新コミットに再作成して再プッシュします。
- `git amend`：直前のコミットを `git commit --amend` で再編集します。追加のオプションはそのまま渡せます。
- `git squash`：直近の複数コミットを対話的にスカッシュします。引数なしで実行すると最近のコミットを表示して選択できます。
- `git track`：現在のブランチにトラッキングブランチを設定します。`git pull` でエラーが出る場合に便利です。
- `git delete-local-branches`：`main` / `master` / `develop` 以外のマージ済みローカルブランチをまとめて削除します。
- `git undo-last-commit`：直近のコミットを取り消し、変更内容をステージング状態のまま残します。
- `git tag-diff`：2つのタグ間の差分を取得し、課題IDを抽出してファイルに出力します。リリースノート作成に便利です。

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
```

Windows で PowerShell を利用している場合は、`./bin/git-newbranch` の代わりに `.\bin\git-newbranch.exe` のようにパスを指定してください。

同じ `go install` を再度実行すると、指定したバージョンのモジュールが改めて取得され、既存のバイナリが上書きされます。


## アンインストール

`go install` で配置したバイナリは、既定では `go env GOBIN`（未設定時は `$(go env GOPATH)/bin`）に保存されます。不要になった場合は、配置先からバイナリを削除してください。

```bash
rm $(go env GOBIN)/git-newbranch
rm $(go env GOBIN)/git-reset-tag
rm $(go env GOBIN)/git-track
rm $(go env GOBIN)/git-delete-local-branches
rm $(go env GOBIN)/git-undo-last-commit
rm $(go env GOBIN)/git-tag-diff
```

PowerShell を利用している場合は、以下のように拡張子付きで削除できます。

```powershell
Remove-Item (go env GOBIN)\git-newbranch.exe
Remove-Item (go env GOBIN)\git-reset-tag.exe
Remove-Item (go env GOBIN)\git-track.exe
Remove-Item (go env GOBIN)\git-delete-local-branches.exe
Remove-Item (go env GOBIN)\git-undo-last-commit.exe
Remove-Item (go env GOBIN)\git-tag-diff.exe
```

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
git track                    # origin/<現在のブランチ名> をトラッキング
git track upstream           # upstream/<現在のブランチ名> をトラッキング
git track origin feature-123 # origin/feature-123 をトラッキング
```

1. 引数なしで実行すると、現在のブランチに対して `origin/<現在のブランチ名>` をトラッキングブランチとして設定します。
2. リモート名を指定すると、そのリモートの同名ブランチをトラッキングします（例: `upstream`）。
3. リモート名とブランチ名の両方を指定すると、そのリモートブランチをトラッキングします。
4. 指定したリモートブランチが存在しない場合はエラーを表示します。

`git pull` 実行時に「There is no tracking information for the current branch」というエラーが出た場合に便利です。

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

## 開発メモ

- Go 1.22 以降でのビルドを想定しています。
- ルートに `go.mod` を置き、各コマンドは `cmd/<name>/main.go` に配置しています。
- 追加のコマンドを作成する場合は `cmd` 配下にディレクトリを増やし、`go build ./cmd/<name>` でビルドしてください。
 