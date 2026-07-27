# LudoTrace Mod Scaffolding Guide

Use this document to scaffold a new LudoTrace capture mod for any game.
Fill in the **Inputs** block, then hand this document (plus `SPEC.md`) to Claude Code.

The mod's output must satisfy the machine-readable contract in `event.schema.json` — validate
it with `lt-validate` (see `SCHEMA.md`). Every scaffold ships a compliance-CI workflow that
runs this check on each push.

---

## How to invoke

> "Use `mods/SCAFFOLD.md` to scaffold a mod for **[GAME_NAME]**."

Optionally add freeform notes about the game's modding situation. Claude will work through
the analysis phases below and produce the full artifact set.

---

## Inputs

```
GAME_NAME:    <full game title, e.g. "The Witcher 3: Wild Hunt">
GAME_ID:      <lowercase snake_case id, e.g. "witcher3">  ← used in filenames and Core registry
ENGINE:       <engine / platform, e.g. "REDengine 3">
PLATFORM:     <PC / console / mobile>
```

`GAME_ID` must match the id pre-registered in Core's game registry and the user's lt-client
config. The events file will be named `lt_<GAME_ID>_events.jsonl`.

---

## Client integration contract

These are hard requirements imposed by lt-client. They constrain all Phase 1 decisions
and must never be violated by the mod.

**File contract:**
- Write to a single, fixed-path, append-only file: `lt_<GAME_ID>_events.jsonl`
- Never truncate, rotate, or delete the file — lt-client owns the file after it reads it
- Never recreate the file from scratch — lt-client watches for WRITE events, not CREATE;
  a new file would reset the client's sidecar offset and cause duplicate uploads
- One complete JSON object per line; no partial flushes mid-line

**Session boundary contract:**
- `session_start` is the **only** structural boundary lt-client uses to split sessions.
  A new `session_start` closes whatever session was open and starts a new one
- `session_end` is **opaque payload** — lt-client buffers it into the current session
  without treating it as a terminator. A single play session (one `session_start`) may
  contain many `session_end` lines (e.g. Fallout 4 writes one per save). This is
  intentional and game-agnostic; never design around `session_end` as a close signal
- Sessions without `session_end` (crash, force-quit, or a post-session parser that runs
  before a clean exit) are flushed by lt-client's 12-minute inactivity window — the mod
  does not need to handle this

**Game identity contract:**
- lt-client derives `game_id` from the user's `config.toml`, not from filename or JSONL
  content. The user's config entry must match Core's registered game_id exactly:
  ```toml
  [[games]]
  game_id     = "witcher3"              ← must match Core registry
  watch_path  = "C:/path/to/game/dir"
  events_file = "lt_witcher3_events.jsonl"
  ```
- Core must have `GAME_ID` registered before lt-client can upload sessions for it.
  Flag this as a backend prerequisite in CLAUDE.md.

---

## Phase 1 — Research the modding surface

Answer each question before writing any code. If the answer is unknown, note the
best available option and flag it as an assumption.

### 1.1 Implementation approach

Pick exactly one:

