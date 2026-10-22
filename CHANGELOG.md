# Changelog

All notable changes to the skills. Each released version's section is the
text of its GitHub release page.

## Unreleased

## v0.2.1 — 2026-10-22

Implements Synod 0.2; follows dsl v0.2.11.

### Changed

- Re-stamped the hand-written files (`SKILL.md`, `AGENTS.md`, `grammar.md`, `phrases.md`, `contexts.md`) for dsl v0.2.11, which changed only the implementation documents and decision 94 about `synod new` and `synod get`; no rule of the skill changed.
- Regenerated the references from dsl v0.2.11; the pins `synod` v0.2.6 and examples v0.2.1 are unchanged.

## v0.2.0 — 2026-10-22

Implements Synod 0.2. The first release of the skills: a Claude Code plugin that teaches an agent to look, write, format, check and run Synod, with the same entry point as `AGENTS.md` for other agents.

### Added

- The `synod` skill (`skills/synod/SKILL.md`): the workflow, the eight rules, the reuse rules, the budget rules and the use of the `synod-editor` tools when they are present.
- Hand-written references condensed from the specification: `grammar.md`, `phrases.md` and `contexts.md`.
- Generated references, pinned to dsl v0.2.10, `synod` v0.2.6 and examples v0.2.1: `canonical.md`, `trace-events.md`, `std-catalog.md`, `errors.md` and `examples/`.
- `AGENTS.md`, generated from the skill for agents that are not Claude Code.
- The Claude Code plugin and marketplace files (`synod@synodlang`).
- `tools/sync`, which regenerates the generated files and, with `--check`, fails on any difference, on a stale `condensed from dsl` stamp and on a `plugin.json` version that is not the newest changelog version.
- CI: the text check, the sync job (`fmt --check` and `check` on the examples and on every whole-file `synod` block of the skill) and the stale job (a newer dsl, compiler or examples release).
- A release workflow that attaches `agent-skills-<tag>.tar.gz` and `SHA256SUMS` and uses this changelog as the release text.
