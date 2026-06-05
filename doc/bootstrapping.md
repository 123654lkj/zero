# Zero Bootstrapping Strategy

## The Self-Hosting Pipeline

### Phase 0 (Current): Go Bootstrap
```
go build -> bootstrap.exe
bootstrap.exe compiles file.zero -> file.zbc
VM executes file.zbc
```

### Phase 1: Self-Hosted Compiler
```
bootstrap.exe compiles src/compiler/*.zero -> compiler.zbc
compiler.zbc compiles file.zero -> file.zbc
```
Result: compiler.zbc and bootstrap.exe output identical bytecode.

### Phase 2: Native Compilation
```
compiler.zbc learns to emit native machine code
Zero no longer needs a VM
```
Result: truly native Zero binaries.

## Verification

At each phase, Zero is verified by running the full test suite:

1. All test .zero files pass on bootstrap.exe
2. All test .zero files pass on compiler.zbc
3. Output bytecode is byte-identical between phases
4. The compiler can compile itself (the ultimate verification)

## The Bootstrap Script

`bootstrap.ps1` automates the entire pipeline:

1. Build Go bootstrap: `go build -o build/bootstrap.exe go/main.go`
2. Run tests: `bootstrap.exe test tests/*.zero`
3. Compile self-hosted compiler: `bootstrap.exe build src/compiler/*.zero -o build/compiler.zbc`
4. Embed compiler.zbc into VM
5. Verify: `zero.exe build src/compiler/*.zero` produces identical output
6. Self-hosting achieved

## Why Bootstrap?

Writing a compiler in its own language is the ultimate test of that language's completeness. If Zero can compile Zero, then:

- The language is expressive enough to write complex programs
- The compiler is correct (it can process itself)
- The runtime is complete (it can run the compiler)
- Zero has achieved independence from Go