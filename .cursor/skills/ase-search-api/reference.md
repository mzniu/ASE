# ASE Search API — reference snippets

Copy this file next to `SKILL.md` when your agent supports multiple docs. Installation paths are product-specific — see **Installing this skill (any agent)** in `SKILL.md`.

## Environment

| Variable | Purpose |
|----------|---------|
| `ASE_BASE_URL` | Service origin, e.g. `http://127.0.0.1:18080` |
| `ASE_API_KEY` | Bearer token for `/v1/*` |

## PowerShell (Windows)

```powershell
$base = if ($env:ASE_BASE_URL) { $env:ASE_BASE_URL } else { "http://127.0.0.1:18080" }
$key = $env:ASE_API_KEY
Invoke-RestMethod -Uri "$base/health" -Method Get
$body = '{"query":"hello","providers":["stub"]}'
Invoke-RestMethod -Uri "$base/v1/search" -Method Post -ContentType "application/json" -Headers @{ Authorization = "Bearer $key" } -Body $body
```

## Problem response shape

```json
{
  "type": "about:blank",
  "title": "short label",
  "detail": "human-readable detail",
  "status": 400
}
```
