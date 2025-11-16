# ================================================================================
# Git 拡張コマンド セットアップスクリプト (Windows)
# ================================================================================
# このスクリプトは以下の処理を実行します：
# 1. ユーザーディレクトリに bin フォルダを作成
# 2. git-plus.exe をビルド
# 3. 28個のgit拡張コマンド用に実行ファイルをコピー
# 4. ユーザーのPATH環境変数にbinディレクトリを追加
#
# これにより、git newbranch、git pr-merge などのコマンドが使用可能になります。
#
# 使用方法:
#   .\setup.ps1
#
# 前提条件:
#   - Go 1.25.3以上がインストールされていること
#   - PowerShell実行ポリシーが適切に設定されていること
# ================================================================================

Write-Host "=== Git 拡張コマンド セットアップ ===" -ForegroundColor Cyan
Write-Host ""

# ================================================================================
# ステップ1: binディレクトリの作成
# ================================================================================
# ユーザープロファイル配下に bin フォルダを作成します
# 既に存在する場合はエラーを抑制します（-Force オプション）
$binPath = "$env:USERPROFILE\bin"
Write-Host "binディレクトリを作成中: $binPath" -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $binPath | Out-Null

Write-Host ""

# ================================================================================
# ステップ2: git-plus.exe のビルド
# ================================================================================
# Goコンパイラを使用してgit-plus.exeをビルドします
# カレントディレクトリ（.）のGoソースコードをコンパイルし、
# binディレクトリに git-plus.exe として出力します
Write-Host "git-plusコマンドをビルド中..." -ForegroundColor Yellow
Write-Host ""

Write-Host "  ビルド中: git-plus... " -NoNewline

# go build コマンドを実行し、標準出力とエラー出力の両方をキャプチャ
$output = go build -o "$binPath\git-plus.exe" . 2>&1

# ビルド結果を確認
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ OK" -ForegroundColor Green
} else {
    # ビルド失敗時はエラーメッセージを表示して終了
    Write-Host "✗ FAILED" -ForegroundColor Red
    Write-Host ""
    Write-Host "エラー: ビルドに失敗しました" -ForegroundColor Red
    Write-Host "    $output" -ForegroundColor Red
    exit 1
}

# ================================================================================
# ステップ3: 各コマンド用の実行ファイルをコピー
# ================================================================================
# git-plus.exe を30個のgit-xxxコマンド用にコピーします
# Windowsでは、main.goの実行ファイル名判定機能により、
# コピーした各実行ファイルが対応するサブコマンドとして動作します
#
# 例: git-newbranch.exe を実行すると、
#     main.goがファイル名から "newbranch" コマンドを推測して実行します
$commands = @(
    "git-newbranch",
    "git-reset-tag",
    "git-amend",
    "git-squash",
    "git-track",
    "git-delete-local-branches",
    "git-undo-last-commit",
    "git-tag-diff",
    "git-tag-diff-all",
    "git-tag-checkout",
    "git-stash-cleanup",
    "git-stash-select",
    "git-recent",
    "git-step",
    "git-sync",
    "git-pr-create-merge",
    "git-pr-merge",
    "git-pr-list",
    "git-pause",
    "git-resume",
    "git-create-repository",
    "git-new-tag",
    "git-browse",
    "git-pr-checkout",
    "git-clone-org",
    "git-batch-clone",
    "git-back",
    "git-issue-create",
    "git-issue-edit",
    "git-release-notes",
    "git-repo-others"
)

Write-Host ""
Write-Host "コマンドのコピーを作成中..." -ForegroundColor Yellow
Write-Host ""

# 成功カウンタの初期化
$successCount = 0

# 各コマンド用に git-plus.exe をコピー
foreach ($cmd in $commands) {
    Write-Host "  作成中: $cmd... " -NoNewline

    # 既存のファイルがある場合は削除（クリーンインストール）
    $targetPath = "$binPath\$cmd.exe"
    if (Test-Path $targetPath) {
        Remove-Item $targetPath -Force
    }

    # git-plus.exe をターゲット名でコピー
    Copy-Item "$binPath\git-plus.exe" $targetPath

    # コピー結果を確認
    if ($LASTEXITCODE -eq 0 -or (Test-Path $targetPath)) {
        Write-Host "✓ OK" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "✗ FAILED" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "コピー作成完了: $successCount 個" -ForegroundColor Green

# ================================================================================
# ステップ4: PATH環境変数への追加
# ================================================================================
# binディレクトリをユーザーのPATH環境変数に追加します
# これにより、どこからでも git xxx コマンドを実行できるようになります
Write-Host ""
Write-Host "PATHの設定を確認中..." -ForegroundColor Yellow

# ユーザー環境変数のPATHを取得
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

# binディレクトリが既にPATHに含まれているか確認
if ($currentPath -notlike "*$binPath*") {
    Write-Host "PATHに追加中: $binPath" -ForegroundColor Yellow

    # 既存のPATHの先頭にbinディレクトリを追加
    # 既存のPATHがない場合は、binディレクトリのみを設定
    $newPath = if ($currentPath) {
        "$binPath;$currentPath"
    } else {
        $binPath
    }

    # ユーザー環境変数を永続的に更新
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    # 現在のPowerShellセッションにも即座に反映
    # これにより、スクリプト実行後すぐにコマンドを使用可能になります
    $env:Path = "$binPath;$env:Path"

    Write-Host "✓ PATHに追加しました" -ForegroundColor Green
    Write-Host ""
    Write-Host "注意: 他のPowerShellウィンドウで使用する場合は再起動が必要です" -ForegroundColor Cyan
} else {
    Write-Host "✓ 既にPATHに含まれています" -ForegroundColor Green
}

# ================================================================================
# セットアップ完了メッセージと使用例
# ================================================================================
Write-Host ""
Write-Host "=== セットアップ完了 ===" -ForegroundColor Green
Write-Host ""
Write-Host "使用例:" -ForegroundColor Cyan
Write-Host "  git newbranch feature-xxx"
Write-Host "  git new-tag feature"
Write-Host "  git browse"
Write-Host ""
Write-Host "ヘルプを表示:" -ForegroundColor Cyan
Write-Host "  git newbranch -h"
Write-Host "  git step -h"
Write-Host ""
