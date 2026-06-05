# Zero Roadmap

## Phase 0: Bootstrap (4-6 weeks)

Goal: Working Zero in Go that compiles and runs Zero programs.

### Weeks 1-2: Foundation
- value.go — value system
- opcode.go — instruction definitions
- chunk.go — bytecode container
- vm.go — interpreter
- builtins.go — print, len, type
- First working: print(" hello\)

###
Weeks
3-4:
Compiler
-
lexer.go
—
27
token
types
-
parser.go
+
ast.go
—
recursive
descent
-
compiler.go
—
expressions,
variables,
functions
-
repl.go
—
interactive
REPL
-
Tests
pass:
arithmetic,
if,
loops,
functions,
recursion,
arrays

###
Weeks
5-6:
Polish
-
Error
messages
with
source
locations
-
Type
checking
pass
-
Closures
-
Module
system
-
Standard
library
(io,
json,
http)
-
All
Phase
0
tests
green

##
Phase
1:
Self-Hosting
(6-8
weeks)

Goal:
Zero
compiler
written
in
Zero,
bootstrapped
via
Phase
0.

-
lexer.zero,
parser.zero,
codegen.zero
-
Bootstrap:
Phase
0
compiles
Phase
1
-
Bytecode
verification
between
phases
-
Compiler
compiles
itself

##
Phase
2:
Unique
Features
(8-12
weeks)

Goal:
Zero's
unique
language-level
features.

-
ai
fn
construct
-
Effect
tracking
-
Image
snapshots
-
Actor
model
-
Pipeline
optimizer
-
Pattern
matching
engine

##
Phase
3:
Browser/AI
Runtime

Goal:
Zero
powers
real-world
AI
agents.

-
std/browser
—
CDP
integration
-
std/ai
—
LLM
integration
via
EchoBird
-
std/web
—
uTLS
fingerprinting
-
Guard
system
—
captcha
detection,
cooling
-
Adapters
—
JD,
WeChat,
Xiaohongshu