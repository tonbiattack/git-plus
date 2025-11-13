#!/bin/bash
# Git Plus セットアップスクリプト (Linux/macOS)
# git-plusコマンドをビルドしてPATHに追加します

set -e

echo "=== Git Plus セットアップ ==="
echo ""

# ~/bin にビルド（存在しない場合は作成）
BIN_PATH="$HOME/bin"
echo "binディレクトリを作成中: $BIN_PATH"
mkdir -p "$BIN_PATH"

echo ""
echo "git-plusコマンドをビルド中..."
echo ""

printf "  ビルド中: git-plus... "

if go build -o "$BIN_PATH/git-plus" . 2>/dev/null; then
    echo "✓ OK"
else
    echo "✗ FAILED"
    echo ""
    echo "エラー: ビルドに失敗しました"
    exit 1
fi

# 各コマンド用のシンボリックリンクを作成
commands=(
    "git-newbranch"
    "git-reset-tag"
    "git-amend"
    "git-squash"
    "git-track"
    "git-delete-local-branches"
    "git-undo-last-commit"
    "git-tag-diff"
    "git-stash-cleanup"
    "git-recent"
    "git-step"
    "git-sync"
    "git-pr-merge"
    "git-pause"
    "git-resume"
    "git-create-repository"
    "git-new-tag"
    "git-browse"
    "git-pr-checkout"
    "git-clone-org"
)

echo ""
echo "シンボリックリンクを作成中..."
echo ""

success_count=0
for cmd in "${commands[@]}"; do
    printf "  作成中: %-30s " "$cmd..."

    # 既存のシンボリックリンクやファイルを削除
    rm -f "$BIN_PATH/$cmd"

    if ln -s "$BIN_PATH/git-plus" "$BIN_PATH/$cmd" 2>/dev/null; then
        echo "✓ OK"
        ((success_count++))
    else
        echo "✗ FAILED"
    fi
done

echo ""
echo -e "\033[32mシンボリックリンク作成完了: $success_count 個\033[0m"

# PATHに追加
echo ""
echo "PATHの設定を確認中..."

# シェルの設定ファイルを検出
if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
else
    SHELL_RC="$HOME/.profile"
fi

# PATHに含まれているか確認
if [[ ":$PATH:" != *":$BIN_PATH:"* ]]; then
    echo "PATHに追加中: $BIN_PATH"
    echo "" >> "$SHELL_RC"
    echo "# Git Plus commands" >> "$SHELL_RC"
    echo "export PATH=\"\$HOME/bin:\$PATH\"" >> "$SHELL_RC"

    # 現在のセッションにも反映
    export PATH="$BIN_PATH:$PATH"

    echo "✓ $SHELL_RC にPATHを追加しました"
    echo ""
    echo -e "\033[36m注意: 新しいターミナルセッションで有効になります\033[0m"
    echo -e "\033[36m      または 'source $SHELL_RC' を実行してください\033[0m"
else
    echo "✓ 既にPATHに含まれています"
fi

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
