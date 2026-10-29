# Examples

Canonical sample programs for Synod: protocols and systems you can run, break and read.

[![ci](https://github.com/synodlang/examples/actions/workflows/ci.yml/badge.svg)](https://github.com/synodlang/examples/actions/workflows/ci.yml)
[![licence](https://img.shields.io/badge/licence-CC%20BY%204.0-lightgrey)](LICENSE)

They form one module, `github.com/synodlang/examples`. Each behaviour
package folder holds a `demo.synod` main file (`# package main`) next to the
library; `webapp/demo.synod` is a pure program in a folder of sub-packages.

## Where it fits

These programs are the acceptance tests of the language: the
[compiler](https://github.com/synodlang/compiler) must check and run every one,
with a story that matches the text below. They are also the best way to learn
Synod after the first chapter of the
[specification](https://github.com/synodlang/dsl).

| Repository | Its part |
| --- | --- |
| [dsl](https://github.com/synodlang/dsl) | The language the examples are written in |
| [std](https://github.com/synodlang/std) | The standard library they import |
| [compiler](https://github.com/synodlang/compiler) | `synod`, which runs them; it keeps a vendored copy for its tests and fails if the copy drifts |
| `examples` (this one) | The programs |

## Run them

Install `synod` (https://synodlang.github.io/install/), for example:

```
brew install synodlang/tap/synod
```

Then, from a clone of this repository:

```
git clone https://github.com/synodlang/examples
cd examples
synod check
synod run ./tcp
```

`synod check` with no path checks every package here and prints nothing when
all is well. `synod run <folder>` simulates the folder's `demo.synod` and
writes `out/trace.jsonl` and the animation `out/scene.json`; add `--summary` to read the run as text, `--seed n`
to try another seed, `--trace file` to write the trace somewhere else, and `--scene file`, `--pace x` and `--mode proportional` to change the scene.
The run is deterministic: the same seed gives the same trace, byte for byte.

## The examples

| Folder | Topic | What to watch | Seed | Run |
| --- | --- | --- | --- | --- |
| `tcp` | TCP three-way handshake | SYN, SYN-ACK, ACK and the state changes on both ends | 1 (default) | `synod run ./tcp` |
| `gossip` | Gossip membership and failure detection | Heartbeats spreading; a crashed node marked down after 4s; generations after restart | 15 | `synod run ./gossip` |
| `paxos` | Single-decree Paxos (the Synod protocol) | Two racing proposers; the higher ballot adopts any accepted value; the invariant "at most one chosen value"; promises and accepted values on `std/disk`, so a restarted node keeps them | 105 | `synod run ./paxos` |
| `raft` + `raftlook` | Raft leader election | Randomized timeouts, votes, a leader crash, a new term; "never two leaders in a term" | 3 | `synod run ./raft` |
| `twopc` | Two-phase commit | The blocking problem when the coordinator crashes between phases; the atomicity invariant | 1 | `synod run ./twopc` |
| `ring` + `ringlook` | Consistent hashing with virtual nodes | Which keys move when a fifth node joins | 1 (default) | `synod run ./ring` |
| `cassandra` | A Cassandra-style cluster: gossip + ring + replication factor 3 + QUORUM + hinted handoff | A write succeeding with a replica down, and the hint replayed after restart | 3 | `synod run ./cassandra` |
| `webapp` | Cache-aside behind a load balancer (`api`, `lb`, `app`, `cache`, `db`) | Hit rate warming up; what happens when the cache crashes and there is no timeout | 9 | `synod run ./webapp` |
| `wal` | Write-ahead log and checkpoint recovery | Which writes survive a crash between checkpoints | 1 (default) | `synod run ./wal` |

Each `demo.synod` asks a question with `lesson.Ask … choosing` at the
interesting moment: three or four choices, one of them what the run shows and
the others common misconceptions. Some demos also add a caption with
`lesson.Note`. They appear in the trace as
`lesson.Asked` and `lesson.Noted`. `raftlook` and `ringlook` are looks: they
draw the cluster in `out/scene.json`. The other examples are drawn by the
default look, `std/look/basic`. The
`always` invariants (`paxos`, `raft`, `twopc`) are checked after every step;
`--summary` prints an `always:` block saying whether each held.

## A real run

This is the end of `synod run ./twopc --summary`, produced with the current
compiler. The coordinator, n1, crashed at 45ms, after the participants voted
yes and before it decided. The last block is the point of the example: n2, n3
and n4 are all `prepared`, and none can commit or abort alone.

```
@32.533ms receive n1
  net.Arrived #4 twopc.Vote{tx 1, yes true} from n3 to n1
@38.926ms receive n1
  net.Arrived #6 twopc.Vote{tx 1, yes true} from n4 to n1
@45ms timer main
  sim.Crashed n1
  sim.Emitted lesson.Asked{Question "The coordinator crashed while votes were in flight. What can a prepared participant do?", Choices {"Commit, because it voted yes", "Abort on its own, because the coordinator is gone", "Wait for the coordinator: it voted yes, so it cannot decide alone", "Elect a new coordinator among the participants"}} by nobody
@45.881ms receive n1
  net.Dropped #5 twopc.Vote{tx 1, yes true} from n2 to n1 (crashed)
  sim.Ended empty
final:
  n1 down  twopc.State "waiting", twopc.Participants {n2, n3, n4}
  n2 up  twopc.State "prepared", twopc.Participants {}
  n3 up  twopc.State "prepared", twopc.Participants {}
  n4 up  twopc.State "prepared", twopc.Participants {}
```

## Layout

```
tcp/  gossip/  paxos/  raft/  twopc/  ring/  cassandra/  wal/
    <name>.synod    the library package
    demo.synod      the main file: builds a cluster, injects requests and faults
raftlook/  ringlook/   look packages used by the raft and ring demos
webapp/
    demo.synod      the main file
    api/  app/  cache/  db/  lb/    one package per role
synod.mod           module github.com/synodlang/examples, synod 0.3
```

Folders that build on others show how packages compose: `cassandra` runs
`gossip` and `ring` on the same nodes, and `webapp` wires five roles together
with phrases such as `lb.Balance` and `app.Wire`.

## Why Synod

Synod is named after the Synod protocol, the core consensus algorithm in Leslie Lamport's paper "The Part-Time Parliament" (ACM Transactions on Computer Systems, 1998; first submitted in 1990). The paper tells it as the story of the parliament of the ancient Greek island of Paxos, whose part-time legislators could agree on decrees although they came and went; today the algorithm is known as Paxos. Lamport later explained it plainly in "Paxos Made Simple" (2001), and in 2013 he received the ACM A.M. Turing Award for fundamental contributions to the theory and practice of distributed and concurrent systems. His "Time, Clocks, and the Ordering of Events in a Distributed System" (1978) and his TLA+ specification language shaped how we reason about such systems. We are grateful for his work, and we hope Synod honours it by making these ideas easier to watch and understand. The `paxos` example is Synod's own namesake algorithm, written in Synod so you can watch it run.

## Things to try (Break it)

- **Paxos:** delete `set best to 0` and the four-line loop below it that
  adopts the highest prior value (the program still checks), change the
  demo's latency to `net.Latency between 1ms and 40ms`, run seeds 1–200
  (`synod run ./paxos --seed 69`), and watch `always` break (it does for 6
  of the 200 seeds; with the demo's own 10ms to 40ms no seed in 1–200
  breaks it).
- **Raft:** shorten the election timeout range to `150ms and 160ms` and
  watch split votes.
- **Two-phase commit:** move the coordinator crash to `after 80ms`; the
  decision now reaches the participants.
- **Cassandra:** set `Quorum` to 3; the second write no longer succeeds
  while a replica is down.
- **Web app:** add a timeout in `app` so requests fail fast when the cache is down.

## Notes for implementers

- These programs are the acceptance tests of `implementation/10-roadmap.md` in [dsl](https://github.com/synodlang/dsl).
- Timings in the demos (for example the 45ms coordinator crash) were
  chosen to land between protocol phases with typical latencies. If the
  golden trace for a seed does not show the intended story, adjust the
  timing or the seed, never the protocol, and record it here.
- tcp: the closing note at 100ms lands after both ends are ESTABLISHED
  (at most 90ms with the default 10-30ms latency).
- cassandra's record timestamps are named `stamp` (`at` is reserved).
- A restarted node starts again from the initial values of its fields,
  including the ones `main` set (for example the peers `gossip.Join`
  sets); only its disk survives.
- Seeds are chosen so that each demo tells its story under the simulator of
  `synod run` (the stories are tested in `compiler`,
  `internal/runtest/story_test.go`). Changes made for that, never to a
  protocol: gossip `sim.Seed` 4 to 15 (with 4 a live node suspects a
  healthy peer between 5s and 8s); paxos 7 to 105 (the late proposer's
  Accept carries the adopted `"x"`, as the text above promises); twopc 11
  to 1 (with 11 a decision was sent before the crash at 45ms and no
  participant stayed prepared); cassandra 5 to 3 (with 5 no hint was kept
  and the restarted replica stayed empty). tcp, raft, webapp, wal and ring
  were unchanged. `math.Hash` mixes FNV-1a with the MurmurHash3 finalizer
  so that keys such as `user:1` and `user:2` land far apart on the ring.

Implements Synod 0.3 (specification: https://github.com/synodlang/dsl).

Licensed under [CC BY 4.0](LICENSE).
