# LudoTrace Mod Spec

The contract any LudoTrace mod must satisfy to be compatible with lt-client and Core.
Implementation language and engine integration are entirely up to the mod author.

---

## Core Contract

A mod writes JSON events — one complete JSON object per line — to a single append-only file
in the game's data directory. That is the full obligation. lt-client reads the file; the mod
never interacts with lt-client directly.

---

## Events File

```
lt_<game_id>_events.jsonl
```

Examples:
```
lt_fo4_events.jsonl
lt_stardew_events.jsonl
lt_rdr2_events.jsonl
```

- One file per game installation. Written to the game's local data or save directory.
- Append-only. The mod never truncates, rotates, or deletes it. lt-client owns the file
  lifecycle after reading.
- `game_id` must match the id registered in Core's game registry and the user's lt-client
  config. lt-client derives the game from its config, not from the filename or file contents.

---

## Line Format

Each line is a self-contained JSON object. No multi-line values. No blank lines.

**Required fields on every event:**

```json
{"type": "event_type"}
```

- `type` — string, identifies the event kind (see Reserved Types below)

**Timestamp fields — all optional:**

- `game_time` — in-world clock time in the game's native format (e.g. `"14:32"` for HH:MM).
  Include when the game has an in-world clock.
- `game_date` — in-world calendar date in ISO 8601 format (e.g. `"2288-02-28"`).
  Include when the game has an in-world calendar. May appear alongside `game_time`.
- `wall_time` — real-world UTC timestamp in ISO 8601 (e.g. `"2024-11-15T14:32:00Z"`).
  Include when the game has no in-world clock, or alongside `game_time` if both are
  available.

Event ordering is preserved by line position in the append-only file. Timestamps are
context for the LLM — they make insights more grounded — but are not structurally required.
Include whichever the game makes available; omit both if neither is accessible.

Additional fields are event-specific and passed through to the LLM prompt as-is.

---

## Reserved Event Types

These types are defined across all games. Mods must use them for the described purposes and
must not use them for other purposes.

### `session_start`

Written when a play session begins (on game load, new game, or equivalent).

```json
{"type": "session_start", "game_time": "00:00", ...character/world snapshot fields}
```

- Must be the first event in every session.
- Should include a snapshot of the player's current state (level, stats, inventory, etc.)
  so Core has baseline context for the session.
- `session_start` is the **only** structural boundary lt-client uses to split sessions — a new `session_start` closes whatever session was open and starts a new one.

### `session_end`

Written when a play session ends (on save, quit, or equivalent).

```json
{"type": "session_end", "game_time": "HH:MM", ...character/world snapshot fields}
```

- Should mirror `session_start` fields so Core can diff start vs. end state.
- `session_start` is the **only** structural boundary lt-client uses to split sessions.
  `session_end` is **opaque payload** — lt-client buffers it into the current session
  without treating it as a terminator. A single play session (one `session_start`) may
  contain many `session_end` lines (e.g. Fallout 4 writes one per save); this is
  intentional and game-agnostic — never design around `session_end` as a close signal.
- A session without a new `session_start` to close it (crash, force-quit) is uploaded as
  an orphan after a configured inactivity window.

### `location`

Player entered a named location.

```json
{"type": "location", "name": "Location Name", "game_time": "14:32"}
```

### `kill`

A kill event involving the player.

```json
{"type": "kill", "target": "Enemy Name", "killer": "", "game_time": "14:45"}
```

### `quest_stage`

Quest progression event.

```json
{"type": "quest_stage", "quest": "Quest Name", "stage": 10, "game_time": "22:10"}
```

### `near_collectible`

Player is near a missable or notable collectible.

```json
{"type": "near_collectible", "name": "Item Name", "game_time": "30:00"}
```

### `stat`

Periodic stat snapshot or stat change.

```json
{"type": "stat", "stat": "Stat Name", "value": 42, "game_time": "15:00"}
```

### `used`

Player used a consumable or item.

```json
{"type": "used", "item": "Item Name", "game_time": "18:22"}
```

---

## Game-Specific Event Types

Mods may define additional event types beyond the reserved list. They are passed through
to the LLM prompt template as-is. Use a consistent naming convention; avoid collisions with
reserved types.

---

## Mod Principles

- **Dumb emitter.** The mod records what happens impartially. No session management, no
  file lifecycle decisions, no awareness of lt-client.
- **Append-only.** Never overwrite or truncate the events file. Only append.
- **One event per line.** Each JSON object must be a single complete line with no embedded
  newlines.
- **No identity.** The events file contains no user ID, account info, or credentials.
  User identity is supplied by lt-client at upload time, out-of-band.
- **No opinions.** Record events; do not filter or editorialize. Core and the LLM draw
  conclusions.
- **File drop install.** Where the game engine supports it, mods should require no
  configuration and no user setup beyond placing files in the right directory.
