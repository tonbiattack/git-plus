# ブランチ操作コマンド

ブランチの作成、切り替え、削除、同期などのブランチ管理に関するコマンドです。

## git newbranch

指定したブランチ名を一度削除してから作り直し、トラッキングブランチとしてチェックアウトします。

```bash
git newbranch feature/awesome
git newbranch -h                 # ヘルプを表示
```

**動作:**
1. 同名のローカルブランチが存在しない場合は、新しいブランチを作成して切り替えます。
2. 同名のローカルブランチが存在する場合は、以下の選択肢が表示されます：
   - `[r]ecreate`: ブランチを削除して作り直し
   - `[s]witch`: 既存のブランチに切り替え
   - `[c]ancel`: 処理を中止
3. recreate を選択すると、既存ブランチを強制削除してから新しいブランチを作成します。
4. switch を選択すると、既存ブランチに `git checkout` で切り替えます。

存在しないブランチを削除しようとした場合のエラーは無視されるため、安全に再作成できます。

## git rename-branch

現在チェックアウトしているブランチ名を変更します。`--push` を指定すると、リネーム後のブランチを `origin`（必要に応じて `--remote` で任意のリモートを指定）へプッシュして upstream を再設定できます。`--delete-remote` を指定すると古いリモートブランチの削除まで自動化します。

```bash
git rename-branch feature/renamed
git rename-branch release/v2 --push
git rename-branch hotfix/login --push --delete-remote
git rename-branch feat/ui --remote upstream --push
```

**処理:**
1. 現在のブランチ名を取得し、新しいブランチ名と重複していないかを確認
2. `git branch -m <old> <new>` でローカルブランチをリネーム
3. `--push` 指定時は `git push --set-upstream <remote> <new>` でリモートにプッシュし upstream を更新
4. `--delete-remote` 指定時は確認プロンプトの後に `git push <remote> --delete <old>` で古いリモートブランチを削除
5. `--push` を指定しない場合でも、手動でリモートを更新するためのコマンド例を表示

**オプション:**
- `--push`: 新しいブランチをリモートにプッシュして upstream を更新
- `--delete-remote`: リモート上の旧ブランチを削除（--push が必須、確認プロンプトは (y/N)）
- `--remote`: 更新対象のリモート名を指定（デフォルト: `origin`）

手作業での `git branch -m` やリモート削除の入力ミスを避け、安全にリネームを進めたい場合に便利です。

## git delete-local-branches

`main` / `master` / `develop` 以外のマージ済みローカルブランチをまとめて削除します。

```bash
git delete-local-branches
git delete-local-branches -h     # ヘルプを表示
```

**動作:**
1. `git branch --merged` に含まれ、`main` / `master` / `develop` 以外のブランチを抽出します。
2. 削除候補を一覧表示し、確認プロンプトで `y` / `yes` が入力されたときのみ削除します。
3. 各ブランチを `git branch -d` で削除します。未統合で削除できなかった場合はエラーを表示し、処理結果を通知します。

## git recent

最近使用したブランチを時系列で表示し、番号で選択して簡単に切り替えられます。

```bash
git recent
git recent -h                    # ヘルプを表示
```

**動作:**
1. 最近コミットがあったブランチを最大10件、時系列順（最新順）に表示します。
2. 現在のブランチは一覧から除外されます。
3. 番号を入力することで、選択したブランチに即座に切り替えられます。
4. 空入力でキャンセルできます。

引数は不要です。頻繁に複数のブランチを行き来する場合や、最近作業していたブランチ名を思い出せない場合に便利です。

## git sync

現在のブランチをリモートのデフォルトブランチ（main/master）の最新状態と同期します。

```bash
git sync                    # 現在のブランチをリモートのデフォルトブランチと同期
git sync feature-branch     # 指定したブランチをリモートのデフォルトブランチと同期
git sync --continue         # コンフリクト解消後に同期を続行
git sync -c                 # コンフリクト解消後に同期を続行（短縮形）
git sync --abort            # 同期を中止して元の状態に戻す
git sync -a                 # 同期を中止して元の状態に戻す（短縮形）
git sync -h                 # ヘルプを表示
```

内部的にはrebaseを使用するため、きれいな履歴を保ちながら最新の変更を取り込めます。

**主な機能:**
- **自動ブランチ検出**: リモートのデフォルトブランチ（origin/main または origin/master）を自動検出します。
- **rebaseベースの同期**: マージコミットを作らずに、きれいな履歴を維持します。
- **コンフリクト処理**: コンフリクトが発生した場合は、解消後に`git sync --continue`（または`git sync -c`）で続行できます。
- **安全な中止**: `git sync --abort`（または`git sync -a`）で同期をキャンセルし、元の状態に戻せます。
- **進行中のrebase検出**: すでにrebase中の場合は適切なメッセージを表示します。

**オプション:**
- `-c, --continue`: コンフリクト解決後にrebaseを続行
- `-a, --abort`: 同期を中止して元の状態に戻す
- `-h, --help`: ヘルプを表示

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

## git abort

進行中の Git 操作（rebase / merge / cherry-pick / revert）を安全に中止します。引数を省略すると現在の状態を自動判定します。

```bash
git abort                 # 状態を判定して進行中の操作を中止
git abort merge           # マージを強制的に中止
git abort rebase          # リベースを強制的に中止
git abort cherry-pick     # チェリーピックを強制的に中止
git abort revert          # リバートを強制的に中止
git abort -h              # ヘルプを表示
```

**動作:**
1. 引数がない場合は `.git/rebase-merge` や `MERGE_HEAD` などのインジケーターファイルを確認して進行中の操作を判定します。
2. `git <操作> --abort` を実行して処理を中止します。
3. 中止結果を表示します。

**自動判定の対象:**
- `rebase`: `.git/rebase-apply` または `.git/rebase-merge` が存在する場合
- `merge`: `.git/MERGE_HEAD` が存在する場合
- `cherry-pick`: `.git/CHERRY_PICK_HEAD` が存在する場合
- `revert`: `.git/REVERT_HEAD` が存在する場合

**注意事項:**
- 操作が検出できない場合はエラーになります。その際はコマンド引数で操作を指定してください。
- すでに操作が完了している場合は `--abort` が失敗することがあります。Git が表示するメッセージに従ってください。
