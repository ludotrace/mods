# AGENTS.md — LudoTrace Mods Spec

Canonical instructions for this repo. `CLAUDE.md` points here.

## What this is

The **public capture-layer contract**. It defines what a LudoTrace mod must produce, and
ships the tooling to prove a mod complies. It contains no game code and no pipeline code.

| File | Role |
|------|------|
| `SPEC.md` | The contract — file naming, event format, reserved types, mod principles |
| `SCHEMA.md` | How to validate against the machine-readable schema |
| `event.schema.json` | The schema itself |
| `SCAFFOLD.md` | Guide for building a new game's mod from scratch |
| `validate.go` | `lt-validate`, the validator binary |
| `examples/` | `valid.jsonl` and `invalid.jsonl` fixtures |
| `VERSION` | Single source of truth for the release tag |

## The core promise

**A mod's output must be useful with zero LudoTrace involvement.** A player can read the
JSONL directly, or paste it into any LLM chat, without an account, an install, or an upload.
The pipeline is one consumer of this format, not its purpose.

Anything that would make the output only meaningful to LudoTrace's own backend does not
belong in the spec.

## Changing the contract

This is a published contract with independently-released implementations against it. Treat
every change as a compatibility decision:

- **Additive is safe** — new optional fields, new game-specific event types.
- **Reserved event types are a shared vocabulary.** Adding one commits every mod author to a
  meaning. Renaming or repurposing one breaks existing mods and stored data.
- **Never make a previously-optional field required** without a version bump.
- **Keep it game-agnostic.** If a rule only makes sense for one game's quirks, it belongs in
  that mod's own repo, not here. Game-specific event types are passed through by design —
  that is the extension point.
- `SPEC.md`, `event.schema.json`, and `examples/` must agree. A change to one is a change to
  all three.

## Validating

```bash
make build     # compile lt-validate with the version stamped in
make test      # valid fixture must pass, invalid must be rejected
```

`make test` is the real gate: it asserts both directions, so a schema loose enough to accept
`examples/invalid.jsonl` fails the build. When adding a rule, add a fixture case that would
fail without it.

## Releasing

`VERSION` is the single source of truth. `make release` guards on a clean tree and on being
on `main`, then tags from `VERSION` and pushes; CI builds and publishes from the tag.

Bump `VERSION` in its own commit before releasing. Never move an existing tag — downstream
mod authors pin to it.

## Scope

This repo is public and self-contained. Keep it that way:

- **Never reference a path, file, or issue number outside this repo.** No relative paths that
  climb out of the working tree, no pointers to private planning documents or infrastructure
  notes. If something here needs outside context, restate the part that matters in this
  repo's own words.
- **Never document the pipeline's internals.** How Core processes an events file, what it
  costs, or how accounts are tiered is not part of the capture contract and must not appear
  here.

## Docs

State what is true now. Don't narrate what a rule used to be or the incident that produced
it — git history and closed issues hold that.

## Issues & PRs

GitHub, single remote. **The repo is `ludotrace/mods`** even though this local folder is
named `mod-spec` — pass `--repo ludotrace/mods` when running outside a clone.

Issues and PRs both via `gh`. All changes go through a branch + PR, never directly to `main`.
Branch naming: `<type>/<short-description>`.

External contributors adding a mod for a new game open a PR against the supported-games table
in `README.md`.
