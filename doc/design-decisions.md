# Zero Design Decisions

## Stack VM vs Register VM
**Stack VM.** Simpler compiler, more instructions. Bottleneck is I/O not instruction throughput.

## Value Representation
**Tagged union + small-value optimization.** 8 bytes, 1 byte type tag. Small ints/strings embedded inline.

## Null Safety
**No null.** Option<T> enforced by the compiler. No null pointer exceptions.

## Error Handling
**Result<T, E> + ! propagation.** No try/catch/throw. Errors are values. The ! operator propagates concisely.

## Memory Management
**ARC + RAII.** No borrow checker, no stop-the-world GC. Predictable, deterministic.

## Type System
**Gradual + full inference.** Start dynamic, add types as needed. Structural typing. any for untyped.

## Effects
**Explicit tracking.** Pure functions guaranteed. IO/NetIO/Error tracked through call chain.

## Concurrency
**Goroutines + Actors.** Lightweight parallelism + stateful processes.

## Bootstrapping
**Three phases.** Go bootstrap -> self-hosted bytecode -> native compilation. Each phase verified against previous.