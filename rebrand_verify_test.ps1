#!/usr/bin/env pwsh
# ==============================================================================
# Pranor Rebrand Verification Test Suite
# ==============================================================================
# Run: powershell -ExecutionPolicy Bypass -File rebrand_verify_test.ps1
# 
# This script scans the entire codebase for any remaining "serv" references
# that should have been renamed to "pranor" during the rebrand. It checks:
#   1. Old brand names (ServGate, ServStore, Servverse, etc.)
#   2. Old env var prefixes (SERV_, SERVGATE_, SERVQUEUE_, etc.)
#   3. Old file extensions (.srv)
#   4. Old binary names (serv.exe, serv-lsp, servd)
#   5. Old URL schemes (serv://, servgate://)
#   6. Old Docker/container references
#   7. Old GitHub repo URLs
#   8. Broken Go identifiers (camelCase serv*)
#   9. Broken URL schemes from replacement (Pranor Gate://)
#  10. Old module paths (github.com/vyuvaraj/serv/)
# ==============================================================================

$ErrorActionPreference = "Continue"
$root = $PSScriptRoot
if (-not $root) { $root = Get-Location }

$passed = 0
$failed = 0
$warnings = 0
$details = @()

function Test-Pattern {
    param(
        [string]$Name,
        [string]$Pattern,
        [string[]]$Include = @("*.go","*.md","*.yml","*.yaml","*.json","*.html","*.js","*.css","*.py","*.ps1","*.sh","*.bat","*.txt","*.xml","*.rb","*.toml","*.iss","*.nuspec","*.pnr"),
        [string[]]$ExcludePatterns = @(),
        [switch]$IsWarning
    )
    
    $hits = @()
    foreach ($ext in $Include) {
        $results = Select-String -Path "$root\**\$ext" -Pattern $Pattern -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike "*\.git*" -and $_.FullName -notlike "*node_modules*" -and $_.FullName -notlike "*rebrand_verify*" }
        if ($ExcludePatterns.Count -gt 0) {
            foreach ($excl in $ExcludePatterns) {
                $results = $results | Where-Object { $_.Line -notmatch $excl }
            }
        }
        if ($results) { $hits += $results }
    }
    
    if ($hits.Count -eq 0) {
        $script:passed++
        Write-Host "  PASS " -ForegroundColor Green -NoNewline
        Write-Host $Name
    } else {
        if ($IsWarning) {
            $script:warnings++
            Write-Host "  WARN " -ForegroundColor Yellow -NoNewline
        } else {
            $script:failed++
            Write-Host "  FAIL " -ForegroundColor Red -NoNewline
        }
        Write-Host "$Name ($($hits.Count) hits)"
        $hits | Select-Object -First 5 | ForEach-Object {
            $line = $_.Line.Trim()
            if ($line.Length -gt 80) { $line = $line.Substring(0, 80) + "..." }
            Write-Host "       $($_.Filename):$($_.LineNumber): $line" -ForegroundColor DarkGray
        }
        if ($hits.Count -gt 5) { Write-Host "       ... and $($hits.Count - 5) more" -ForegroundColor DarkGray }
        $script:details += @{ Name = $Name; Count = $hits.Count; Hits = $hits }
    }
}

Write-Host ""
Write-Host "=" * 70
Write-Host "  PRANOR REBRAND VERIFICATION TEST SUITE"
Write-Host "=" * 70
Write-Host "  Root: $root"
Write-Host "  Date: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Host "=" * 70
Write-Host ""

