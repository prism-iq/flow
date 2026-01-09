# Claude Context - Flow Programming Language

**Last Updated:** 2026-01-09
**Version:** 0.2.0
**Repository:** /opt/flow

---

## What is Flow?

Flow is a programming language that combines:
- **Power of C++** - Compiles to native code
- **Clarity of Python** - Clean, readable syntax
- **Natural vocabulary** - English words, not jargon

```
hello.flow → [Claude API] → hello.cpp → [g++] → ./hello
```

**The AI is the transpiler.**

---

## Flow Vocabulary

| Flow | Traditional | C++ Translation |
|------|-------------|-----------------|
| `do` | fn/func/def | function |
| `thing` | struct/class | struct |
| `kind` | enum | enum class |
| `when` | match/switch | switch or if-else chain |
| `need` | import/use | #include |
| `give` | return | return |
| `say` | print | std::cout |
| `each` | for | for loop |
| `vary` | mut/var | non-const |
| `me` | self/this | this-> |
| `yes`/`no` | true/false | true/false |
| `and`/`or`/`not` | &&/\|\|/! | &&/\|\|/! |
| `nothing` | nil/null | std::nullopt |
| `good`/`fail` | Ok/Err | success/error |
| `start` | spawn/async | std::async |
| `wait` | await | .get() |
| `later` | defer | RAII/destructor |
| `pipe` | channel | queue |
| `skip`/`stop` | continue/break | continue/break |
| `check` | guard | if (!cond) return |

---

## Transpilation Rules

### Functions

```flow
do greet(name):
    say "Hello, #{name}"
```
→
```cpp
void greet(std::string name) {
    std::cout << "Hello, " << name << std::endl;
}
```

```flow
do add(a, b):
    give a + b
```
→
```cpp
auto add(auto a, auto b) {
    return a + b;
}
```

```flow
do double(x) = x * 2
```
→
```cpp
auto double_fn(auto x) { return x * 2; }
```

### Variables

```flow
name = "Flow"
vary count = 0
count = count + 1
```
→
```cpp
const auto name = std::string("Flow");
auto count = 0;
count = count + 1;
```

### Things (Structs)

```flow
thing Point:
    x: int
    y: int

    do show(me) = "(#{me.x}, #{me.y})"
```
→
```cpp
struct Point {
    int x;
    int y;

    std::string show() const {
        return "(" + std::to_string(x) + ", " + std::to_string(y) + ")";
    }
};
```

### Loops

```flow
each i in 1..10:
    say i
```
→
```cpp
for (int i = 1; i < 10; i++) {
    std::cout << i << std::endl;
}
```

```flow
each item in items:
    say item
```
→
```cpp
for (const auto& item : items) {
    std::cout << item << std::endl;
}
```

### Conditions

```flow
if age >= 18:
    say "adult"
else:
    say "minor"
```
→
```cpp
if (age >= 18) {
    std::cout << "adult" << std::endl;
} else {
    std::cout << "minor" << std::endl;
}
```

### Pattern Matching

```flow
when value:
    0 => say "zero"
    1 => say "one"
    _ => say "other"
```
→
```cpp
switch (value) {
    case 0: std::cout << "zero" << std::endl; break;
    case 1: std::cout << "one" << std::endl; break;
    default: std::cout << "other" << std::endl; break;
}
```

### String Interpolation

```flow
say "Hello, #{name}! You are #{age} years old."
```
→
```cpp
std::cout << "Hello, " << name << "! You are " << age << " years old." << std::endl;
```

### Booleans & Logic

```flow
if active and not banned:
    allow()

enabled = yes
disabled = no
```
→
```cpp
if (active && !banned) {
    allow();
}

auto enabled = true;
auto disabled = false;
```

---

## File Structure

```
/opt/flow/
├── cmd/flow/main.go          # CLI entry point
├── internal/
│   ├── cli/cli.go            # Commands: run, build, show
│   ├── transpiler/           # Claude API integration
│   ├── compiler/             # g++ wrapper + feedback loop
│   └── config/               # Environment config
├── docs/
│   ├── SYNTAX.md             # Full language reference
│   └── PRINCIPLES.md         # Design decisions
├── examples/
│   ├── hello.flow            # Hello world
│   ├── fibonacci.flow        # Recursion
│   ├── loop.flow             # Loops demo
│   └── things.flow           # Structs demo
├── CLAUDE.md                 # This file
└── go.mod                    # Go module
```

---

## Commands

```bash
flow run hello.flow      # Transpile + compile + run
flow build hello.flow    # Transpile + compile (creates binary)
flow show hello.flow     # Show generated C++
```

---

## Environment Variables

```bash
ANTHROPIC_API_KEY=sk-ant-...    # Required
FLOW_COMPILER=g++               # Default: g++
FLOW_STD=c++17                  # Default: c++17
FLOW_DEBUG=false                # Show debug output
```

---

## System Prompt for Transpilation

When transpiling Flow to C++, Claude should:

1. **Output ONLY valid C++ code** - no markdown, no explanations
2. **Use modern C++17** features
3. **Include necessary headers** - iostream, string, vector, etc.
4. **Translate Flow vocabulary:**
   - `do` → function
   - `thing` → struct
   - `say` → std::cout <<
   - `each` → for loop
   - `give` → return
   - `vary` → mutable variable
   - `yes`/`no` → true/false
   - `and`/`or`/`not` → &&/||/!
   - `#{expr}` → string concatenation
5. **Handle indentation** - Flow uses Python-style blocks
6. **Return 0 from main()**

---

## Examples

### hello.flow
```flow
do main:
    say "Hello, World!"
```

### fibonacci.flow
```flow
do fib(0) = 0
do fib(1) = 1
do fib(n) = fib(n-1) + fib(n-2)

do main:
    each i in 0..10:
        say fib(i)
```

### things.flow
```flow
thing Person:
    name: str
    age: int

    do greet(me):
        say "Hi, I'm #{me.name}"

    do adult?(me) = me.age >= 18

do main:
    bob = Person("Bob", 30)
    bob.greet()

    if bob.adult?():
        say "Bob is an adult"
```

---

## Feedback Loop

When g++ fails, the error is sent back to Claude:

```
Compile Error: [error message]
Original Flow: [source]
Generated C++: [code]

Fix the C++ to resolve this error.
```

Max 3 retries before giving up.

---

## TL;DR for Claude

Flow uses natural English words:
- `do` = function
- `thing` = struct
- `say` = print
- `each` = for
- `give` = return
- `when` = switch/match
- `yes`/`no` = true/false
- `and`/`or`/`not` = &&/||/!
- `vary` = mutable
- `#{x}` = string interpolation

Transpile to clean C++17. No markdown. Just code.
