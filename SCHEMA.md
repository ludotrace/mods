# Event Schema

[`event.schema.json`](./event.schema.json) is the machine-readable form of the
[SPEC](./SPEC.md) — a [JSON Schema](https://json-schema.org) (draft 2020-12) you can
validate a mod's output against to prove it's compliant.

[`validate.go`](./validate.go) builds `lt-validate`, a **self-contained binary** that runs
the check. The schema is embedded in the binary, so validating a mod's output is *download
one file and run it* — no runtime, no package manager, nothing to install. That matters:
mod authors write Papyrus, C#, or Lua, and shouldn't need a JavaScript or Python toolchain
just to check a log file.

## What it checks

The schema describes **one line** of an events file. A LudoTrace events file is
[JSONL](https://jsonlines.org) — one JSON object per line — so validation means splitting
the file on newlines and validating each non-blank line. `lt-validate` does exactly that.

It enforces the mechanical parts of the SPEC's *Line Format* and *Reserved Event Types*:

- every line is a JSON object with a required string `type`;
- reserved types carry their required fields, correctly typed
  (`location`→`name`, `kill`→`target`, `quest_stage`→`quest`+`stage`,
  `near_collectible`→`name`, `stat`→`stat`+`value`, `used`→`item`);
- `game_date` / `wall_time`, when present, are valid ISO 8601 date / date-time.

By design it stays **loose** where the SPEC is loose:

- **Game-specific types pass** — `type` is any string; unknown types only need the base shape.
- **Extra fields pass through** — additional fields are handed to the LLM as-is, so the
  schema never rejects unknown properties.
- **`session_start` / `session_end` have no required fields** — the SPEC *recommends* a state
  snapshot but does not require one, so the schema doesn't either.
- **`stat.value` is `number`** — matches the SPEC example. If a game needs a non-numeric stat
  value, that's a SPEC change first, then a schema change.

## What it can't check

A per-line schema can't see behavior or lifecycle. These still live in lt-client and review,
not here:

- append-only writing (no truncation / rotation),
- `session_start` must be the first event in a session,
- no user identity or credentials in the file,
- correct file name and location (`lt_<game_id>_events.jsonl` in the game's data dir).

## Validate your mod's output

**Download the binary** for your platform from the [latest release](https://github.com/ludotrace/mods/releases/latest)
(`lt-validate-linux-amd64`, `lt-validate-macos-arm64`, `lt-validate-windows-amd64.exe`, …), then:

```
./lt-validate path/to/lt_<game_id>_events.jsonl
```

Exit code `0` means every line is compliant; `1` prints each violation as
`file:line: what's wrong`.

**Or, if you have Go**, run it straight from source with no download:

```
go run github.com/ludotrace/mods@latest path/to/events.jsonl
```

## CI for your mod repo

Drop a sample events file in your repo (e.g. `examples/sample.jsonl`) and validate it against
this contract on every push. Because the schema is embedded in the binary, your mod repo
carries no copy of the schema to drift:

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

## Worked examples / test suite

`examples/valid.jsonl` and `examples/invalid.jsonl` double as the CI test (see
`.github/workflows/validate.yml`) and as worked examples: `valid.jsonl` shows every reserved
type plus a game-specific one; `invalid.jsonl` shows one of each failure mode the validator
catches.

Contributors can run the whole thing locally with just a Go toolchain:

```
make test          # build with version stamped in, then run both fixtures
./lt-validate --version
```

## Releasing the validator

`VERSION` is the single source of truth for the tool's version (same convention as the mods).
To cut a release:

1. Bump `VERSION` (e.g. `v0.2.0`) on `main` and merge it.
2. `make release` — guards a clean tree on `main`, tags from `VERSION`, and pushes the tag.
3. `.github/workflows/release.yml` fires on the `v*` tag: it checks the tag matches `VERSION`,
   runs `make test`, cross-compiles `lt-validate` for linux/macos/windows with the version
   stamped in (`-ldflags -X main.version=…`), and attaches the binaries to a GitHub Release.

The `spec-compliance.yml` a mod repo runs always pulls `releases/latest`, so mods pick up new
validator versions automatically.