# ==============================================================================
# Category 1: PascalCase component names
# ==============================================================================
Write-Host "--- Category 1: PascalCase Component Names ---" -ForegroundColor Cyan
Test-Pattern "ServGate" "ServGate" -ExcludePatterns @("PranorGate")
Test-Pattern "ServStore" "ServStore" -ExcludePatterns @("PranorVault")
Test-Pattern "ServQueue" "ServQueue" -ExcludePatterns @("PranorPulse")
Test-Pattern "ServCron" "ServCron" -ExcludePatterns @("PranorChrono")
Test-Pattern "ServAuth" "ServAuth" -ExcludePatterns @("PranorAuth")
Test-Pattern "ServCache" "ServCache" -ExcludePatterns @("PranorCache")
Test-Pattern "ServMesh" "ServMesh" -ExcludePatterns @("PranorMesh")
Test-Pattern "ServTrace" "ServTrace" -ExcludePatterns @("PranorTrace")
Test-Pattern "ServConsole" "ServConsole" -ExcludePatterns @("PranorConsole")
Test-Pattern "ServPool" "ServPool" -ExcludePatterns @("PranorPool")
Test-Pattern "ServMail" "ServMail" -ExcludePatterns @("PranorNotify")
Test-Pattern "ServFlow" "ServFlow" -ExcludePatterns @("PranorFlow")
Test-Pattern "ServCloud" "ServCloud" -ExcludePatterns @("PranorDeploy")
Test-Pattern "ServTunnel" "ServTunnel" -ExcludePatterns @("PranorTunnel")
Test-Pattern "ServShared" "ServShared" -ExcludePatterns @("PranorCore")
Test-Pattern "ServLock" "ServLock" -ExcludePatterns @("PranorLock")
Test-Pattern "ServSecret" "ServSecret" -ExcludePatterns @("PranorSecret")
Test-Pattern "ServRegistry" "ServRegistry" -ExcludePatterns @("PranorHub")
Test-Pattern "Serv-lang" "Serv-lang"
Test-Pattern "Servverse" "Servverse|ServVerse|servverse"
Write-Host ""

# ==============================================================================
# Category 2: Lowercase service/binary names
# ==============================================================================
Write-Host "--- Category 2: Lowercase Service Names ---" -ForegroundColor Cyan
Test-Pattern "servgate" "\bservgate\b" -ExcludePatterns @("service|server|observe|reserved")
Test-Pattern "servstore" "\bservstore\b"
Test-Pattern "servqueue" "\bservqueue\b"
Test-Pattern "servcron" "\bservcron\b"
Test-Pattern "servauth" "\bservauth\b"
Test-Pattern "servcache" "\bservcache\b"
Test-Pattern "servmesh" "\bservmesh\b"
Test-Pattern "servtrace" "\bservtrace\b"
Test-Pattern "servconsole" "\bservconsole\b"
Test-Pattern "servpool" "\bservpool\b"
Test-Pattern "servmail" "\bservmail\b"
Test-Pattern "servflow" "\bservflow\b"
Test-Pattern "servcloud" "\bservcloud\b"
Test-Pattern "servtunnel" "\bservtunnel\b"
Test-Pattern "servregistry" "\bservregistry\b"
Test-Pattern "servlockctl" "\bservlockctl\b"
Test-Pattern "servsecretctl" "\bservsecretctl\b"
Write-Host ""

# ==============================================================================
# Category 3: Environment variable prefixes
# ==============================================================================
Write-Host "--- Category 3: Environment Variables ---" -ForegroundColor Cyan
Test-Pattern "SERV_ prefix" "\bSERV_" -ExcludePatterns @("PRANOR_")
Test-Pattern "SERVGATE_ prefix" "SERVGATE_"
Test-Pattern "SERVQUEUE_ prefix" "SERVQUEUE_"
Test-Pattern "SERVSTORE_ prefix" "SERVSTORE_"
Test-Pattern "SERVAUTH_ prefix" "SERVAUTH_"
Test-Pattern "SERVCACHE_ prefix" "SERVCACHE_"
Test-Pattern "SERVTRACE_ prefix" "SERVTRACE_"
Test-Pattern "SERVMESH_ prefix" "SERVMESH_"
Test-Pattern "SERVVERSE_ prefix" "SERVVERSE_"
Test-Pattern "SERVLOCK_ prefix" "SERVLOCK_"
Test-Pattern "SERVSECRET_ prefix" "SERVSECRET_"
Write-Host ""

