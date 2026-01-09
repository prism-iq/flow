# Flow Language Reference

**Version:** 1.1.0
**Repo:** https://github.com/prism-iq/flow

---

## Philosophy

Flow reads like English, compiles to C++.

```
hello.flow → Parser → AST → Codegen → hello.cpp → g++ → ./hello
```

**100% local. Zero API dependency. Zero cost.**

---

## Core Principles

1. **Readable > Clever** - Code reads like prose
2. **Explicit > Implicit** - No hidden behavior
3. **Immutable by default** - Use `can change` for mutable
4. **Zero cost abstractions** - Compiles to native C++
5. **One way to do things** - Consistency over flexibility

---

## Quick Reference

### Variables

```flow
name is "Flow"              // const auto name = "Flow"
count is 0, can change      // auto count = 0 (mutable)
count becomes count + 1     // count = count + 1
```

### Functions

```flow
to greet someone:
    say "Hello, {someone}!"

to add a and b:
    return a + b
```

### Entry Point

```flow
to start:
    say "Program begins here"
```

---

## Complete Syntax

### Data Types

| Flow | C++ |
|------|-----|
| `"text"` | `const char*` / `std::string` |
| `42` | `int` |
| `3.14` | `double` |
| `yes` / `no` | `true` / `false` |
| `[1, 2, 3]` | initializer list / vector |

### Type Conversions

| Flow | C++ |
|------|-----|
| `to_int x` | `std::stoi(x)` |
| `to_float x` | `std::stod(x)` |
| `to_string x` | `std::to_string(x)` |
| `length x` | `x.size()` |

### String Operations

| Flow | Description |
|------|-------------|
| `upper text` | Convert to uppercase |
| `lower text` | Convert to lowercase |
| `trim text` | Remove leading/trailing whitespace |
| `split text ","` | Split string into list |
| `join items " "` | Join list into string |
| `contains text "sub"` | Check if substring exists |
| `starts_with text "pre"` | Check prefix |
| `ends_with text "suf"` | Check suffix |
| `replace_all text "old" "new"` | Replace all occurrences |

### Math Functions

| Flow | Description |
|------|-------------|
| `abs x` | Absolute value |
| `min a b` | Minimum of two values |
| `max a b` | Maximum of two values |
| `floor x` | Round down |
| `ceil x` | Round up |
| `round x` | Round to nearest |
| `sqrt x` | Square root |
| `pow x y` | x to the power y |
| `sin x` | Sine |
| `cos x` | Cosine |
| `tan x` | Tangent |
| `log x` | Natural logarithm |
| `log10 x` | Base-10 logarithm |
| `exp x` | e^x |

### List Functions

| Flow | Description |
|------|-------------|
| `sum items` | Sum all elements |
| `product items` | Multiply all elements |
| `sort items` | Sort ascending |
| `reverse items` | Reverse order |
| `unique items` | Remove duplicates |
| `first items` | First element |
| `last items` | Last element |
| `empty items` | Check if empty |

### Random Numbers

| Flow | Description |
|------|-------------|
| `random` | Random float 0.0-1.0 |
| `random 1 100` | Random int in range |

### Time Functions

| Flow | Description |
|------|-------------|
| `now` | Current datetime (YYYY-MM-DD HH:MM:SS) |
| `today` | Current date (YYYY-MM-DD) |
| `clock` | Current time (HH:MM:SS) |

### Filesystem Functions

| Flow | Description |
|------|-------------|
| `exists path` | Check if path exists |
| `isfile path` | Check if path is a file |
| `isdir path` | Check if path is a directory |
| `filesize path` | Get file size in bytes |
| `listdir path` | List directory contents |
| `basename path` | Get filename from path |
| `dirname path` | Get directory from path |
| `extension path` | Get file extension |

### Variables & Assignment

| Flow | C++ |
|------|-----|
| `name is "value"` | `const auto name = "value";` |
| `x is 5, can change` | `auto x = 5;` (mutable) |
| `x becomes 10` | `x = 10;` (reassign) |
| `a, b is func args` | `auto [a, b] = func(args);` |

### Control Flow

```flow
// Conditionals
if condition:
    do_something
otherwise if other:
    do_other
otherwise:
    fallback

// Loops
for each item in list:
    process item

for each i in 1 to 10:
    say i

repeat 5 times:
    say "hello"

while condition:
    keep_going

// Loop control
skip        // continue
stop        // break
```

### Functions & Returns

```flow
// Simple function
to greet name:
    say "Hello, {name}!"

// With return
to square x:
    return x * x

// Multiple returns
to minmax a and b:
    if a < b:
        return a and b
    otherwise:
        return b and a

// Usage with unpacking
min, max is minmax 5 and 3
```

### Structs & Methods

```flow
// Define struct
a Person has:
    name as text
    age as number

// Define method
a Person can greet:
    say "Hi, I'm {my name}"

// Usage
bob is Person "Bob" 25
bob's greet
```

### Collections

```flow
// Lists
items is [1, 2, 3, 4, 5]
first is items at 0

// List comprehension
squares is [x * x for each x in 1 to 5]
evens is [x for each x in 1 to 10 where x % 2 == 0]

// Slicing
middle is items from 1 to 4
tail is items from 2
head is items to 3

// Piping
result is 5 | double | square
```

### I/O Operations

