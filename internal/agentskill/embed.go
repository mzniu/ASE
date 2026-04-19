package agentskill

import _ "embed"

// SKILLMD and ReferenceMD are embedded copies of the agent skill files served at GET /skills/ase-search-api/*
// (also mirrored under .cursor/skills/ase-search-api/ in the repo — update both when editing).
//
//go:embed SKILL.md
var SKILLMD []byte

//go:embed reference.md
var ReferenceMD []byte
