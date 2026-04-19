#Requires -Version 5.1
<#
.SYNOPSIS
    拉取最新 ASE 代码，重建并轮换 Docker 中的 ase 服务（OpenSearch 保持运行）。

.DESCRIPTION
    在仓库根目录执行。先构建镜像（旧容器继续服务），再 up --force-recreate 轮换，缩短不可用时间。
    等价于：git pull → docker compose build ase → up -d --no-deps --force-recreate ase

.PARAMETER RepoRoot
    ASE 仓库根路径（含 docker-compose.yml）。默认为本脚本上两级目录。

.PARAMETER NoCache
    构建镜像时使用 --no-cache。

.EXAMPLE
    .\scripts\deploy\restart-ase.ps1
.EXAMPLE
    .\scripts\deploy\restart-ase.ps1 -NoCache
#>
param(
    [string] $RepoRoot = "",
    [switch] $NoCache
)

$ErrorActionPreference = "Stop"

if (-not $RepoRoot) {
    $RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
}
Set-Location $RepoRoot

if (-not (Test-Path (Join-Path $RepoRoot "docker-compose.yml"))) {
    Write-Error "未找到 docker-compose.yml（RepoRoot: $RepoRoot）"
}

function Test-DockerComposePlugin {
    $null = docker compose version 2>&1
    return ($LASTEXITCODE -eq 0)
}

function Invoke-DockerCompose {
    param([Parameter(Mandatory)][string[]] $ComposeArgs)
    if (Test-DockerComposePlugin) {
        & docker compose @ComposeArgs
    } elseif (Get-Command docker-compose -ErrorAction SilentlyContinue) {
        & docker-compose @ComposeArgs
    } else {
        Write-Error "未检测到 Docker Compose（需要 docker compose 或 docker-compose）。"
    }
}

$null = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Error "无法连接 Docker，请先启动 Docker Desktop 或 Docker 守护进程。"
}

if (Test-Path (Join-Path $RepoRoot ".git")) {
    Write-Host ">>> git pull（当前分支）"
    git -C $RepoRoot pull --ff-only
} else {
    Write-Host "提示：当前目录不是 git 仓库，跳过 git pull。"
}

Write-Host ">>> 构建 ase 镜像（构建期间旧容器仍运行）"
if ($NoCache) {
    Invoke-DockerCompose -ComposeArgs @("build", "--no-cache", "ase")
} else {
    Invoke-DockerCompose -ComposeArgs @("build", "ase")
}

Write-Host ">>> 轮换 ase 容器（不拉起/重启 opensearch；中断仅发生在停旧启新）"
Invoke-DockerCompose -ComposeArgs @("up", "-d", "--no-deps", "--force-recreate", "ase")

Write-Host ""
Invoke-DockerCompose -ComposeArgs @("ps")
Write-Host ""

$port = $env:ASE_HOST_PORT
if (-not $port) { $port = "18080" }
$healthUrl = "http://127.0.0.1:$port/health"
Write-Host "探活: $healthUrl"
try {
    $r = Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec 5
    Write-Host $r.Content
} catch {
    Write-Host "（若失败请检查端口映射、ASE_HOST_PORT 与防火墙）"
}
