## カスタムGitコマンド一覧

### git reset tag {tag_name}

```
git tag -d {tag_name}
git push -d origin {tag_name}
git tag {tag_name}
git push origin {tag_name}
```

指定したタグを削除し、再作成してリモートにプッシュするコマンドです。

---

### git create branch {branch_name}

```
git branch -d {branch_name}
git checkout -b {branch_name}
```

指定したブランチを削除し、新規作成してチェックアウトするコマンドです。

---

### git delete local branches

マージ済みローカルブランチを一括削除するコマンドです。確認後、削除可能です。

---

### git undo last commit

最後のコミットを取り消し、変更は残したままにするコマンドです。

---

### git diff commit files {commit_hash}

指定したコミットで変更されたファイルごとに差分を出力するコマンドです。
出力は `diff_output_{commit_hash}` ディレクトリに格納され、
変更ファイルリストも `changed_files.txt` として保存されます。

---

### git ai review

現在の未コミットの変更点を `ai_review_diff.txt` に出力し、AIレビューに使いやすくするコマンドです。

---

### git amend

直前のコミット内容やメッセージを修正するコマンドです。

---

### git commit branch [メッセージ]

現在のブランチ名をメッセージ冒頭に付けてコミットするコマンドです。
- 引数なし → エディタでメッセージ入力
- 引数あり → 1行メッセージとして即時コミット

ステージングにファイルがない場合はエラーになります。

---

### git squash {コミット数}

指定した数の直近コミットをまとめる（スカッシュ）コマンドです。
インタラクティブリベースでコミットを整理するのに便利です。

---

## 権限付与

```
chmod +x {ファイル名}
```

スクリプトに実行権限を付与します。

## フォルダの設定

```
echo 'export PATH="$PATH:/c/apps/Command"' >> ~/.bashrc
source ~/.bashrc
```

カスタムGitコマンド用フォルダにPATHを通す手順です。

## 全てに権限付与

```
chmod +x *
```

フォルダ内のすべてのファイルに実行権限を付与します。

