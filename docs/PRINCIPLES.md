# Architecture Principles - Flow Language

**Version:** 0.1.0
**Last Updated:** 2026-01-08

---

## Core Philosophy

```
"Simple syntax, powerful output."
"The AI is the transpiler, not the feature."
```

---

## Principle 1: AI-First Transpilation

**Decision:** Use Claude API as the primary transpilation engine, not traditional parsing.

**Rationale:**
- Traditional compilers need hundreds of transformation rules
- AI can infer intent from context
- Handles edge cases gracefully
- Can be "taught" new patterns via prompt engineering

**Trade-offs:**
- Requires API key (cost)
- Non-deterministic (same input might produce slightly different output)
- Network latency

**Mitigation:**
- Cache successful transpilations
- Fallback to cached patterns when offline
- Deterministic seed in API calls

---

## Principle 2: Feedback Loop

**Decision:** When C++ compilation fails, automatically send error back to Claude for fixing.

**Rationale:**
- AI sometimes produces invalid C++
- Error messages contain the fix hint
- Faster than human debugging
- Self-correcting system

**Implementation:**
```
Flow → Claude → C++ → g++
                  ↓
              ERROR?
                  ↓
         Claude (with error) → Fixed C++ → g++ (retry)
         (max 3 attempts)
```

**Limits:**
- Max 3 retry attempts
- Total timeout: 30 seconds
- Give up and show error if still failing

---

## Principle 3: Self-Documenting

**Decision:** The transpiler reads its own `.md` files for context.

**Rationale:**
- Documentation stays in sync with behavior
- AI can use docs as reference
- Single source of truth
- Easy to update behavior by editing docs

**Implementation:**
```go
// Load syntax reference for Claude context
syntaxDoc, _ := os.ReadFile("docs/SYNTAX.md")
prompt := fmt.Sprintf(`
Reference documentation:
%s

Transpile this Flow code to C++:
%s
`, syntaxDoc, flowCode)
```

---

## Principle 4: Zero Runtime

**Decision:** Flow compiles to pure C++ with no runtime dependencies.

**Rationale:**
- Maximum performance
- No "Flow runtime" to install
- Output is portable standard C++
- Can be compiled anywhere with g++/clang++

**Implementation:**
- All Flow constructs map to C++ stdlib
- `print()` → `std::cout`
- `vec<T>` → `std::vector<T>`
- No custom headers required

---

## Principle 5: Progressive Complexity

**Decision:** Simple things should be simple, complex things should be possible.

**Rationale:**
- Hello World should be 2 lines
- But full programs should also be writable
- Don't force boilerplate for simple cases

**Examples:**
```flow
// Simplest possible
fn main():
    print("Hello")

// Full complexity when needed
struct Server:
    port: int
    routes: map<str, fn(Request) -> Response>

    fn listen(self):
        // Complex implementation
```

---

## Principle 6: Python-Like Syntax

**Decision:** Use indentation-based syntax inspired by Python.

**Rationale:**
- Familiar to most developers
- Reduces visual noise (no `{}`)
- Forces readable code
- Simpler parsing rules

**Non-goals:**
- Not trying to be Python
- Not Python-compatible
- Just visually similar

---

## Principle 7: Type Inference with Optional Annotations

**Decision:** Types are inferred by default but can be explicitly declared.

**Rationale:**
- Less boilerplate for simple code
- Explicit types for clarity when needed
- AI can infer types from context
- C++ `auto` makes this natural

**Examples:**
```flow
let x = 42              // AI infers: int
let name = "Flow"       // AI infers: str
let data: vec<int> = [] // Explicit for empty collections
```

---

## Principle 8: Fail Fast, Fail Clear

**Decision:** Provide clear error messages with suggestions.

**Rationale:**
- Cryptic g++ errors are frustrating
- AI can translate errors to human-readable form
- Show the Flow line that caused the issue
- Suggest fixes when possible

**Implementation:**
```
Error in hello.flow:5

    5 |     prnt("Hello")
            ^^^^
    Unknown function 'prnt'. Did you mean 'print'?
```

---

## Principle 9: Cache Everything

**Decision:** Cache successful transpilations to avoid redundant API calls.

**Rationale:**
- Save money (API costs)
- Faster recompilation
- Work offline with cached code
- Deterministic rebuilds

**Implementation:**
```json
// cache/patterns.json
{
  "sha256_of_flow_code": {
    "cpp": "...",
    "timestamp": "...",
    "flow_version": "0.1.0"
  }
}
```

