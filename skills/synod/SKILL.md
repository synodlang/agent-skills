---
name: synod
description: Read, write, check and run Synod (.synod) programs that model
  distributed systems, system designs and database internals for animation.
  Use whenever a task mentions Synod, .synod files, synod.mod, or asks to
  "animate" or "simulate" a protocol in a Synod workspace.
---

<!-- written for dsl v0.3.1 -->

# Writing Synod

Synod programs are plain-text sentences in `.synod` files. You write them,
`synod` checks and runs them. Never hand back unchecked code.

## Workflow

1. **Look.** Run `synod doc` to list the packages and `synod doc <package>` for
   one package's phrases. Search the workspace for an existing package. To use
   the examples, `synod get github.com/synodlang/examples@v0.3.0` (the user's
   step when you have no shell), then `use github.com/synodlang/examples/twopc`
   (it is copied to `vendor/`).
2. **Write.** Prefer only a main file (`define main`, typically 10-30 lines)
   that uses an existing package. Write a new package only when none models
   the behaviour.
3. **Format.** `synod fmt <paths>` makes the code canonical.
4. **Check.** `synod check` until it is clean. Piped output is the one-line
   form with a hint, which is what you read.
5. **Run.** `synod run <folder> --summary` prints a compact trace summary.
   Read its last lines and confirm the scenario shows what was asked (for
   example, that a leader was actually elected), then stop.
6. **Confirm** in your answer what the run showed, in a sentence or two.

## The rules

1. Look before writing (workflow step 1).
2. Every argument is one item: parenthesize every nested phrase and every
   multi-item condition. `if (n.Role is "leader")`, `set n.x to (n.x plus 1)`.
   A single-item condition is bare: `if grant`. The formatter removes
   parentheses but never adds them.
3. Qualify every imported name (`net.Send`, `gossip.Join`,
   `n.gossip.Known`); never qualify your own package's names (`n.Known`,
   `Quorum`).
4. Node code names its node with a receiver: `define [n: Node] Version`,
   `when [n: Node] (net.Received …)`, `define [n: Node] start`. It touches
   only `n`'s state and talks to other nodes with `net.Send m from n to p`.
   `deliver` and faults live in `define main`.
5. Types are explicit: `[ns: list of Node]`, `[m: map of node to number]`.
   A `node` has no fields; `Node` or `pkg.Node` does.
6. Looks and themes are optional: `std/look/basic` with `std/theme/default`
   animates everything by default (see Themes).
7. Run `synod fmt`, then `synod check` until clean, then `synod run` and
   confirm the trace shows what was asked.
8. Use `synod check --json` for spans, edits and quick fixes, or
   `--style rich` to see the source line. `synod explain <code>` explains
   any diagnostic with an example and its fix.

## A package and its main file

A module has `synod.mod` at its root (`synod new <dir>` creates one). A folder
holds one package; `# package <name>` is the first line of a file. A main
program is its own folder with `# package main`.

```synod
# package ping
use std/net
use std/sim

Go has
  set rounds to 0

Ping has
  set seq to 0

Pong has
  set seq to 0

Node has
  set Replies to 0
  set peers to {}

define Join [n: Node] to [seed: node]
  set n.peers to {seed}

when [n: Node] (sim.Delivered Go)
  each p in n.peers
    net.Send (message Ping with seq Go.rounds) from n to p

when [n: Node] (net.Received Ping from sender)
  net.Send (message Pong with seq Ping.seq) from n to sender

when [n: Node] (net.Received Pong from sender)
  set n.Replies to (n.Replies plus 1)
```

```synod
# package main
use github.com/acme/demo/ping
use std/net
use std/sim

define main
  sim.Seed 1
  net.Latency between 10ms and 30ms
  set nodes to (sim.Cluster of 2 running ping)
  ping.Join (item 0 of nodes) to (item 1 of nodes)
  deliver (message ping.Go with rounds 1) to (item 0 of nodes)
  sim.StopAt 2s
```

The module here is `github.com/acme/demo`. Run it with `synod run <folder>`.

## Reuse rules

- Run `synod doc` before writing any phrase; use what exists.
- Never copy or rewrite an existing package. Import it.
- Never edit `vendor/`: it holds verified copies, and `synod.sum` pins them.
- Do not invent phrases. If `synod check` says `unknown-phrase`, look the
  phrase up with `synod doc <package>`.
- A new package gets its own folder, one package per folder.

