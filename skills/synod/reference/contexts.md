<!-- condensed from dsl v0.3.0 spec/05 -->

# Nodes, state and contexts

## Node types and running

- Every library package has a **node type**, `Node` inside the package and
  `gossip.Node` outside. A node of type `gossip.Node` runs `gossip`.
- `sim.Spawn raft` creates a node of type `raft.Node`. Running a package also
  runs, transitively, every package it uses: a node running `cassandra` also
  runs `gossip` and `ring`, so a `cassandra.Node` is accepted wherever a
  `gossip.Node` is expected.
- When a node starts, each package it runs is started once, in dependency
  order (dependencies first).
- `node` is the type of any node. It carries identity only: it can be
  compared, used as a map key, sent to, and named with `sim.Id n` and
  `sim.Name n`, and it has no fields.

## Node state

```synod
# package state
use std/lists

Node has
  set Heartbeat to 0
  set Known to {}

define [n: Node] Beat
  set n.Heartbeat to (n.Heartbeat plus 1)

define [n: Node] Learn [k: node]
  set n.Known to (lists.Append k to n.Known)
```

- `Node has` declares the fields of the package's node type. Every node
  running the package has its own copy, initialized when the node is created.
  A package without `Node has` has a node type with no fields.
- Fields are reached through a node value with `.`: `n.Heartbeat` in the
  declaring package, `n.gossip.Heartbeat` from another package (exported
  fields only, and `n`'s type must run `gossip`). A field is read or written
  only through a node value. A `field` hole names a field without reading it:
  bare in its own package, `raft.Role` elsewhere.
- Identity comes from `std/sim` phrases, readable in every context: `sim.Id n`
  (1, 2, 3… in creation order) and `sim.Name n` (`"n1"`, `"n2"`…).

## Receivers and start

Node code names its node with a **receiver**: a hole of a node type before
the head of a definition, or after `when`.

```synod
# package versions
use std/lists
use std/net

Node has
  set Generation to 0
  set Heartbeat to 0
  set peers to {}

define [n: Node] Version
  result ((n.Generation times 1000000) plus n.Heartbeat)

when [n: Node] (net.Received digest from sender)
  if (sender is not in n.peers)
    set n.peers to (lists.Append sender to n.peers)
```

- The receiver of a phrase has type `Node` or `p.Node`. The receiver of
  `start` and of a `when` has type `Node`.
- `define [n: Node] start` is the node entry. It runs on every node running
  the package when the node is created, and again on every restart. A package
  has at most one `start`.
- A file root holds only declarations: `use`, `has`, `define`, `when`,
  `always`. No statement runs at file root.
- Blocks inside node code (`after`, the block of `time.Every`) see the
  receiver by ordinary lexical scope:

```synod
# package heart
use std/disk
use std/time

Node has
  set Generation to 0
  set Heartbeat to 0

define [n: Node] start
  set n.Generation to ((disk.Read "generation" on n or 0) plus 1)
  disk.Write n.Generation at "generation" on n
  time.Every 1s
    set n.Heartbeat to (n.Heartbeat plus 1)
```

- Receivers, `start`, `when` on node events and `Node has` are declared only
  in library packages: a main package never runs on nodes, so it has no node
  type (`bad-receiver`).

## The four contexts

Every statement runs in exactly one context. A context starts at an entry, and
everything that code runs keeps it: nested blocks, `after` blocks, block
arguments, and phrases without a receiver called from it.

| Context | Starts at | Reads | Writes | May use |
| --- | --- | --- | --- | --- |
| **node** | `start`; `when [n: Node]` on node events; receiver phrases | the receiver's state; identity of any node | the receiver's state; locals | `net.Send`, `disk.*`, `after` (simulated), `emit … by n` |
| **main** | `define main` | any node's state, through `Node` and `p.Node` values | any node's state (setup and faults); locals | `deliver`, `after` (simulated), `emit … by sim.Nobody`, `sim.*` setup and fault phrases, `net.*` configuration |
| **check** | `always` | any node's state | locals only | nothing with effects |
| **look** | `when` on trace events | exported state of any node, as it was after that trace event; a lower-case field is not readable, even of the same package, and neither is a lower-case field of a record (`context`) | locals and shapes only | `draw`, `std/look`, `after` (presentation) |

- **Receiver phrases are node code.** A receiver phrase is called only in node
  context, on the receiver in scope. A call from main, check or look context
  is a `context` error.
- **Realism.** In node context, node state is reached only through the
  receiver of the `define` or `when` that lexically contains the code:
  `n.Known`, `n.gossip.Known`, `n.Version`. Any other node value (a peer, a
  hole, a list item) gives identity only, and other nodes learn through
  `net.Send`. Anything else is a `foreign-state` error, reported at compile
  time.
- **Explicit naming.** In node context, `net.Send`, `disk.*` and `emit` name
  their node with the words of the trace fact they record:
  `net.Send m from n to p`, `disk.Write v at k on n`,
  `disk.Read k on n or d`, `emit e by n`. The named node must be the receiver
  in scope; any other node is a `foreign-state` error. These phrases are
  node-context only, except `emit`, which main also uses with `sim.Nobody`:
  `emit (message X with Y 1) by sim.Nobody`; in main any other node after `by`
  is a `context` error. `after` schedules on the context it runs in: in node
  context, on the receiver's node. `lesson.Ask` (both spellings, with and
  without `choosing`) and `lesson.Note` are main context only.
