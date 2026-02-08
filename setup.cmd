@echo off
setlocal EnableExtensions EnableDelayedExpansion

REM ================================================================================
REM Git extension commands setup script (Windows CMD)
REM ================================================================================
REM For environments where PowerShell is restricted.
REM
REM Steps:
REM 1. Create %USERPROFILE%\bin
REM 2. Build git-plus.exe
REM 3. Copy git-plus.exe to git-xxx command names
REM 4. Add bin to user PATH (with duplicate check)
REM
REM Usage:
REM   setup.cmd
REM
REM Requirements:
REM   - Go 1.25.3 or later
REM ================================================================================

echo === Git extension commands setup (CMD) ===
echo.

REM Step 1: Create bin directory
set "binPath=%USERPROFILE%\bin"
echo Creating bin directory: %binPath%
if not exist "%binPath%" mkdir "%binPath%"
if errorlevel 1 (
  echo ERROR: Failed to create bin directory
  exit /b 1
)
echo.

REM Step 2: Build git-plus.exe
echo Building git-plus...
echo.
echo   Building: git-plus
go build -o "%binPath%\git-plus.exe" .
if errorlevel 1 (
  echo FAILED
  echo.
  echo ERROR: Build failed
  exit /b 1
) else (
  echo OK
)

REM Step 3: Copy executables for each command
set "commands=git-newbranch git-rename-branch git-reset-tag git-amend git-squash git-track git-delete-local-branches git-undo-last-commit git-tag-diff git-tag-diff-all git-tag-checkout git-stash-cleanup git-stash-select git-recent git-step git-sync git-pr-create-merge git-pr-merge git-pr-list git-pause git-resume git-create-repository git-new-tag git-browse git-pr-checkout git-clone-org git-batch-clone git-abort git-issue-list git-issue-create git-issue-edit git-issue-bulk-close git-release-notes git-repo-others git-pr-browse git-pr-issue-link git-worktree-new git-worktree-switch git-worktree-delete"

echo.
echo Creating command copies...
echo.

set /a successCount=0
for %%C in (%commands%) do (
  echo   Creating: %%C
  if exist "%binPath%\%%C.exe" del /f /q "%binPath%\%%C.exe" >nul 2>&1
  copy /y "%binPath%\git-plus.exe" "%binPath%\%%C.exe" >nul 2>&1
  if exist "%binPath%\%%C.exe" (
    echo     OK
    set /a successCount+=1
  ) else (
    echo     FAILED
  )
)

echo.
echo Copies created: %successCount%

REM Step 4: Add bin to PATH
echo.
echo Checking PATH...
echo %PATH% | find /I "%binPath%" >nul 2>&1
if errorlevel 1 (
  echo Adding to PATH: %binPath%
  REM Persist user PATH (setx does not update current session)
  setx PATH "%binPath%;%PATH%" >nul
  set "PATH=%binPath%;%PATH%"
  echo OK: PATH updated
  echo.
  echo NOTE: Restart other command prompts to pick up PATH changes
) else (
  echo OK: PATH already includes bin
)

REM Done
echo.
echo === Setup complete ===
echo.
echo Examples:
echo   git newbranch feature-xxx
echo   git new-tag feature
echo   git browse
echo.
echo Help:
echo   git newbranch -h
echo   git step -h
echo.
endlocal
