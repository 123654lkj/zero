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

`ash
# Clone
git clone https://github.com/yourusername/zero.git
cd zero

# Build bootstrap compiler
./bootstrap.ps1

# Run your first Zero program
zero run hello.zero
`

`zero
// hello.zero
print("Hello from Zero!")
`

---

## Language at a Glance

`zero
// No null. No nil. No None.
// Option<T> and Result<T, E> are the only ways to express absence.

// Algebraic types with pattern matching
enum FetchResult {
    Ok(data: string, latency: float)
    Blocked(reason: string, cooldown: int)
    Captcha(image: bytes)
    Error(message: string)
}

match result {
    Ok(data, _) if data.len() > 0 -> process(data)
    Ok(_, _) -> retry()
    Blocked(_, secs) -> wait(secs) |> retry()
    Captcha(img) -> ai.solve(img)
    Error(_) -> escalate()
}

// AI as a first-class language construct
ai fn extract_price(html: string) -> float {
    "Extract the product price from this HTML. Return only the number."
}

// Declarative data pipelines
let results = from "https://jd.com"
            |> search "iPhone 16"
            |> extract price, title, stock
            |> filter price < 5000
            |> save "./data/prices.json"

// Browser interaction as native I/O
fn get_price(product_id: string) -> float {
    let browser = web.launch({ headless: true })
    let page = browser.open("https://item.jd.com/{product_id}.html")
    wait page.selector(".sku-price") timeout: 10s
    let price = page.text(".sku-price")?.clean_price()
    return price
}

// Effects tracking
fn pure_compute(x: int) -> int  // Pure
fn read_config() -> string !IO   // Has IO effect
fn fetch(url: string) -> string !NetIO !Error  // Network + errors
`

---

## Key Innovations

### 1. No Null
Zero has no null. Every nullable value uses Option<T>. No null pointer exceptions. Ever.

### 2. Errors Are Not Exceptions
Errors are values. Every fallible function returns Result<T, E>. The ! operator propagates errors.

### 3. AI-Native
ai fn is a language construct, not a library call. The compiler understands AI semantics.

### 4. Effects Tracking
The compiler tracks side effects through the type system. Pure functions guaranteed pure.

### 5. Image Snapshots
The runtime can snapshot its entire state and restore later. Debugging at the language level.

### 6. Actor Model
Lightweight actors with spawn/send/receive as core primitives, from Erlang.

### 7. Pipes as First-Class
The |> operator is compiler-optimized and auto-parallelizes where possible.

---

## Bootstrapping Strategy

| Phase | Compiler In | Output | Status |
|-------|-------------|--------|--------|
| 0 | Go | Bytecode (.zbc) | Building |
| 1 | Zero (self-hosted) | Bytecode (.zbc) | Planned |
| 2 | Zero | Native binary | Long-term |

---

## Project Structure

`
zero/
├── go/              # Bootstrap compiler (Go)
│   ├── main.go      # Entry point
│   ├── value/       # Value system
│   ├── opcode/      # Instruction set
│   ├── chunk/       # Bytecode container
│   ├── compiler/    # Lexer, Parser, Compiler
│   ├── vm/          # Bytecode interpreter
│   └── repl/        # Interactive REPL
├── src/             # Self-hosted compiler (Zero)
├── tests/           # Test programs
├── doc/             # Documentation
├── bootstrap.ps1    # Build and self-hosting script
└── README.md
`

---

## License

Zero is open source. License TBD.
