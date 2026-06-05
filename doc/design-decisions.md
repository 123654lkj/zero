# Zero Design Decisions

## Stack VM vs Register VM

**Decision: Stack VM**

Stack VMs produce more instructions but simpler compilers. For Zero's use cases (browser automation, AI pipelines), execution bottlenecks are I/O and page rendering, not instruction throughput. The simplicity wins.

## Value Representation

**Decision: Tagged union with small-value optimization**

8-byte values with 1-byte type tag. Small ints and short strings embedded directly. No heap allocation for common cases. Borrowed from LuaJIT.

## Null Safety

**Decision: No null**

Zero has no null, nil, None, or undefined. Every nullable value is `Option<T>`. The `?` operator unwraps safely. Pattern matching ensures exhaustive handling.

## Error Handling

**Decision: Result<T, E> with ! propagation**

Errors are values, not control flow. `!` suffix propagates errors up the call chain. Pattern matching handles specific error cases. No try/catch/throw.

## Memory Management

**Decision: ARC + RAII**

Automatic Reference Counting with RAII destructors. No borrow checker, no stop-the-world GC. Predictable performance, deterministic cleanup.

## Type System

**Decision: Gradual typing with full inference**

Start with dynamic typing, add types as needed. The compiler infers everything it can. `any` type for untyped code. Structural typing for compatibility.

## Effects

**Decision: Explicit effect tracking**

Every function's side effects are part of its type. Pure functions are guaranteed pure. IO, NetIO, Error effects are tracked through the call chain.

## Concurrency

**Decision: Goroutines + Actors**

Goroutines for lightweight parallelism. Actors for stateful concurrent processes. `spawn`, `send`, `receive` as core primitives.

## Bootstrapping

**Decision: Three-phase self-hosting**

Phase 0: Go bootstrap. Phase 1: Self-hosted bytecode. Phase 2: Native compilation. Each phase verifies against the previous.