# ==============================================================================
# Category 4: File extension and binary names
# ==============================================================================
Write-Host "--- Category 4: File Extensions & Binaries ---" -ForegroundColor Cyan
Test-Pattern ".srv extension" "\.srv\b" -ExcludePatterns @("observe|reserved|conserv|preserv")
Test-Pattern "serv.exe binary" "\bserv\.exe\b"
Test-Pattern "serv-lsp binary" "\bserv-lsp\b"
Test-Pattern "servd daemon" "\bservd\b" -ExcludePatterns @("observed|reserved|conserved|preserved")
Write-Host ""

# ==============================================================================
# Category 5: URL schemes
# ==============================================================================
Write-Host "--- Category 5: URL Schemes ---" -ForegroundColor Cyan
Test-Pattern "serv:// scheme" 'serv://' -ExcludePatterns @("pranor://|observe://|reserved://")
Test-Pattern "Pranor Gate:// (broken)" "Pranor Gate://"
Test-Pattern "Pranor Vault:// (broken)" "Pranor Vault://"
Test-Pattern "Pranor Pulse:// (broken)" "Pranor Pulse://"
Write-Host ""

# ==============================================================================
# Category 6: Docker & Container references
# ==============================================================================
Write-Host "--- Category 6: Docker References ---" -ForegroundColor Cyan
Test-Pattern "ghcr.io/vyuvaraj/serv" "ghcr\.io/vyuvaraj/serv"
Test-Pattern "servverse-net network" "servverse-net"
Test-Pattern "docker serv- prefix" 'docker.*"serv-'
Write-Host ""

# ==============================================================================
# Category 7: GitHub/Module paths
# ==============================================================================
Write-Host "--- Category 7: GitHub & Module Paths ---" -ForegroundColor Cyan
Test-Pattern "vyuvaraj/serv/ (old repo)" "vyuvaraj/serv/" -ExcludePatterns @("vyuvaraj/pranor")
Test-Pattern "vyuvaraj/serv-ee" "vyuvaraj/serv-ee"
Test-Pattern "packages/Serv (old path)" "packages/Serv"
Test-Pattern "homebrew-serv" "homebrew-serv"
Test-Pattern "scoop-serv" "scoop-serv"
Write-Host ""

# ==============================================================================
# Category 8: camelCase Go identifiers
# ==============================================================================
Write-Host "--- Category 8: camelCase Identifiers (Go files) ---" -ForegroundColor Cyan
Test-Pattern "servGate (camelCase)" "\bservGate" -Include @("*.go")
Test-Pattern "servStore (camelCase)" "\bservStore" -Include @("*.go")
Test-Pattern "servQueue (camelCase)" "\bservQueue" -Include @("*.go")
Test-Pattern "servAuth (camelCase)" "\bservAuth" -Include @("*.go")
Test-Pattern "servCache (camelCase)" "\bservCache" -Include @("*.go")
Test-Pattern "servMesh (camelCase)" "\bservMesh" -Include @("*.go")
Test-Pattern "servCron (camelCase)" "\bservCron" -Include @("*.go")
Test-Pattern "servTrace (camelCase)" "\bservTrace" -Include @("*.go")
Test-Pattern "servFlow (camelCase)" "\bservFlow" -Include @("*.go")
Test-Pattern "servPool (camelCase)" "\bservPool" -Include @("*.go")
Test-Pattern "servMail (camelCase)" "\bservMail" -Include @("*.go")
Test-Pattern "servCloud (camelCase)" "\bservCloud" -Include @("*.go")
Test-Pattern "servTunnel (camelCase)" "\bservTunnel" -Include @("*.go")
Write-Host ""

# ==============================================================================
# Category 9: Internal references
# ==============================================================================
Write-Host "--- Category 9: Internal References ---" -ForegroundColor Cyan
Test-Pattern "_serv_migrations table" "_serv_migrations"
Test-Pattern ".serv/ config dir" '\.serv/'
Test-Pattern ".serv-build-cache" "\.serv-build-cache"
Test-Pattern "serv-build module" "\bserv-build\b"
Test-Pattern "serv/runtime import" '"serv/runtime"'
Write-Host ""

