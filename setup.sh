#!/bin/bash
# ================================================================================
# Git 拡張コマンド セットアップスクリプト (Linux/macOS)
# ================================================================================
# このスクリプトは以下の処理を実行します：
# 1. ホームディレクトリに ~/bin フォルダを作成
# 2. git-plus バイナリをビルド
# 3. 28個のgit拡張コマンド用にシンボリックリンクを作成
# 4. シェル設定ファイル（.zshrc/.bashrc/.profile）にPATHを追加
#
# これにより、git newbranch、git pr-merge などのコマンドが使用可能になります。
#
# 使用方法:
#   ./setup.sh
#
# 前提条件:
#   - Go 1.25.3以上がインストールされていること
#   - 実行権限が付与されていること (chmod +x setup.sh)
# ================================================================================

# エラー発生時に即座にスクリプトを終了
set -e

echo "=== Git 拡張コマンド セットアップ ==="
echo ""

# ================================================================================
# ステップ1: binディレクトリの作成
# ================================================================================
# ホームディレクトリ配下に bin フォルダを作成します
# 既に存在する場合でもエラーにならないよう -p オプションを使用
BIN_PATH="$HOME/bin"
echo "binディレクトリを作成中: $BIN_PATH"
mkdir -p "$BIN_PATH"

echo ""

# ================================================================================
# ステップ2: git-plus バイナリのビルド
# ================================================================================
# Goコンパイラを使用してgit-plusバイナリをビルドします
# カレントディレクトリ（.）のGoソースコードをコンパイルし、
# binディレクトリに git-plus として出力します
echo "git-plusコマンドをビルド中..."
echo ""

printf "  ビルド中: git-plus... "

# go build コマンドを実行し、エラー出力を抑制
if go build -o "$BIN_PATH/git-plus" . 2>/dev/null; then
    echo "✓ OK"
else
    # ビルド失敗時はエラーメッセージを表示して終了
    echo "✗ FAILED"
    echo ""
    echo "エラー: ビルドに失敗しました"
    exit 1
fi

# ================================================================================
# ステップ3: 各コマンド用のシンボリックリンクを作成
# ================================================================================
# git-plus バイナリへの29個のシンボリックリンクを作成します
# Unix系OSでは、main.goの実行ファイル名判定機能により、
# シンボリックリンク名から対応するサブコマンドが推測されて実行されます
#
# 例: git-newbranch を実行すると、
#     main.goがファイル名から "newbranch" コマンドを推測して実行します
#
# シンボリックリンクを使用することで、ディスク容量を節約できます
# （Windows版は実行ファイルのコピーを使用）
commands=(
    "git-newbranch"
    "git-reset-tag"
    "git-amend"
    "git-squash"
    "git-track"
    "git-delete-local-branches"
    "git-undo-last-commit"
    "git-tag-diff"
    "git-tag-checkout"
    "git-stash-cleanup"
    "git-stash-select"
    "git-recent"
    "git-step"
    "git-sync"
    "git-pr-create-merge"
    "git-pr-merge"
    "git-pr-list"
    "git-pause"
    "git-resume"
    "git-create-repository"
    "git-new-tag"
    "git-browse"
    "git-pr-checkout"
    "git-clone-org"
    "git-batch-clone"
    "git-back"
    "git-issue-create"
    "git-issue-edit"
    "git-release-notes"
    "git-repo-others"
)

echo ""
echo "シンボリックリンクを作成中..."
echo ""

# 成功カウンタの初期化
success_count=0

# 各コマンド用にシンボリックリンクを作成
for cmd in "${commands[@]}"; do
    printf "  作成中: %-30s " "$cmd..."

    # 既存のシンボリックリンクやファイルを削除（クリーンインストール）
    rm -f "$BIN_PATH/$cmd"

    # git-plus へのシンボリックリンクを作成
    if ln -s "$BIN_PATH/git-plus" "$BIN_PATH/$cmd" 2>/dev/null; then
        echo "✓ OK"
        ((success_count++))
    else
        echo "✗ FAILED"
    fi
done

echo ""
echo -e "\033[32mシンボリックリンク作成完了: $success_count 個\033[0m"

# ================================================================================
# ステップ4: PATH環境変数への追加
# ================================================================================
# binディレクトリをシェルのPATH環境変数に追加します
# これにより、どこからでも git xxx コマンドを実行できるようになります
echo ""
echo "PATHの設定を確認中..."

# ================================================================================
# シェルの種類を検出して適切な設定ファイルを選択
# ================================================================================
# - ZSH: ~/.zshrc
# - Bash: ~/.bashrc
# - その他: ~/.profile (POSIX互換シェル)
if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
else
    SHELL_RC="$HOME/.profile"
fi

# binディレクトリが既にPATHに含まれているか確認
# コロンで囲んで検索することで、部分一致を防ぎます
if [[ ":$PATH:" != *":$BIN_PATH:"* ]]; then
    echo "PATHに追加中: $BIN_PATH"

    # シェル設定ファイルにPATH設定を追記
    echo "" >> "$SHELL_RC"
    echo "# Git extension commands" >> "$SHELL_RC"
    echo "export PATH=\"\$HOME/bin:\$PATH\"" >> "$SHELL_RC"

    # 現在のシェルセッションにも即座に反映
    # これにより、スクリプト実行後すぐにコマンドを使用可能になります
    export PATH="$BIN_PATH:$PATH"

    echo "✓ $SHELL_RC にPATHを追加しました"
    echo ""
    echo -e "\033[36m注意: 新しいターミナルセッションで有効になります\033[0m"
    echo -e "\033[36m      または 'source $SHELL_RC' を実行してください\033[0m"
else
    echo "✓ 既にPATHに含まれています"
fi

# ================================================================================
# セットアップ完了メッセージと使用例
# ================================================================================
echo ""
echo -e "\033[32m=== セットアップ完了 ===\033[0m"
echo ""
echo -e "\033[36m使用例:\033[0m"
echo "  git newbranch feature-xxx"
echo "  git new-tag feature"
echo "  git browse"
echo ""
echo -e "\033[36mヘルプを表示:\033[0m"
echo "  git newbranch -h"
echo "  git step -h"
echo ""