**Invalidation:**
- Flow version changes
- Manual `flow clean`
- Cache older than 30 days

---

## Principle 10: Go for Tooling

**Decision:** Use Go for the CLI and all tooling.

**Rationale:**
- Single binary distribution
- Fast compilation
- Excellent stdlib (HTTP, JSON, CLI)
- Cross-platform
- Good concurrency for parallel compilation

**Stack:**
- `cobra` for CLI
- `net/http` for API calls
- `os/exec` for g++ invocation
- Standard `encoding/json` for cache

---

## Design Decisions

### Why Not Traditional Parsing?

Traditional approach:
```
Lexer → Parser → AST → Semantic Analysis → IR → Code Gen
```

Flow approach:
```
Source → Context Builder → Claude → C++
```

**Benefits:**
- 10x less code to maintain
- Handles ambiguity naturally
- Easy to add new features (just update prompt)
- AI improves over time

**Risks:**
- API dependency
- Cost per compilation
- Non-determinism

---

### Why C++ as Target?

Options considered:
1. **C** - Too low-level, no stdlib
2. **C++** - Good balance, modern features, universal
3. **LLVM IR** - Too complex, loses readability
4. **Rust** - Borrow checker complexity

C++ wins because:
- Universal availability (g++, clang++)
- Modern features (C++17/20)
- Good performance
- Readable output (can debug generated code)

---

### Why Not Just Use Python?

"If Flow looks like Python, why not just use Python?"

| Aspect | Python | Flow |
|--------|--------|------|
| Execution | Interpreted | Compiled native |
| Performance | Slow | Fast (C++ speed) |
| Distribution | Needs Python installed | Single binary |
| Type safety | Runtime errors | Compile-time errors |
| Dependencies | pip, venv, etc. | None |

Flow = Python ergonomics + C++ performance

---

## Anti-Patterns Avoided

### No Build System Complexity

**Avoided:** CMake, Makefiles, Bazel, etc.

**Why:** Flow compiles to a single C++ file. Just run g++.

**Reality:**
```bash
# This is the entire "build system"
g++ -std=c++17 -o output generated.cpp
```

---

### No Package Manager (Yet)

**Avoided:** npm-style dependency hell

**Why:** Start simple. Add packages when there's actual need.

**Future:** Maybe a simple `flow.toml` with dependencies

---

### No Complex Type System

**Avoided:** Generics, traits, lifetimes, etc.

**Why:** Let C++ handle the complexity. Flow just describes intent.

**Reality:** AI maps Flow types to appropriate C++ types

---

## Success Metrics

### v0.1 (MVP)
- [ ] hello.flow compiles and runs
- [ ] Feedback loop fixes at least 50% of errors
- [ ] < 5 second compile time (with cache)

### v0.5
- [ ] 90% of example programs compile first try
- [ ] Cache hit rate > 80%
- [ ] < 2 second compile time (with cache)

### v1.0
- [ ] Full syntax support (structs, enums, generics)
- [ ] Offline mode with cached patterns
- [ ] Editor support (LSP)
- [ ] Package system

---

## Cost Analysis

### API Costs (Claude)

Assumptions:
- Average Flow file: 500 tokens
- Average C++ output: 1000 tokens
- Haiku pricing: $0.25/1M input, $1.25/1M output

Per compilation:
- Input: 500 tokens × $0.25/1M = $0.000125
- Output: 1000 tokens × $1.25/1M = $0.00125
- **Total: ~$0.0014 per compilation** (0.14 cents)

With caching (80% hit rate):
- **Effective cost: ~$0.0003 per compilation** (0.03 cents)

Budget 30€/month = ~100,000 compilations

---

## Future Considerations

### LSP (Language Server Protocol)
- Syntax highlighting
- Autocomplete (AI-powered)
- Inline errors
- Hover documentation

### Playground
- Web-based Flow editor
- Live transpilation preview
- Shareable examples

### Package Manager
- `flow.toml` for dependencies
- Central package registry
- Vendoring support

### Multiple Backends
- C++ (current)
- Rust (future)
- Go (future)
- WASM (future)

---

**TL;DR:**

AI IS the transpiler. Feedback loop auto-fixes errors. Self-documenting via .md files. Zero runtime (pure C++). Python-like syntax. Type inference. Cache everything. Go for tooling. Simple things simple, complex things possible.