```flow
// Console
say "Hello World"               // with newline
print "Loading..."              // without newline
name is ask "Enter name: "
input is ask                    // no prompt

// Timing
pause 1000                      // sleep 1 second (milliseconds)

// Files
content is read "file.txt"
write "data" to "output.txt"
append "more" to "output.txt"

// Environment
home is env "HOME"
path is env "PATH"

// Shell commands
output is run "ls -la"
```

### Context Managers

```flow
using file is open "/tmp/test.txt":
    say "File auto-closes at block end"
```

### Generators

```flow
to count_up limit:
    for each i in 1 to limit:
        yield i

for each n in count_up 5:
    say n    // 1, 2, 3, 4, 5
```

### Decorators

```flow
to doubled value:
    return value * 2

@doubled
to get_value x:
    return x + 1

// get_value 5 returns 12 (not 6)
```

---

## Advanced Features

### HTTP Requests

```flow
response is fetch "http://api.example.com/data"
say response
```

### Regex Operations

```flow
// Test match
if match "[0-9]+" in text:
    say "Found numbers"

// Find all matches
matches is find "[a-z]+" in text

// Replace
cleaned is replace "[0-9]+" in text with "X"
```

### Cryptographic Hashing

```flow
sha is hash sha256 "secret"
md is hash md5 "secret"
sha1 is hash sha1 "secret"
```

### Concurrent Execution

```flow
do together:
    task_a
    task_b
    task_c
// All run in parallel, waits for all to complete
```

### Async/Await

```flow
result is wait fetch "http://slow.api/data"
```

### WebSockets

```flow
socket is connect "ws://server:8080/ws"
send "Hello" to socket
```

### Logging

```flow
log info "Application started"
log warn "Low memory"
log error "Connection failed"
// Outputs: 2024-01-15 10:30:45 [INFO] Application started
```

### Testing

```flow
test "addition works":
    result is add 2 and 3
    assert result == 5, "2 + 3 should be 5"
```

### Error Handling

```flow
try:
    result is risky_operation
catch e:
    say "Error occurred"

throw "Something went wrong"
```

---

## Operators

### Arithmetic
| Flow | C++ |
|------|-----|
| `a + b` | `a + b` |
| `a - b` | `a - b` |
| `a * b` | `a * b` |
| `a / b` | `a / b` |
| `a % b` | `a % b` |

### Comparison
| Flow | C++ |
|------|-----|
| `a < b` | `a < b` |
| `a > b` | `a > b` |
| `a <= b` | `a <= b` |
| `a >= b` | `a >= b` |
| `a == b` | `a == b` |
| `a != b` | `a != b` |
| `a is b` | `a == b` |

### Logical
| Flow | C++ |
|------|-----|
| `a and b` | `a && b` |
| `a or b` | `a \|\| b` |
| `not a` | `!a` |

### Special
| Flow | C++ |
|------|-----|
| `value \| func` | `func(value)` |
| `obj's field` | `obj.field` |
| `my field` | `this->field` |
| `list at 0` | `list[0]` |
| `{expr}` | string interpolation |

---

## CLI Commands

```bash
flow run hello.flow       # Parse + compile + run
flow build hello.flow     # Parse + compile (creates binary)
flow show hello.flow      # Show generated C++ code

# Options
flow run file.flow --keep    # Keep .cpp file
flow run file.flow --debug   # Show debug output
```

---

## Project Structure

```
/opt/flow/
├── cmd/flow/main.go          # CLI entry
├── internal/
│   ├── lexer/lexer.go        # Tokenizer
│   ├── parser/parser.go      # AST builder
│   ├── codegen/codegen.go    # C++ generator
│   └── compiler/compiler.go  # g++ wrapper
├── examples/                  # .flow examples
├── docs/                      # Documentation
└── scripts/                   # Build/test scripts
```

---

## Examples

### Hello World
```flow
to start:
    say "Hello, World!"
```

### FizzBuzz
```flow
to start:
    for each i in 1 to 100:
        if i % 15 == 0:
            say "FizzBuzz"
        otherwise if i % 3 == 0:
            say "Fizz"
        otherwise if i % 5 == 0:
            say "Buzz"
        otherwise:
            say i
```

### Fibonacci
```flow
to fib n:
    if n <= 1:
        return n
    return fib (n - 1) + fib (n - 2)

to start:
    for each i in 0 to 10:
        say fib i
```

### File Processing
```flow
to start:
    content is read "/etc/hostname"
    lines is split content "\n"
    for each line in lines:
        if length line > 0:
            say line
```

### HTTP + JSON
```flow
to start:
    response is fetch "http://httpbin.org/get"
    say response
```

### Forensic Analysis
```flow
to start:
    log info "Starting analysis"

    content is read "evidence.txt"

    // Find all emails
    emails is find "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}" in content

    for each email in emails:
        say email

    // Hash the evidence
    hash is hash sha256 content
    log info "Evidence hash: computed"
    say hash
```

---

## Compilation

Flow generates standard C++17 code that compiles with:
- g++ or clang++
- Links: pthread, ssl, crypto
- Headers: iostream, string, vector, regex, thread, future, etc.

---

## Security Notes

**Never commit:**
- API keys, tokens, secrets
- SSH keys
- Passwords
- .env files

**Pre-commit check:**
```bash
git diff --staged | grep -iE "(password|secret|token|key)" && echo "WARNING" || echo "Clean"
```
