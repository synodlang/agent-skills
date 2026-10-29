# Changelog

All notable changes to the skills. Each released version's section is the
text of its GitHub release page.

## v0.3.0 — 2026-10-29

Implements Synod 0.3; follows dsl v0.3.1 and compiler v0.3.0.

### Added

- A "Themes" section in `SKILL.md` and `AGENTS.md`: what a theme is, how a main file selects one with `use`, the looks module `github.com/synodlang/looks` (looks `lumen` and `doodle`; themes `solarized/light`, `solarized/dark`, `terminal`, `sketch` and `chalk`), that any look combines with any theme, that a theme never changes a run, and where the table of theme values is.
- The references now cover `std/theme`, `std/theme/default`, the new `std/look` phrases and the errors `theme-outside-main`, `extra-theme` and `bad-theme`.

### Changed

- Re-stamped the hand-written files for dsl v0.3.1 and regenerated the references; the pins are `synod` v0.3.0 and examples v0.3.0.
- `tools/sync` writes the `synod 0.3` module for the code-block check, and the stale job looks for `v0.3.x` tags.
- The std catalog says `sim.StateRows` rows have at most 24 characters, and the `ringlook` example draws discs with a column of keys per owner.
- The vendored `paxos` example keeps its promise, accepted proposal and round on `std/disk`, so it stays correct after a crash and restart.

## v0.2.3 — 2026-10-28

Implements Synod 0.2; follows dsl v0.2.14 and compiler v0.2.7.

### Changed

- `SKILL.md` and `AGENTS.md` now point to the documentation at https://synodlang.github.io/docs/v0.2/: the 15-chapter tutorial, the Library pages with a runnable example for each std package, the error pages and the guide for agents.
- Re-stamped the hand-written files for dsl v0.2.14, which added the tutorial, the guides, the Library examples and decisions 103 to 111, and regenerated the references.
- The pinned `synod` is now v0.2.7, whose `synod explain --json` example files include `synod.mod`; the examples pin v0.2.1 is unchanged.

## v0.2.2 — 2026-10-23

Implements Synod 0.2; follows dsl v0.2.12.

### Changed

- The prompt bar of the Mac app now restricts Claude Code's built-in tools to Read, Glob and Grep, so `SKILL.md` and `AGENTS.md` tell an agent with the synod-editor tools that there is no shell, that `synod get` is the user's step, and that `Bash`, `Write` and `Edit` are not available.
- Re-stamped the hand-written files for dsl v0.2.12 and regenerated the references; the pins `synod` v0.2.6 and examples v0.2.1 are unchanged.

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
