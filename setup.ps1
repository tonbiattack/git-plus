# Git Plus セットアップスクリプト (Windows)
# git-plusコマンドをビルドしてPATHに追加します

Write-Host "=== Git Plus セットアップ ===" -ForegroundColor Cyan
Write-Host ""

# ユーザーディレクトリ配下に bin フォルダを作成
$binPath = "$env:USERPROFILE\bin"
Write-Host "binディレクトリを作成中: $binPath" -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $binPath | Out-Null

Write-Host ""
Write-Host "git-plusコマンドをビルド中..." -ForegroundColor Yellow
Write-Host ""

Write-Host "  ビルド中: git-plus... " -NoNewline

$output = go build -o "$binPath\git-plus.exe" . 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ OK" -ForegroundColor Green
} else {
    Write-Host "✗ FAILED" -ForegroundColor Red
    Write-Host ""
    Write-Host "エラー: ビルドに失敗しました" -ForegroundColor Red
    Write-Host "    $output" -ForegroundColor Red
    exit 1
}

# 各コマンド用のコピーを作成
$commands = @(
    "git-newbranch",
    "git-reset-tag",
    "git-amend",
    "git-squash",
    "git-track",
    "git-delete-local-branches",
    "git-undo-last-commit",
    "git-tag-diff",
    "git-stash-cleanup",
    "git-recent",
    "git-step",
    "git-sync",
    "git-pr-merge",
    "git-pause",
    "git-resume",
    "git-create-repository",
    "git-new-tag",
    "git-browse",
    "git-pr-checkout",
    "git-clone-org"
)

Write-Host ""
Write-Host "コマンドのコピーを作成中..." -ForegroundColor Yellow
Write-Host ""

$successCount = 0

foreach ($cmd in $commands) {
    Write-Host "  作成中: $cmd... " -NoNewline

    # 既存のファイルを削除
    $targetPath = "$binPath\$cmd.exe"
    if (Test-Path $targetPath) {
        Remove-Item $targetPath -Force
    }

    Copy-Item "$binPath\git-plus.exe" $targetPath

    if ($LASTEXITCODE -eq 0 -or (Test-Path $targetPath)) {
        Write-Host "✓ OK" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "✗ FAILED" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "コピー作成完了: $successCount 個" -ForegroundColor Green

# PATHに追加
Write-Host ""
Write-Host "PATHの設定を確認中..." -ForegroundColor Yellow

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($currentPath -notlike "*$binPath*") {
    Write-Host "PATHに追加中: $binPath" -ForegroundColor Yellow

    $newPath = if ($currentPath) {
        "$binPath;$currentPath"
    } else {
        $binPath
    }

    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    # 現在のセッションにも反映
    $env:Path = "$binPath;$env:Path"

    Write-Host "✓ PATHに追加しました" -ForegroundColor Green
    Write-Host ""
    Write-Host "注意: 他のPowerShellウィンドウで使用する場合は再起動が必要です" -ForegroundColor Cyan
} else {
    Write-Host "✓ 既にPATHに含まれています" -ForegroundColor Green
}

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
