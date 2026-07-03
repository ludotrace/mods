# LudoTrace Mods

LudoTrace mods are the open capture layer. Each mod watches a game, writes a structured event log, and hands it off to the LudoTrace pipeline for processing.

**Mods are open source. The pipeline is not.** Anyone can build a LudoTrace mod for any game.

---

## Supported games

| Game | Status | Mod repo | Nexus |
|------|--------|----------|-------|
| Fallout 4 | Active | [ludotrace/fallout4](https://github.com/ludotrace/fallout4) | [Nexus Mods](https://www.nexusmods.com/fallout4/mods/106076) |
| Stardew Valley | Active | [ludotrace/stardew](https://github.com/ludotrace/stardew) | [Nexus Mods](https://www.nexusmods.com/stardewvalley/mods/48026) |

---

## How a mod works

A LudoTrace mod does one thing: append JSON events to a single file as the game runs.

```
lt_fo4_events.jsonl      ← Fallout 4
lt_stardew_events.jsonl  ← Stardew Valley (example)
```

Each line is a self-contained JSON object:

```json
{"type": "session_start", "game_time": "00:00", "level": 12, ...}
{"type": "location", "name": "Diamond City", "game_time": "01:14"}
{"type": "kill", "target": "Raider", "game_time": "01:22"}
{"type": "session_end", "game_time": "02:45", "level": 13, ...}
```

The LudoTrace desktop client watches the file, extracts complete sessions, and uploads them. The mod never interacts with the client directly.

---

## Build a mod

Read [`SPEC.md`](./SPEC.md) — it defines the full contract: file naming, event format, reserved types, and mod principles.

Then validate your mod's output against [`event.schema.json`](./event.schema.json) to prove it's compliant — see [`SCHEMA.md`](./SCHEMA.md).

The implementation language and engine integration are entirely up to you. Existing mods use:
- **Papyrus** (Fallout 4, Skyrim, Starfield)
- **SMAPI** (Stardew Valley)
- **ScriptHookV** (GTA V)
- Anything else that can write to a file

If you build a mod, open an issue or PR to add it to the table above.

---

## Resources

- [SPEC.md](./SPEC.md) — full schema contract
- [SCHEMA.md](./SCHEMA.md) — machine-readable `event.schema.json` + validator
- [ludotrace/fallout4](https://github.com/ludotrace/fallout4) — reference implementation (Papyrus)
- [ludotrace.github.io](https://ludotrace.github.io) — project site

## Found a bug, or want to contribute?

Open an issue or a pull request on [github.com/ludotrace/mods](https://github.com/ludotrace/mods). If you're adding a mod for a new game, open a PR against the table above once it's working.
