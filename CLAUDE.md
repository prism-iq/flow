# Claude Context - Flow Programming Language

**Version:** 0.8.0 (Full Local - No API)
**Repository:** https://github.com/prism-iq/flow

---

## What is Flow?

Flow reads like English, compiles to C++.

```
hello.flow → [Parser] → [AST] → [Codegen] → hello.cpp → [g++] → ./hello
```

**100% local. Zero API dependency.**

---

## Axioms

1. **Explicit > Implicit** - No hidden behavior
2. **Errors are values** - No exceptions
3. **Null doesn't exist** - Use `maybe`
4. **Immutable by default** - `can change` for mutable
5. **No hidden allocations** - Stack default
6. **One way to do things** - No overloading
7. **Composition > Inheritance** - Embed, don't extend
8. **Zero cost abstractions** - Compiles to concrete types
9. **Fail at compile time** - Strong typing
10. **Readable > Clever** - Code reads like prose

---

## Flow Vocabulary → C++

| Flow | C++ |
|------|-----|
| `name is "x"` | `const auto name = "x";` |
| `x becomes 5` | `x = 5;` |
| `x, can change` | mutable variable |
| `to greet someone:` | `void greet(auto someone) {` |
| `return x` | `return x;` |
| `a Person has:` | `struct Person {` |
| `a Person can greet:` | method definition |
| `my name` | `this->name` |
| `bob's name` | `bob.name` |
| `if/otherwise` | `if/else` |
| `for each x in list:` | `for (auto x : list) {` |
| `repeat 5 times:` | `for (int i=0; i<5; i++) {` |
| `while x < 10:` | `while (x < 10) {` |
| `skip` | `continue;` |
| `stop` | `break;` |
| `and/or/not` | `&&/\|\|/!` |
| `yes/no` | `true/false` |
| `say "text"` | `std::cout << "text" << std::endl;` |
| `{expr}` in string | string interpolation |
| `numbers at 0` | `numbers[0]` |
| `try x` | result/error handling |
| `maybe x` | `std::optional` |
| `wait fetch url` | async/await |
| `do together:` | concurrent execution |
| `on heap` | heap allocation |
| `[x*2 for each x in items]` | list comprehension |
| `[x for each x in 1 to 10 where x > 5]` | filtered comprehension |
| `value \| func` | piping (func(value)) |
| `read "file.txt"` | read file content |
| `write content to "file.txt"` | write to file |
| `append content to "file.txt"` | append to file |
| `env "HOME"` | get environment variable |
| `run "ls -la"` | execute shell command |
| `return a and b` | return multiple values |
| `x, y is func args` | unpacking assignment |
| `using x is open "f":` | context manager (RAII) |
| `items from 1 to 5` | list slicing |
| `items from 2` | slice to end |
| `items to 3` | slice from start |

---

## Compilation Examples

### Variables

```flow
name is "Flow"
age is 25
count is 0, can change
count becomes count + 1
```
→
```cpp
const auto name = std::string("Flow");
const auto age = 25;
auto count = 0;
count = count + 1;
```

### Functions

```flow
to greet someone:
    say "Hello, {someone}!"

to add a and b:
    return a + b
```
→
```cpp
void greet(const std::string& someone) {
    std::cout << "Hello, " << someone << "!" << std::endl;
}

auto add(auto a, auto b) {
    return a + b;
}
```

### Structs

```flow
a Person has:
    name as text
    age as number

a Person can introduce:
    say "I'm {my name}, {my age} years old"
```
→
```cpp
struct Person {
    std::string name;
    int age;

    void introduce() const {
        std::cout << "I'm " << name << ", " << age << " years old" << std::endl;
    }
};
```

### Loops

```flow
for each item in items:
    say item

repeat 5 times:
    say "hello"

for each i in 1 to 10:
    say i
```
→
```cpp
for (const auto& item : items) {
    std::cout << item << std::endl;
}

for (int i = 0; i < 5; i++) {
    std::cout << "hello" << std::endl;
}

for (int i = 1; i <= 10; i++) {
    std::cout << i << std::endl;
}
```

### Conditions

```flow
if age >= 18:
    say "Adult"
otherwise:
    say "Minor"
```
→
```cpp
if (age >= 18) {
    std::cout << "Adult" << std::endl;
} else {
    std::cout << "Minor" << std::endl;
}
```

### List Comprehensions

