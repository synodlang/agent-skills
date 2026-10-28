<!-- condensed from dsl v0.2.14 spec/02 -->

# Grammar

Parsing has two stages. **Structure** turns a file into lines, blocks and
sentences and knows nothing about meaning. **Resolution** matches each
sentence against the language forms and the phrases in scope (`phrases.md`).

## Structure

```ebnf
file         = { line } ;
line         = sentence [ block { continuation } ] NEWLINE ;
block        = NEWLINE INDENT line { line } DEDENT ;
continuation = word block ;
sentence     = items { "," items } ;
items        = item { item } ;
item         = word | member | value | hole | "(" sentence ")" | list ;
hole         = "[" word ":" type "]" ;      (* in define patterns and in when *)
type         = word | word "." word
             | "list" "of" type
             | "map" "of" type "to" type ;
value        = number | duration | text | color | "true" | "false" ;
list         = "{" [ item { "," item } ] "}" ;
```

- Comments attach to the line that follows (or the line they trail) and form
  no block or statement. A file may start with the package header
  (`# package name`), which is also a comment.
- `use` lines stand only at the top level (nested `use` is a `syntax` error);
  `()` is a `syntax` error.
- An **item** is what an inline hole receives: one token, one parenthesized
  sentence, or one list literal. This is why matching is exact.
- A **member** is one item: `n.Known`, `Vote.term`, `server.tcp.State`.
- **Commas** stand only between list items, between `message … with`
  settings and between the two `each` names; anywhere else is a `syntax`
  error.
- **Continuation lines.** A line with a single bare word and a block, at the
  same indentation as the statement before it and directly after that
  statement's block, joins that statement as its next word and block hole.
  `else` continues `if`; `otherwise` continues `Retry … times`. A
  continuation line that follows no block is a `syntax` error. Blank and
  comment lines may stand between a block and its continuation line.

## Language words

`arg` is one item that is not a hole.

```ebnf
(* files and entries *)
import     = "use" path [ "as" word ] ;
definition = "define" pattern ( block | "native" text ) ;
pattern    = [ hole ] head { word | hole } ;     (* first hole: the receiver *)
head       = word ;                              (* never a reserved word *)
main       = "define" "main" block ;
start      = "define" hole "start" block ;       (* hole: the receiver [n: Node] *)

(* state and records *)
typedecl   = word "has" block ;                  (* block holds only "set" word "to" arg lines *)
assignment = "set" place "to" arg ;
place      = word | member | "item" arg "of" ( word | member ) ;

(* events *)
handler    = "when" [ hole ] arg block ;         (* receiver: present exactly on node events *)
delay      = "after" arg block ;
delivery   = "deliver" arg "to" arg ;
emission   = "emit" arg "by" arg ;

(* control *)
choice     = "if" arg block [ "else" block ] ;   (* else is a continuation line *)
loop       = "each" word [ "," word ] "in" arg block ;
return     = "result" arg ;

(* checking *)
invariant  = "always" arg ;

(* visual *)
drawing    = "draw" arg "as" word [ "for" arg ] [ block ] ;
attachment = word "of" arg ;                     (* look context *)

(* values *)
record     = "message" ( word | word "." word ) "with" setting { "," setting } ;
setting    = word arg ;
element    = "item" arg "of" arg ;
compare    = arg ( "is" | "is not" | "is in" | "is not in"
                 | "less than" | "more than" | "at least" | "at most" ) arg ;
logic      = arg ( "and" | "or" ) arg | "not" arg ;
arithmetic = arg ( "plus" | "minus" | "times" | "divided by" ) arg ;
```

- `Node has` declares the fields of the package's node type; `T has` with any
  other word `T` declares the record type `T`.
- Members are items, not language forms. They name a package member
  (`net.Send`, `gossip.Node`), a record field (`Vote.term`, `api.Get.Key`), a
  node field (`n.Known`, `server.tcp.State`) or a receiver phrase
  (`n.Version`, `n.Learn k at v`).

## Consequences

- **No precedence.** An inline hole takes one item, so every nested phrase is
  parenthesized: `((a plus b) times c)` is the only spelling of that meaning.
- **Conditions are parenthesized** like every multi-item argument:
  `if (n.Role is "leader")`, `always (Count less than 2)`. A single-item
  condition is bare: `if grant`, `if Vote.granted`.
- **A zero-hole phrase is one item**, written bare: `time.Now`,
  `theme.Accent`, `Quorum`, `raft.Leaders`, `n.Version`.
- **Members are never parenthesized:** `n.Known`, `Vote.term`.
- **Parentheses are removed, never added, by the formatter.** It unwraps a
  parenthesized single item but never wraps a multi-item argument; a missing
  pair is reported by the compiler (`unknown-phrase`).
- **Entries.** `define main` and `define [n: Node] start` are language forms;
  `main` and `start` are reserved words and never phrase heads.
- **`else`** is a continuation line: its own line, at its `if`'s indentation,
  directly after the `if` block. `else if` does not exist; nest an `if`
  inside `else`.
- **Binding in `when`.** `when [n: Node] (net.Received digest from sender)`
  binds the receiver `n`, the record `digest` and the peer `sender` for the
  block.
- **Locals.** `set x to …` with a bare word `x` creates a local for the rest
  of the enclosing block, nested blocks included, or updates it. Locals are
  **block scoped**: one created inside an `if` branch, an `else` branch, an
  `each` body or another block is not visible after that block, so a value
  needed afterwards is set before the block. Hole parameters, `when`
  bindings and `each` names cannot be set (`bad-place`), nor can a local from
  outside an `after` block or a block argument be set inside it (`bad-place`;
  the block holds a copy). A local is **used** when it is read, or is the
  base of `set item … of x`; setting it again does not use it
  (`unused-name`).

```synod
# package counter
use std/lists

Node has
  set Total to 0
  set Seen to {}

define [n: Node] Add [v: number]
  set doubled to (v times 2)
  if (v more than 10)
    set n.Total to (n.Total plus doubled)
  else
    set n.Seen to (lists.Append v to n.Seen)
```
