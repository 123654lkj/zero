# Zero

**A programming language that learns from every language that came before — and fixes what they got wrong.**

Zero is not another language. It is the synthesis of 60 years of programming language design. Every feature in Zero exists because it was proven to work somewhere. Every missing feature exists because it was proven to break somewhere.

> From C to Rust, from Haskell to Erlang, from Smalltalk to SQL — Zero absorbs their truths and rejects their compromises.

---

## Philosophy in One Line

**Zero doesn't add features. Zero removes problems.**

---

## Quick Start

Note: Zero is currently in bootstrap phase. The first compiler is written in Go, and will be replaced by Zero itself once self-hosting is achieved.

\\\ash
git clone https://github.com/zero-lang/zero.git
cd zero
./bootstrap.ps1
zero run hello.zero
\\\

\\\zero
print(" Hello from Zero!\)
\\\

---

##
Key
Innovations

###
1.
No
Null
—
Option<T>
replaces
null.
Compiler
enforced.

###
2.
Errors
Are
Values
—
Result<T,
E>
with
!
propagation.
No
try/catch.

###
3.
AI-Native
—
ai
fn
is
a
language
construct.
Compiler
understands
AI
semantics.

###
4.
Effects
Tracking
—
Pure
functions
guaranteed
pure.
IO/NetIO
tracked
through
the
type
system.

###
5.
Image
Snapshots
—
Snapshot,
restore,
replay
the
entire
VM
state.

###
6.
Actor
Model
—
spawn/send/receive
as
core
primitives.

###
7.
Declarative
Pipelines
—
|>
operator
auto-parallelizes
data
flow.

---

##
Language
at
a
Glance

\\\zero
//
No
null.
Patterns
instead
of
exceptions.
enum
Result
{

Ok(data:
string)

Error(msg:
string)

Blocked(secs:
int)
}

match
res
{

Ok(d)
->
process(d)

Error(m)
->
log(m)

Blocked(s)
->
wait(s)
|>
retry()
}

//
AI
calls
are
language-level
ai
fn
extract_price(html:
string)
->
float

//
Data
pipelines
auto-parallelize
let
data
=
from
\https://jd.com\

|>
search
\iPhone\

|>
extract
price,
title

|>
filter
price
<
5000

//
Browser
is
native
I/O
fn
get_price(id:
string)
->
float
{

let
p
=
web.open(\https://item.jd.com/  \)

wait
p.selector(\.sku-price\)

return
p.text(\.sku-price\)?.clean_price()
}
\\\

---

##
Bootstrapping

|
Phase
|
Written
In
|
Output
|
Status
|
|-------|-----------|--------|--------|
|
0
|
Go
|
Bytecode
(.zbc)
|
Building
|
|
1
|
Zero
|
Bytecode
(.zbc)
|
Planned
|
|
2
|
Zero
|
Native
binary
|
Future
|

---

##
Project
Structure

\\\
zero/
├──
go/
#
Bootstrap
compiler
(Go)
│
├──
value/
#
Value
system
│
├──
opcode/
#
Instruction
set
│
├──
chunk/
#
Bytecode
container
│
├──
compiler/
#
Lexer,
Parser,
Compiler
│
└──
vm/
#
Bytecode
interpreter
├──
src/
#
Self-hosted
compiler
├──
tests/
#
Test
programs
├──
doc/
#
Design
docs
└──
README.md
\\\

---

##
License

MIT