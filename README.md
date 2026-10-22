# agent-skills

Skills that teach coding agents, such as Claude Code, to read, write, check and run Synod.

[![ci](https://github.com/synodlang/agent-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/synodlang/agent-skills/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

Install it as a Claude Code plugin, or point any other agent at [AGENTS.md](AGENTS.md). Both teach the agent to look for an existing package, write Synod with the compiler's own checks, and confirm the run shows what was asked. The plan is in [08-agent-skills.md](https://github.com/synodlang/dsl/blob/main/implementation/08-agent-skills.md). You need the `synod` command ([install page](https://synodlang.github.io/install/)); the skills pin `synod` v0.2.6.

## Install

### Claude Code plugin

```
claude plugin marketplace add synodlang/agent-skills
claude plugin install synod@synodlang
```

Inside a Claude Code session the same two steps are `/plugin marketplace add synodlang/agent-skills` and `/plugin install synod@synodlang`. The `synod` skill then loads when a task mentions Synod, `.synod` files or `synod.mod`. The command spellings follow the Claude Code documentation (code.claude.com/docs/en/discover-plugins); they have not yet been run against a Claude Code install from this repository, so treat them as to-verify until the first release is installed.

### Other agents

Copy or point your agent at [AGENTS.md](AGENTS.md). It is the same entry point as the skill, generated from it, and it names the reference files under `skills/synod/reference/`. Keep that folder next to it.

### The `synod` tools over MCP

An agent without a shell, or any Claude Code session, can use the compiler's commands as MCP tools (`doc`, `explain`, `check`, `run_summary`, `fmt`):

```
claude mcp add synod -- synod mcp
```

`synod mcp` needs `synod` v0.2.6 or later on your `PATH`. The Mac app's editor adds its own `synod-editor` tools; the skill explains how to use them when they are present.

## Where it fits

An agent follows these skills to write Synod, and uses the `synod` command to check its work. The [Mac app](https://github.com/synodlang/mac-app) will load the main skill into the Claude Code CLI when you ask it to animate something. Any other agent that can follow the same files works the same way.

| Repository | What it is | Status |
| --- | --- | --- |
| [dsl](https://github.com/synodlang/dsl) | The language specification and implementation guides | Spec v0.2 |
| [compiler](https://github.com/synodlang/compiler) | The `synod` command: `fmt`, `check`, `run`, `doc`, `explain`, `new`, `get`, `mcp` | v0.2.6 released |
| [std](https://github.com/synodlang/std) | The standard library packages | In use by the compiler |
| [examples](https://github.com/synodlang/examples) | Sample programs and acceptance tests | Available |
| [player](https://github.com/synodlang/player) | Canvas library that plays `scene.json` | Planned (M7) |
| [mac-app](https://github.com/synodlang/mac-app) | The Mac app | Planned (M8, M9, M10) |
| **agent-skills** (this repository) | Skills for coding agents | v0.2.0 |
| [synodlang.github.io](https://github.com/synodlang/synodlang.github.io) | The website and documentation | Install page live; rest planned (M11) |
| [homebrew-tap](https://github.com/synodlang/homebrew-tap) | Homebrew formula for `synod` | Updated by compiler releases |
| [scoop-bucket](https://github.com/synodlang/scoop-bucket) | Scoop manifest for `synod` | Updated by compiler releases |
| [.github](https://github.com/synodlang/.github) | Organization profile and shared workflows | Available |

## What the skills teach

The entry point is a short `SKILL.md`. In outline it tells the agent to:

1. Look before writing: run `synod doc` and search the workspace and the examples for an existing package, and write only a main file when one fits.
2. Treat every argument as one item: parenthesize every nested phrase and every multi-item condition.
3. Qualify imported names (`net.Send`) and never qualify the package's own names.
4. Name the node with a receiver in node code, and touch only that node's state.
5. Write explicit types in holes.
6. Leave looks optional, since the basic look animates everything.
7. Run `synod fmt`, then `synod check` until clean, then `synod run`, and confirm the trace shows what was asked. Never hand back unchecked code.

## Layout

```
.claude-plugin/plugin.json       the plugin
.claude-plugin/marketplace.json  the marketplace "synodlang"
skills/synod/SKILL.md            the entry point (Claude Code skill format)
skills/synod/reference/
    grammar.md                   the language structure and reserved words
    phrases.md                   pattern rules, resolution, results, blocks
    contexts.md                  the four contexts and what each may do
    canonical.md                 generated: the canonical form
    trace-events.md              generated: the trace event table
    std-catalog.md               generated by `synod doc --catalog`
    errors.md                    generated by `synod explain --catalog`
    examples/                    generated: a copy of the examples repository
AGENTS.md                        generated: the same entry point for other agents
versions.env                     the pinned dsl, synod and examples releases
tools/sync/                      regenerates the generated files
```

`grammar.md`, `phrases.md` and `contexts.md` are written by hand, each stamped with the `dsl` tag it was condensed from. Everything marked generated comes from the releases pinned in `versions.env`: `dsl` v0.2.10, `synod` v0.2.6 and examples v0.2.1. The copy in `reference/examples/` is under the examples repository's [CC BY 4.0](https://github.com/synodlang/examples/blob/main/LICENSE) licence.

## Using synod from an agent today

You do not need to wait for the skills. Install `synod` from https://synodlang.github.io/install/ and use these forms.

- `synod fmt` rewrites files to canonical form. The compiler accepts only canonical code.
- `synod check` checks the whole program. When its output is piped, it prints one line per diagnostic, with a hint, and the same text is unchanged on every run. This is the form agents should read.
- `synod check --json` prints `synod-check/1` JSON with exact spans (`line`, `col`, `end`) and fixes (`help` with `edits`). Use it when you want to apply a fix. See [04-formats.md](https://github.com/synodlang/dsl/blob/main/implementation/04-formats.md).
- `synod check --style rich` shows the source line with the problem underlined, as a person sees it on a terminal.
- `synod run <folder>` simulates the program and writes `out/trace.jsonl`. The same program and seed give a byte-identical trace. `--seed n` changes the seed.
- `synod run --summary` prints a compact text summary of the run instead of the status line: the steps that recorded facts, the invariant results and the final state of every node. Read this, not the whole trace.

Never paste whole traces into a prompt, and never invent coordinates, colors or timings: looks decide those.

## Contributing

Changes to the language start in [dsl](https://github.com/synodlang/dsl); this repository follows its releases, each with its own patch number. Do not edit generated files. To regenerate them, check out `dsl` and `examples` at the pins, put the pinned `synod` on your `PATH`, and run:

```
go run ./tools/sync -dsl ../dsl -examples ../examples
go run ./tools/sync -check -dsl ../dsl -examples ../examples
```

CI has three jobs. **text** checks line endings, trailing spaces and the final newline. **sync** fails on any difference after regeneration, on a hand-written file stamped with another `dsl` tag, on a `plugin.json` version that is not the newest changelog version, and when `synod fmt --check` or `synod check` fails on the examples or on a whole-file `synod` block of the skill. **stale** fails (on pushes, pull requests and daily) when a newer `v0.2.x` of `dsl`, `compiler` or `examples` exists than the pin.

## Why Synod

Synod is named after the Synod protocol, the core consensus algorithm in Leslie Lamport's paper "The Part-Time Parliament" (ACM Transactions on Computer Systems, 1998; first submitted in 1990). The paper tells it as the story of the parliament of the ancient Greek island of Paxos, whose part-time legislators could agree on decrees although they came and went; today the algorithm is known as Paxos. Lamport later explained it plainly in "Paxos Made Simple" (2001), and in 2013 he received the ACM A.M. Turing Award for fundamental contributions to the theory and practice of distributed and concurrent systems. His "Time, Clocks, and the Ordering of Events in a Distributed System" (1978) and his TLA+ specification language shaped how we reason about such systems. We are grateful for his work, and we hope Synod honours it by making these ideas easier to watch and understand. These skills are kept concise, so agents can read and write Synod with care.

Implements Synod 0.2 (specification: https://github.com/synodlang/dsl).

Licensed under the [Apache License 2.0](LICENSE).
