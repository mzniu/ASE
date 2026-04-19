---
name: ase-search-api
description: >-
  Calls the ASE REST API for agent-oriented web search: POST /v1/search returns Markdown;
  optional index ingest via POST /v1/documents. Use when loading this skill in any AI agent or
  IDE that supports custom skills (Cursor, Claude Code, Copilot extensions, bespoke agents), or
  when the user mentions ASE search, ASE_BASE_URL, or DEV_API_KEY.
---

# ASE Search API (Agent)

## Installing this skill (any agent)

- **What to copy**: this directory’s **`SKILL.md`** (and optionally **`reference.md`**) from the ASE repository path **`.cursor/skills/ase-search-api/`** — the `.cursor` segment is only the **repo folder name**; you are **not** required to use Cursor.
- **Or download from your ASE server** (no GitHub; same host as the API): `GET /skills/ase-search-api/SKILL.md` and `GET /skills/ase-search-api/reference.md`, or `GET /skills/ase-search-api/bundle.zip` for both files.
- **Where to put it**: your product’s documented **skills / rules / tools** directory (create a subfolder e.g. `ase-search-api` and place `SKILL.md` inside). Examples:
  - **Cursor**: `~/.cursor/skills/ase-search-api/SKILL.md` (Windows: `%USERPROFILE%\.cursor\skills\ase-search-api\SKILL.md`).
  - **Other agents**: follow that tool’s docs; keep the **YAML frontmatter** and body semantics unchanged.
- **After install**: set **`ASE_BASE_URL`** and **`ASE_API_KEY`** (or deployment key) in the environment or as documented by your agent.

## Preconditions

- **Base URL**: from env **`ASE_BASE_URL`**, default `http://127.0.0.1:18080` (no trailing slash).
- **API key**: env **`ASE_API_KEY`** (or **`DEV_API_KEY`** if that is what the deployment documents). Send as `Authorization: Bearer <token>`.

## Quick health check

```bash
curl -sS "${ASE_BASE_URL:-http://127.0.0.1:18080}/health"
```

Expect `200` and JSON containing `"status":"ok"`.

- **HTML project homepage**: `GET /` (browser or `curl` — embedded docs + Skill setup UI).
- **JSON discovery**: `GET /api/info` — service name and path links for tooling.

## Primary: search (Markdown)

`POST /v1/search` — body JSON, response **plain Markdown** (`200`, `Content-Type: text/markdown; charset=utf-8`).

Headers:

- `Content-Type: application/json`
- `Authorization: Bearer <API_KEY>`

Body:

```json
{
  "query": "natural language question",
  "providers": ["stub"]
}
```

- **`query`**: required, non-empty string.
- **`providers`**: optional string array (`baidu`, `bing`, `google`, `tavily`, `stub`, …). Omit to use server defaults.

**curl example** (bash; set `ASE_API_KEY`):

```bash
curl -sS -X POST "${ASE_BASE_URL:-http://127.0.0.1:18080}/v1/search" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ASE_API_KEY}" \
  -d '{"query":"hello world","providers":["stub"]}'
```

On failure the API returns **`application/problem+json`** (RFC 9457-style): read `status`, `title`, `detail`. Common cases: **401** auth, **400** validation, **429** rate limit, **504** deadline, **503** dependency unavailable.

## Optional: index document (OpenSearch)

`POST /v1/documents` — JSON body, **204** on success. Returns **501** if indexing is not configured.

```json
{
  "id": "doc-id",
  "title": "Title",
  "body_text": "Plain text body for the index"
}
```

## Optional: metrics

`GET /metrics` — Prometheus text; no auth. Use for ops, not for normal agent Q&A.

## Agent workflow

1. Confirm **`ASE_BASE_URL`** (and key) with the user if unknown.
2. Run **health** when connectivity is unclear.
3. Call **POST /v1/search**; pass the user’s question in **`query`**; choose **`providers`** only if the user specified engines or the task requires a named provider.
4. Return the **Markdown body** to the user (or summarize if they asked for a summary). On errors, surface **`title`/`detail`** from the problem JSON when helpful.

## More detail

- Product and error semantics: see your ASE deployment’s API docs or repository `docs/SEARCH_API_V1.md`.
