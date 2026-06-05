# Zero Bootstrapping Strategy

## Three Phases

### Phase 0: Go Bootstrap (Current)
\\\
go build -> bootstrap.exe
bootstrap.exe compile file.zero -> file.zbc
VM executes file.zbc
\\\

### Phase 1: Self-Hosted Compiler
\\\
bootstrap.exe compile src/compiler/*.zero -> compiler.zbc
compiler.zbc compile file.zero -> file.zbc
\\\
Both compilers produce identical bytecode.

### Phase 2: Native Compilation
\\\
compiler.zbc emits native code instead of bytecode
Zero no longer needs a VM
\\\

## Verification

1. All test .zero files pass on bootstrap.exe
2. All test .zero files pass on compiler.zbc
3. Output bytecode is byte-identical between phases
4. The compiler compiles itself (ultimate verification)

## bootstrap.ps1

Automated pipeline:
1. go build -o build/bootstrap.exe go/main.go
2. bootstrap.exe test tests/*.zero
3. bootstrap.exe build src/compiler/*.zero -o build/compiler.zbc
4. Embed compiler.zbc into VM
5. Verify: zero.exe build src/compiler/*.zero produces byte-identical output
6. Self-hosting achieved.