| Approach | Use when |
|----------|----------|
| **In-process scripting** | Game has a supported scripting API that can write files (Papyrus, SMAPI/C#, Lua, etc.) |
| **Post-session parser** | Game writes replay or log files after matches; no in-process file write API |
| **External companion** | No scripting API, no replay file; requires log-tail, overlay, or external observer |

> Fallout 4: in-process (Papyrus via Hydra).
> Stardew Valley: in-process (SMAPI C#).
> AoE2:DE: post-session parser (`.aoe2record` via `mgz`).

```
APPROACH:     <in-process | post-session | external>
MOD_LOADER:   <SMAPI 4.x | Hydra | mgz + watchdog | ScriptMerger | none | ...>
LANGUAGE:     <C# | Papyrus | Python | Lua | ...>
FILE_WRITE:   <how the mod appends to disk, e.g. "File.AppendAllText" | "Hydra:IO:File.AppendLine">
```

### 1.2 Timestamp strategy

The SPEC defines three optional timestamp fields. Use whichever the game makes available.

| Field | Use when | Format |
|-------|----------|--------|
| `game_time` | Game has an in-world clock | `"HH:MM"` |
| `game_date` | Game has an in-world calendar | ISO 8601 if mappable; game-native string otherwise (document the deviation) |
| `wall_time` | No in-world clock, or supplement game clock with real-world time | ISO 8601 UTC, e.g. `"2025-06-21T14:00:00Z"` |

> Fallout 4: `game_date` (`"2288-02-27"`) + `game_time` (`"14:32"`).
> Stardew Valley: `game_date` (`"Spring 15 Y1"` — non-ISO, documented deviation) + `game_time`.
> AoE2:DE: `wall_time` on session boundaries; custom `match_time` field (`"MM:SS"`) for in-match elapsed.

```
TIMESTAMPS:   <which fields, formats, and any non-ISO deviations>
```

### 1.3 Session boundaries

```
SESSION_START_TRIGGER:  <what fires when a play session begins, e.g. OnPostLoadGame, SaveLoaded, match start>
SESSION_END_TRIGGER:    <what fires when a session ends, e.g. OnPostSaveGame, Saving, match end>
```

If `SESSION_END_TRIGGER` fires multiple times per play session (e.g. each save in Fallout 4),
document that — it affects what the `session_end` snapshot should include and what Core/LLM
sees. Each `session_end` line is valid; lt-client passes them all through as payload.

### 1.4 Event mapping

For each SPEC reserved type, decide: **applicable**, **not applicable**, or
**implemented via a non-obvious hook** (explain the hook).

| Reserved type | Applicable? | Hook / trigger |
|---------------|-------------|----------------|
| `session_start` | always | see 1.3 |
| `session_end` | always | see 1.3 |
| `location` | if named locations exist | ? |
| `kill` | if combat exists | ? |
| `quest_stage` | if quests exist | ? |
| `near_collectible` | if missable items exist | ? |
| `stat` | if numeric progression exists | ? |
| `used` | if consumable items exist | ? |

Then list **game-specific event types** that add LLM coaching value beyond the reserved set.
For each: name, trigger, payload fields, coaching rationale.

---

## Phase 2 — Decisions log

Before generating files, write a short decisions log. This gets summarised into CLAUDE.md.

```
APPROACH:       <chosen approach and why>
TIMESTAMP:      <which fields and why the others were skipped>
SESSION_START:  <trigger>
SESSION_END:    <trigger and whether it fires multiple times per play session>
RESERVED_USED:  <which reserved types apply>
CUSTOM_EVENTS:  <list of game-specific types>
DEVIATIONS:     <any departures from SPEC defaults and why>
PREREQS:        <backend tasks — Core game_id registration, etc.>
```

---

## Phase 3 — Generate artifacts

Produce exactly these files. No more, no less.

### 3.1 `<game_id>/CLAUDE.md`

Living context document. Follow this structure exactly:

```markdown
# <GAME_NAME> LudoTrace Mod

## What it does
One paragraph. Dumb emitter. Append-only. JSONL to `lt_<game_id>_events.jsonl`.

## Repo structure
<directory tree with one-line descriptions>

## Tech
- Approach: <in-process | post-session | external>
- Mod loader / API: <tool and version>
- Install: <exact install steps, file-drop preferred>

## Events file path
<exact path, using variable notation if machine-specific, e.g. `%GAME%\lt_witcher3_events.jsonl`>

## lt-client config entry
<exact toml block the user pastes into their lt-client config>

## Timestamp format
<which fields are used and their exact format, including any non-ISO deviations>

## Event types

### Reserved (from SPEC)
| Type | Hook / trigger |
|------|----------------|
...

### Game-specific
| Type | Fields | Hook / trigger |
|------|--------|----------------|
...

## session_start / session_end snapshot
<example JSON — show all fields the snapshot includes>

## Implementation notes
<non-obvious decisions, Harmony patches, deduplication, rate-limiting, quirks>

## Build
<exact commands — make build, dotnet build, etc.>

## Release
make release  — tags HEAD, pushes tag, GitHub Actions builds zip and publishes

## Dependencies
<list with pinned versions>

## Backend prerequisites
- [ ] Core game registry: `<GAME_ID>` registered before first upload
```

### 3.2 `<game_id>/src/<main_file>`

Skeleton implementation. Must include:

- [ ] Events file path constant (`lt_<game_id>_events.jsonl`)
- [ ] Append-only write function with try/catch — never crashes game or host process
- [ ] `session_start` written on the session-start trigger with full snapshot
- [ ] `session_end` written on the session-end trigger with full snapshot
- [ ] At least one mid-session event handler (location or kill, whichever applies first)
- [ ] All timestamp fields populated per Phase 2 decisions
- [ ] All field names match `SPEC.md` exactly (no aliases, no camelCase variants)
- [ ] Version injected from `VERSION` file at build time (pattern: `__VERSION__` placeholder
  substituted by the build step, as in FO4's `make build`)

### 3.3 `<game_id>/<project_file>`

| Language | Project file |
|----------|-------------|
| C# (SMAPI) | `<ModName>.csproj` + `manifest.json` |
| Papyrus | No project file; build documented in CLAUDE.md and Makefile |
| Python | `pyproject.toml` |
| Lua | mod descriptor JSON (game-specific format) |
| JavaScript | `package.json` |

Minimal — only declared dependencies and target runtime.

For SMAPI mods, `manifest.json` must include `"Version"` that matches `VERSION`.

### 3.4 `<game_id>/VERSION`

Single line: `v0.1.0`

This is the source of truth for the version. Makefile and CI read it; it is injected
into source at build time and into the manifest. Never hardcode the version anywhere else.

### 3.5 `<game_id>/Makefile`

Three targets, always:

```makefile
build:    # compile and produce dist/ or equivalent build output
          # inject VERSION into source (sed __VERSION__ or equivalent)
          # restore source after compile (never leave injected version in source)

run:      # launch the game with the mod installed, for local testing
          # may be a no-op stub if not automatable (e.g. post-session parsers)

release:
          # guard: fail if uncommitted changes exist
          # guard: fail if not on main branch
          # read VERSION, git tag, git push origin <tag>
          # GitHub Actions handles the rest
```

Adapt language and toolchain (WSL paths, cmd.exe delegation for Windows-native compilers, etc.).
The FO4 Makefile is the reference pattern for in-process mods with Windows-native build tools.

### 3.6 `<game_id>/docs/coaching_prompt.md`

Standalone Claude.ai prompt for users without lt-client. Required in every mod.
Structure:

```markdown
# <GAME_NAME> LudoTrace — Claude.ai Prompt

Copy and paste this prompt into Claude.ai, then paste your `lt_<game_id>_events.jsonl`
contents below it.

---

You are a <GAME_NAME> playthrough coach. I'm going to paste a session log generated
by the LudoTrace mod.

<2-3 sentences describing the player's context — what kind of game it is, what
decisions the player makes, what coaching looks like for this game.>

Based on the log, give me:
1. <game-appropriate coaching question 1>
2. <game-appropriate coaching question 2>
3. <game-appropriate coaching question 3>

Be specific — reference actual <items / units / quests / stats> from the log.
<any negative constraints — "don't spoil endings", "don't push difficulty changes", etc.>
```

### 3.7 `<game_id>/README.md`

Community-facing. Written for players on NexusMods, Steam Workshop, CurseForge, or
wherever this game's mod community lives. Structure:

```markdown
# LudoTrace for <GAME_NAME>

<One-sentence description — what it does and why.>

## Requirements
<mod loader and dependencies, with links>

## Install
<file-drop or installer steps — as few lines as possible>

## Usage
1. Play. The mod records automatically. Log is written to:
   `<events file path>`
2. Get coaching. <brief description of Claude.ai standalone flow or lt-client flow>

## Snapshot format
<3–5 example JSONL lines showing the event variety — use realistic-looking values,
not placeholders. This is what sells the mod to community members.>

## Notes
<any gotchas, known limitations, or important caveats>
```

### 3.8 `<game_id>/.github/workflows/release.yml`

CI/CD pipeline. Triggered by `git push origin v*` (from `make release`).

```yaml
name: Release

on:
  push:
    tags: ['v*']
  workflow_dispatch:
    inputs:
      version:
        description: 'Version tag (e.g. v0.2.0)'
        required: true

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    env:
      VERSION: ${{ github.event.inputs.version || github.ref_name }}
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Build release zip
        run: |
          # <game-specific build commands>
          # zip contents: whatever goes in the community release package
          # always include docs/coaching_prompt.md in the zip

      - name: Create GitHub Release
        id: gh_release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ env.VERSION }}
          files: <ZipFilename>-${{ env.VERSION }}.zip
          generate_release_notes: true

      - name: Update CHANGELOG.md
        run: |
          VERSION="${{ env.VERSION }}"
          DATE=$(date -u +%Y-%m-%d)
          NOTES="${{ steps.gh_release.outputs.body }}"
          { printf '## %s — %s\n\n' "$VERSION" "$DATE"; echo "$NOTES"; echo ""; cat CHANGELOG.md 2>/dev/null || true; } > CHANGELOG.md.tmp
          mv CHANGELOG.md.tmp CHANGELOG.md
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add CHANGELOG.md
          git commit -m "Update CHANGELOG.md for $VERSION [skip ci]"
          git push origin HEAD:main

      # Include the community upload step that matches this game:
      - name: Upload to <community>
        # see Publishing targets below
```

### 3.9 `<game_id>/CHANGELOG.md`

Empty initial file with the scaffold version:

```markdown
## v0.1.0 — <date>

Initial release.
```

---

### 3.10 `<game_id>/examples/sample.jsonl`

A short, realistic sample of the mod's output — enough lines to cover `session_start`, the
main mid-session events, any game-specific types, and `session_end`. This is both a fixture
for compliance CI (3.11) and a reference for contributors. Use realistic values, not
placeholders.

### 3.11 `<game_id>/.github/workflows/spec-compliance.yml`

Validates `examples/sample.jsonl` against the LudoTrace event schema on every push. The
validator (`lt-validate`) embeds the schema, so this repo carries no copy to drift. See
[SCHEMA.md](./SCHEMA.md) for the full contract.

```yaml
name: spec-compliance
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: fetch validator
        run: |
          curl -sSL -o lt-validate \
            https://github.com/ludotrace/mods/releases/latest/download/lt-validate-linux-amd64
          chmod +x lt-validate
      - run: ./lt-validate examples/sample.jsonl
```

---

## Publishing targets

Every mod publishes to two places: **GitHub** (always) and the **game's mod community**
(game-dependent). The community upload step replaces the placeholder in `release.yml`.

| Game / platform | Community | CI step |
|-----------------|-----------|---------|
| Fallout 4 | NexusMods | `Nexus-Mods/upload-action@v1.0.0-beta.8` |
| Stardew Valley | NexusMods | `Nexus-Mods/upload-action@v1.0.0-beta.8` |
| Stardew Valley | CurseForge | `itsmeow/curseforge-upload@v3` (secondary) |
| AoE2:DE | GitHub only | no mod community for standalone tools |
| <new game> | ? | research: NexusMods, Steam Workshop, game-specific portal |

**NexusMods secrets and vars** (set per-repo in GitHub):

| Key | Kind | Value |
|-----|------|-------|
| `NEXUSMODS_API_KEY` | Secret | API key from nexusmods.com/users/myaccount |
| `NEXUSMODS_GROUP_ID` | Var | `file_id` within the mod page — the ID of the file created during the first manual upload. Find it via the NexusMods API: `curl -H "apikey: KEY" "https://api.nexusmods.com/v1/games/<game>/mods/<id>/files.json"` |

`upload-action v1.0.0-beta.8` uses `file_id` (not `mod_id` or `game_domain_name` — those are invalid for this version).
The action adds new versions to the file identified by `file_id` and archives the previous version when `archive_existing_version: true`.

NexusMods requires the first upload to be done manually to create the mod page and get the IDs.
After that, `upload-action` handles subsequent versions.

**Org publishing:**
All mods live under the `ludotrace` GitHub org (`github.com/ludotrace/<game_id>`).
Each repo is public. The README is the public face; the CLAUDE.md is the internal face.

---

## Constraints (non-negotiable)

From `SPEC.md` and the client integration contract above:

1. **No identity.** No user ID, Steam ID, account info, or credentials in the events file.
   In-game names (`player.name`, `farm_name`) are fiction and allowed.

2. **Append-only.** Never truncate, rotate, or delete the events file.

3. **Never recreate the file.** Always open in append mode. A new file resets lt-client's
   sidecar offset and causes duplicate uploads.

4. **One event per line.** Single complete JSON object, no embedded newlines, no blank lines.

5. **No opinions.** Record events impartially. No filtering by "importance". Core decides.

6. **Dumb emitter.** No session management, no upload logic, no lt-client awareness.

7. **Field names from SPEC exactly.** `type`, `game_time`, `game_date`, `wall_time`,
   `kill.target`, `kill.killer`, `stat.stat`, `stat.value`, `location.name`, etc.
   `lt-validate` (see [SCHEMA.md](./SCHEMA.md)) checks this mechanically for reserved types —
   run it on sample output before considering the scaffold done.

8. **`session_start` is the only structural boundary.** The mod must write one at the
   start of every play session. Crashes and force-quits are handled by lt-client's
   12-minute inactivity window — no mod-side fallback needed.

9. **VERSION is the single version source of truth.** Never hardcode version strings
   in source, manifests, or CI independently.

---

## Quality bar

The scaffold is complete when:

- [ ] `CLAUDE.md` answers every question a new contributor would have without reading code
- [ ] `CLAUDE.md` includes the exact lt-client `config.toml` entry the user needs
- [ ] `CLAUDE.md` lists Core game registry registration as a backend prerequisite
- [ ] The skeleton compiles / runs without errors
- [ ] `make build`, `make run`, `make release` are all defined (stubs allowed for run/release)
- [ ] `VERSION` file exists and its value appears in the build output
- [ ] `docs/coaching_prompt.md` is present and game-appropriate
- [ ] `README.md` is present and community-facing (a player can install from it alone)
- [ ] `.github/workflows/release.yml` is present with the correct community upload step
- [ ] `CHANGELOG.md` has an initial entry
- [ ] A sample play produces SPEC-compliant JSONL — validated with `lt-validate
  lt_<game_id>_events.jsonl`, not just `jq` (jq only checks JSON well-formedness; `lt-validate`
  checks the reserved-type field contract). See [SCHEMA.md](./SCHEMA.md).
- [ ] First line of any session is `session_start`; no structural reliance on `session_end`