# ==============================================================================
# Category 10: Standalone "serv" word (warning - may have false positives)
# ==============================================================================
Write-Host "--- Category 10: Standalone 'serv' (informational) ---" -ForegroundColor Cyan
Test-Pattern "Standalone 'serv' word" "\bserv\b" -ExcludePatterns @("service|server|observe|reserved|deserve|conserve|preserve|Observe|Server|Reserve") -IsWarning
Write-Host ""

# ==============================================================================
# Category 11: Build verification
# ==============================================================================
Write-Host "--- Category 11: Build Verification ---" -ForegroundColor Cyan
$langDir = Join-Path $root "lang"
if (Test-Path $langDir) {
    Push-Location $langDir
    $buildOutput = & go build -o pranor.exe . 2>&1
    if ($LASTEXITCODE -eq 0) {
        $script:passed++
        Write-Host "  PASS " -ForegroundColor Green -NoNewline
        Write-Host "go build -o pranor.exe . (compiles successfully)"
        Remove-Item "pranor.exe" -Force -ErrorAction SilentlyContinue
    } else {
        $script:failed++
        Write-Host "  FAIL " -ForegroundColor Red -NoNewline
        Write-Host "go build failed:"
        $buildOutput | Select-Object -First 10 | ForEach-Object { Write-Host "       $_" -ForegroundColor DarkGray }
    }
    Pop-Location
} else {
    Write-Host "  SKIP " -ForegroundColor Yellow -NoNewline
    Write-Host "lang/ directory not found"
}
Write-Host ""

# ==============================================================================
# Category 12: File system checks
# ==============================================================================
Write-Host "--- Category 12: File System ---" -ForegroundColor Cyan
$srvFiles = Get-ChildItem -Path $root -Filter "*.srv" -Recurse -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike "*\.git*" }
if ($srvFiles.Count -eq 0) {
    $script:passed++
    Write-Host "  PASS " -ForegroundColor Green -NoNewline
    Write-Host "No .srv files exist on disk"
} else {
    $script:failed++
    Write-Host "  FAIL " -ForegroundColor Red -NoNewline
    Write-Host "Found $($srvFiles.Count) .srv files:"
    $srvFiles | Select-Object -First 5 | ForEach-Object { Write-Host "       $($_.FullName.Replace($root + '\', ''))" -ForegroundColor DarkGray }
}

$exeFiles = Get-ChildItem -Path $root -Filter "*.exe" -Recurse -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notlike "*\.git*" }
if ($exeFiles.Count -eq 0) {
    $script:passed++
    Write-Host "  PASS " -ForegroundColor Green -NoNewline
    Write-Host "No stale .exe files committed"
} else {
    $script:warnings++
    Write-Host "  WARN " -ForegroundColor Yellow -NoNewline
    Write-Host "Found $($exeFiles.Count) .exe files (should be in .gitignore)"
    $exeFiles | Select-Object -First 3 | ForEach-Object { Write-Host "       $($_.FullName.Replace($root + '\', ''))" -ForegroundColor DarkGray }
}
Write-Host ""

# ==============================================================================
# Summary
# ==============================================================================
Write-Host "=" * 70
Write-Host "  RESULTS"
Write-Host "=" * 70
Write-Host "  Passed:   $passed" -ForegroundColor Green
Write-Host "  Failed:   $failed" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Green" })
Write-Host "  Warnings: $warnings" -ForegroundColor $(if ($warnings -gt 0) { "Yellow" } else { "Green" })
Write-Host "  Total:    $($passed + $failed + $warnings)"
Write-Host "=" * 70

if ($failed -gt 0) {
    Write-Host ""
    Write-Host "  REBRAND INCOMPLETE - $failed test(s) failed" -ForegroundColor Red
    exit 1
} else {
    Write-Host ""
    Write-Host "  REBRAND VERIFIED - all checks passed" -ForegroundColor Green
    exit 0
}