- **Observers see the state after their fact.** In
  `when (sim.Changed f of n from old to new)`, `n`'s field already holds
  `new`. Only exported fields are readable, node fields and record fields
  alike; the trace masks the others, so a look cannot depend on them.
- **Purity of looks.** Observers cannot `set` node state, `deliver`, `emit`,
  or send. A look can never change what happens, only how it looks.
- **Shared phrases.** A phrase without a receiver takes the context of its
  caller and is checked once per context it is reached from. In node context
  it has no receiver, so it reaches no node state. In main and check context
  it reaches nodes through its holes: `define Join [n: Node] to [seed: node]`
  sets `n.peers` from `define main`. A context error inside a shared phrase
  is reported once, at the statement in its body.
- **Std phrases.** The Context column of `std-catalog.md` is enforced at the
  call site of every std phrase: `time.Now` in look context is a `context`
  error, though `time.Every`'s body would allow it.

## Node events vs trace events

`when` takes an event phrase. Event phrases are native and come in two kinds:

- **Node events** are delivered to one node: `net.Received`, `sim.Delivered`.
  A `when` on a node event has a receiver, `when [n: Node] (…)`. It runs on
  every node running the package that contains it, with `n` bound to that
  node, in node context.
- **Trace events** are facts recorded in the trace: `net.Sent`, `sim.Changed`,
  `sim.Emitted`… (full list in `trace-events.md`). A `when` on a trace event
  has no receiver. It runs once per occurrence, in look context, for every
  package in the program that contains it.

A receiver on a trace-event `when`, or a node-event `when` without one, is a
`bad-receiver` error.

## Matching events

In `when [n: Node] (net.Received Prepare from proposer)`:

- A **record type name** in a `msg` hole filters by type and binds the record
  under that name (`Prepare.number`). A qualified type binds under the
  qualified name (`api.Get.Key`).
- A **typed hole** `[x: p.Node]` in a `node` hole filters to nodes running `p`
  and binds `x` with that type: `when (sim.Added [n: raft.Node])`.
- In a `field` hole (`sim.Changed`), a bare word that is a node field of the
  current package, or an exported `p.F`, filters by that field, and `old` and
  `new` bound later take its type; any other bare word binds a value of type
  `field`, with `old` and `new` of type `value`. Filtering by an unexported
  field of another package is `unexported`.
- A **new bare word** in any other hole binds the value with the hole's type
  (`proposer`: `node`).
- An **existing name or a literal** filters by equality.

## Invariants

```synod
# package decide
use std/lists
use std/sim

Node has
  set Decided to 0

define Decisions [ns: list of Node]
  set values to {}
  each n in ns
    if (n.Decided is not in values)
      set values to (lists.Append n.Decided to values)
  result values

always ((lists.Size (Decisions (sim.NodesRunning decide))) at most 1)
```

`always` takes a `bool`. It is evaluated after every simulation step in check
context. When it becomes false the trace records `sim.Broken`, once per
transition from true to false; the run continues.