## Themes

A theme is the skin of a program (colors, type, stroke, sketch, glow, shadow,
vignette, easing, node shape and fill). Its files start `# theme <word>` and
hold only `define <Name>` blocks with one `result <literal>` each, one for
every name of the theme table. The table, with the default values, is in
`std-catalog.md` (`std/theme`, `std/theme/default`) or `synod doc
std/theme/default`; copy it and change values, never write a partial theme
(`bad-theme` names the first missing name).

- A program's theme is the one its main file imports, else
  `std/theme/default`: `use github.com/synodlang/looks/terminal`. Only a main
  file imports a theme, and at most one (`theme-outside-main`, `extra-theme`).
  Never write `use std/theme/default` (`non-canonical`).
- A theme exports no phrase; a look reads a value as `theme.<Name>`.
- A theme never changes a run: the trace is the same under every theme. To
  restyle, change the look or theme; never to fix behavior.
- A theme is skin; a look is what is drawn and how it moves (arcs, trails,
  entrances, effects, camera). Any look combines with any theme.
- The module `github.com/synodlang/looks` (`synod get`) holds the looks
  `lumen` and `doodle` and the themes `solarized/light`, `solarized/dark`,
  `terminal`, `sketch` and `chalk`, for example `use
  github.com/synodlang/looks/lumen` with `use
  github.com/synodlang/looks/solarized/dark`.

## Commands

| Command | Use |
| --- | --- |
| `synod doc [package]` | list packages, or one package's phrases (`--json`) |
| `synod fmt <paths>` | rewrite files to canonical form (`--check` to test) |
| `synod check [paths]` | diagnostics; `--json` for spans and fixes |
| `synod explain <code>` | what a diagnostic code means, with an example |
| `synod run <folder> --summary` | check, simulate, print a compact trace summary |
| `synod get <module>@<version>` | download a module into `vendor/` |
| `synod new <dir>` | start a module with a runnable demo |

## If the synod-editor tools are present

When tools named `mcp__synod-editor__*` exist, the Mac app's editor holds the
files. Do not write files any other way.

0. **Look.** There is no shell in this mode. Use the `synod` MCP tools `doc`,
   `explain`, `check` and `run_summary`, and Read, Glob and Grep on the
   reference files. When only the editor and `synod` tools exist, `synod get`
   is the user's step: if a needed package is not in `vendor/`, stop and ask
   the user to run `synod get github.com/synodlang/examples@v0.3.0` (or use the
   bundled examples under the reference folder as a model).
1. `read_buffer` a file before editing it (it shows unsaved changes and
   returns the `version`).
2. Change it with `apply_edits` (small edits, needs the current `version`) or
   `replace_buffer` (whole text). After a `version-mismatch`, read again.
3. Make new files with `create_file`.
4. Call `get_diagnostics` (fresh) until it is clean.
5. `save_files`, which saves and formats like the user's Save.
6. Only then `run_summary`.

Use the `synod` MCP tools `doc`, `explain`, `check` and `run_summary` when
present; they give the same output as the commands above. Do not try `Bash`,
`Write` or `Edit`; they are not available.

## Budget rules

- Never paste whole traces or scenes into a prompt or an answer.
- Never ask for or invent coordinates, colors or timings; looks decide those.
- Read `--summary`, not the raw trace.
- Make small edits to existing files instead of regenerating them.

## Reference files

They are in `reference/` next to this file.

| File | Read it for |
| --- | --- |
| `grammar.md` | how a line is parsed; the language words and their shapes |
| `phrases.md` | defining and calling phrases; patterns, resolution, names |
| `contexts.md` | nodes, state, the four contexts, events, `always` |
| `canonical.md` | the canonical form `synod fmt` writes |
| `std-catalog.md` | every std phrase, its type, context and meaning; the theme table |
| `trace-events.md` | the facts a run records, for `when` in looks |
| `errors.md` | every diagnostic code with an example and its fix |
| `examples/` | complete working programs (gossip, raft, paxos, twopc, …) |

For the user, or when you can read the web, the documentation at
https://synodlang.github.io/docs/v0.3/ has a 15-chapter tutorial (`tutorial/`),
the Library pages (`library/`, with a runnable example for each std package,
also in `std/examples` of the dsl repository) and the error pages (`errors/`).
Its guide "Use Synod with Claude or another agent" explains this skill.

Written for dsl v0.3.1.