```flow
squares is [x * x for each x in 1 to 5]
evens is [x for each x in 1 to 10 where x % 2 == 0]
```
→
```cpp
auto squares = /* lambda building vector */;  // [1, 4, 9, 16, 25]
auto evens = /* lambda with filter */;        // [2, 4, 6, 8, 10]
```

### Piping

```flow
result is 5 | twice | square
```
→
```cpp
auto result = square(twice(5));
```

### File I/O

```flow
write "Hello" to "output.txt"
append "\nWorld" to "output.txt"
content is read "output.txt"
```
→
```cpp
{ std::ofstream _f("output.txt"); _f << "Hello"; }
{ std::ofstream _f("output.txt", std::ios::app); _f << "\nWorld"; }
auto content = /* lambda reading file */;
```

### Environment Variables

```flow
home is env "HOME"
user is env "USER"
```
→
```cpp
auto home = std::getenv("HOME");
auto user = std::getenv("USER");
```

### Run Command

```flow
output is run "ls -la"
say output
```
→
```cpp
auto output = /* popen execution */;
```

### Multiple Returns & Unpacking

```flow
to get_minmax a and b:
    if a < b:
        return a and b
    otherwise:
        return b and a

min, max is get_minmax 5 and 3
```
→
```cpp
auto get_minmax(auto a, auto b) {
    if (a < b) return std::make_tuple(a, b);
    else return std::make_tuple(b, a);
}
const auto [min, max] = get_minmax(5, 3);
```

### Context Managers

```flow
using file is open "/tmp/test.txt":
    say "File open, auto-closes at end"
```
→
```cpp
{ auto file = std::fstream("/tmp/test.txt");
    std::cout << "File open..." << std::endl;
}  // file auto-closed
```

### Slicing

```flow
items is [1, 2, 3, 4, 5]
middle is items from 1 to 4
tail is items from 2
head is items to 3
```
→
```cpp
auto middle = std::vector(items.begin() + 1, items.begin() + 4);
auto tail = std::vector(items.begin() + 2, items.end());
auto head = std::vector(items.begin(), items.begin() + 3);
```

---

## Entry Point

```flow
to start:
    say "Hello, World!"
```
→
```cpp
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}
```

---

## File Structure

```
/opt/flow/
├── cmd/flow/main.go          # CLI entry point
├── internal/
│   ├── cli/cli.go            # Commands: run, build, show
│   ├── lexer/lexer.go        # Tokenizer with indent tracking
│   ├── parser/parser.go      # Recursive descent parser
│   ├── codegen/codegen.go    # AST → C++ code generation
│   ├── compiler/             # g++ wrapper
│   └── config/               # Environment config
├── docs/
│   ├── SYNTAX.md             # Full language reference
│   └── PRINCIPLES.md         # Design decisions
├── examples/                 # .flow examples
├── CLAUDE.md                 # This file
└── go.mod
```

---

## Commands

```bash
flow run hello.flow      # Parse + compile + run
flow build hello.flow    # Parse + compile (creates binary)
flow show hello.flow     # Show generated C++
```

---

## TL;DR for Flow Syntax

Flow v0.8 uses natural English:
- `name is "x"` = assignment
- `x becomes y` = reassignment
- `to do something:` = function
- `a Thing has:` = struct
- `a Thing can act:` = method
- `my field` = self.field
- `x's field` = x.field
- `for each/repeat/while` = loops
- `if/otherwise` = conditionals
- `yes/no` = true/false
- `and/or/not` = &&/||/!
- `say` = print
- `{x}` = interpolation
- `[expr for each x in items]` = list comprehension
- `value | func` = piping
- `read/write/append` = file I/O
- `env "VAR"` = environment variable
- `run "cmd"` = shell command
- `return a and b` = multiple returns
- `x, y is func` = unpacking
- `using x is open:` = context manager
- `items from 1 to 5` = slicing

**Generates valid C++20.**

---

## GITHUB & SECURITY

**Repo:** https://github.com/prism-iq/flow

### NEVER COMMIT:
- API keys, tokens, secrets
- SSH keys (~/.ssh/*)
- Passwords
- .env files

### Before Committing
```bash
git diff --staged | grep -iE "(sk-ant|password|secret|token|ssh|BEGIN.*KEY)" && echo "SECRET DETECTED" || echo "Clean"
```

### Auto-Commit Protocol
```bash
cd /opt/flow && git add -A && git commit -m "[type]: description" && git push
```
Types: feat, fix, docs, refactor, chore
