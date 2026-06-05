#!/usr/bin/env pwsh
# Zero Bootstrap Script
# Builds the bootstrap compiler and verifies the toolchain.

$ErrorActionPreference = "Stop"
$RootDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BuildDir = Join-Path $RootDir "build"

Write-Host "=== Zero Bootstrap ===" -ForegroundColor Cyan

# Step 1: Build Go bootstrap compiler
Write-Host "[1/4] Building Go bootstrap compiler..." -ForegroundColor Yellow
$null = New-Item -ItemType Directory -Path $BuildDir -Force
Set-Location $RootDir
go build -o (Join-Path $BuildDir "bootstrap.exe") ./go/main.go
if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
Write-Host "  -> $BuildDir\bootstrap.exe" -ForegroundColor Green

# Step 2: Run tests
Write-Host "[2/4] Running tests..." -ForegroundColor Yellow
$testDir = Join-Path $RootDir "tests"
if (Test-Path $testDir) {
    Get-ChildItem $testDir -Filter "*.zero" | ForEach-Object {
        $testFile = $_.FullName
        Write-Host "  Testing: $($_.Name)" -ForegroundColor Gray
        & (Join-Path $BuildDir "bootstrap.exe") run $testFile
        if ($LASTEXITCODE -ne 0) { throw "Test failed: $($_.Name)" }
    }
    Write-Host "  All tests passed" -ForegroundColor Green
} else {
    Write-Host "  No tests directory found, skipping" -ForegroundColor Yellow
}

# Step 3: Check if self-hosted compiler exists
$srcDir = Join-Path $RootDir "src" "compiler"
if (Test-Path $srcDir) {
    Write-Host "[3/4] Compiling self-hosted compiler..." -ForegroundColor Yellow
    $zeroFiles = Get-ChildItem $srcDir -Filter "*.zero" | ForEach-Object { $_.FullName }
    if ($zeroFiles.Count -gt 0) {
        & (Join-Path $BuildDir "bootstrap.exe") build -o (Join-Path $BuildDir "compiler.zbc") @zeroFiles
        if ($LASTEXITCODE -ne 0) { throw "Self-hosted compiler build failed" }
        Write-Host "  -> $BuildDir\compiler.zbc" -ForegroundColor Green
    }
} else {
    Write-Host "[3/4] No self-hosted compiler yet, skipping" -ForegroundColor Yellow
}

# Step 4: Summary
Write-Host "[4/4] Bootstrap complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "Usage:" -ForegroundColor White
Write-Host "  build\bootstrap.exe run file.zero   - Run a Zero program"
Write-Host "  build\bootstrap.exe build file.zero - Compile to .zbc bytecode"
Write-Host "  build\bootstrap.exe repl            - Start interactive REPL"