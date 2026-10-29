<!-- condensed from dsl v0.3.1 spec/03 -->

# Phrases

A **phrase** is a sentence pattern of words and typed holes, defined once and
then used like a built-in. Everything except the language words and forms is
a phrase, including the whole standard library.

```synod
# package relay
use std/net

define Forward [m: msg] to [n: node]
  net.Send m from n to n
```

## Pattern rules

1. A pattern contains at least one word. Its first word is the **head**.
2. **Receiver.** A pattern may start with one hole before its head, whose type
   is a node type, `Node` or `p.Node`: `define [n: Node] Learn [k: node] at [v: number]`.
   A receiver phrase is node-context code (`contexts.md`) and is called with
   its receiver joined to its head by `.`: `n.Learn k at v`. Every other hole
   comes after the head.
3. No two inline holes are adjacent. `[a: number] plus [b: number]` is fine;
   `[a: number] [b: number]` is not.
4. Every hole is typed, element types included: `[n: node]`,
   `[ns: list of Node]`. Never `[n]` or `[ns: list]`.
5. **Block holes come last.** A hole of type `block` is filled by an indented
   block, not inline. At least one word or inline hole stands between the head
   and the first block hole. After a block hole come only further pairs of one
   word and one block hole; each such word starts a continuation line.
   ```
   define Retry [n: number] times [b: block] otherwise [c: block]
   ```
   is used as
   ```
   Retry 3 times
     net.Send request from n to server
   otherwise
     emit (message GaveUp with Tries 3) by n
   ```
6. The head carries the export capital and, outside its package, the
   qualifier: `net.Send m to n`, `n.gossip.Learn k at v`,
   `look.Fill card with theme.Accent`.
7. Reserved words are never heads and may appear as later words:
   `Send [m: msg] to [n: node]`, `Majority of [n: number]`. Words after the
   head start lower-case (`[n: node] to`, never `To`); only the head carries
   the export capital.
8. Hole types name known types. An unknown type, such as `[x: Foo]` or `p.Node`
   where `p` is not an imported library package, is a `bad-pattern` error.
9. A head is never directly followed by a comparison operator (`is`,
   `is not`, `is in`, `is not in`, `less than`, `more than`, `at least`,
   `at most`): `Count is [n: number]` is a `bad-pattern` error.

## Names and uniqueness

- A phrase's **name** is its word skeleton, with the receiver and holes
  removed: `Learn … at`.
- Within a package every name is defined **once**. There is no overloading by
  hole types, by number of holes, or by having a receiver.
- One namespace per package holds its phrases, its record types, its node type
  `Node` and its node fields. Record fields are scoped to their record type.
- No phrase head, record type or field in any package may be named `Node`
  (`duplicate-name`). A record type may not be named like a built-in type
  (`text`, `number`, `bool`, `time`, `node`, `msg`, `shape`, `color`, `field`,
  `package`, `block`, `value`; `duplicate-name`).
- A package must not define a phrase whose skeleton equals a language form's
  skeleton.

## Resolution

Each sentence is resolved independently.

1. **Candidates** come from the sentence's first item: the language forms; a
   bare word `H`, the current package's phrases without a receiver whose head
   is `H`; a member `p.H` (`p` an imported package), `p`'s exported phrases
   without a receiver with head `H`; a member `x.H` or `x.p.H` (`x` a node
   value), the receiver phrases with head `H` of the current package, or the
   exported ones of `p`.
2. **Match** is by **shape**: the same number of items, literal words agree
   position by position, and the block structure agrees. Types never select
   between phrases. Once exactly one pattern matches, each hole's item must
   have an accepted type; a failure is `type-mismatch`, not `unknown-phrase`.
   For a receiver phrase the first item `x.H` stands for receiver and head
   together: `n.Learn k at v` matches `[n: Node] Learn [k: node] at [v: number]`.
   A hole takes exactly one item, so there is no partial or longest match.
3. **Result.** Exactly one candidate must match by shape. Zero is
   `unknown-phrase`; two is `ambiguous-phrase`, naming both definitions.
4. **Event phrases** are candidates only after `when`, and a `when` takes only
   event phrases; a miss either way is `unknown-phrase`.
5. **The attachment** `w of x` is a candidate only when `w` is a bare word that
   is not the head of a phrase of the current package, so `Majority of n`
   calls the phrase.

### Names

A bare word in an argument position is a **name**, never a literal word of the
pattern. A name is exactly one of:

- a block local or an `each` name;
- a name bound by the enclosing `when`, its receiver included;
- a hole parameter of the enclosing `define`, its receiver included;
- a zero-hole phrase without a receiver of the current package;
- a package name: an import, or the current package's own name as a `package`
  value.

Two of these with the same spelling in scope is a `shadowing` error. This
includes a local, binding, hole or `each` name spelled like a record type of
the current package, and a node field or phrase head spelled like one of the
file's import qualifiers (`n.gossip` would be ambiguous). Imports are scoped
to their file. Inside a `when` event phrase, a bare word that is not yet a
name binds a new one. In a `field` hole, a bare word names a node field of
the current package; it does not read it.

### Members

A member is read left to right. Its first word is a name; each following word
is a named member of what comes before it:

| Before `.` | After `.` | Examples |
| --- | --- | --- |
| a package `p` | an exported phrase head, record type or `Node` of `p`; in a `field` hole, an exported node field of `p` | `net.Send`, `api.Get`, `gossip.Node`, `raft.Role` |
| a record value of type `R` | a field of `R` | `Vote.term`, `best.stamp`, `h.mutation` |
| a node value of type `Node` or `q.Node` | a node field or receiver phrase head of the current package; or a package `p` the node runs, then an exported node field or receiver phrase head of `p` | `n.Known`, `n.Version`, `n.gossip.Known`, `server.tcp.State` |

- A `when` that matches a qualified record type binds the record under that
  spelling, so `api.Get` names the record value and `api.Get.Key` reads its
  field.
- A value of type `node` has no members. `msg` values have no members.
- A member that ends at a field is a value, and can also be a place for
  `set`. A member that ends at a phrase head is a call.

## Results

- A **value phrase** gives its value with `result`, which ends the phrase.
  Every path must reach a `result` (`missing-result`) and all results have one
  type, their join. An **action phrase** has no `result`. A value phrase may
  also have effects.
- A value phrase used as a statement, or an action phrase used as an argument,
  is an error.
- `result` appears only directly in a phrase definition, where the phrase is
  still running. It is a `context` error in `define main`, `start`, `when` or
  `always`, and inside an `after` block or a block argument, because that block
  runs after the phrase has returned.

```synod
# package maths
use std/math

define Majority of [n: number]
  result ((math.Floor (n divided by 2)) plus 1)
```

## Block holes

Inside a `define`, the name of a block hole used alone as a statement **runs
the block**. The block runs in the caller's context and sees the caller's
names as they were when the phrase was called. It holds a copy of the caller's
locals, so setting one of them inside the block is a `bad-place` error. In
node code the block sees the receiver.

```synod
# package timers
define Tick [d: time] [b: block]
  b
  after d
    Tick d
      b
```

A block hole is filled only by an indented block, never by an argument.

## Native phrases

`define <pattern> native "<id>"` binds a pattern to a runtime primitive. Only
packages under `std/` may declare native phrases; you never write one. A std
phrase is usable only in the contexts listed in `std-catalog.md`, and the
compiler enforces that column at every call site.
