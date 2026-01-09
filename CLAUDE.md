# Claude Context - Flow Programming Language

**Version:** 0.4.0 (Full Local - No API)
**Repository:** https://github.com/prism-iq/flow

---

## What is Flow?

Flow reads like English, compiles to C++.

```
hello.flow → [Claude API] → hello.cpp → [g++] → ./hello
```

**The AI is the transpiler.**

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

---

## Transpilation Examples

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
│   ├── transpiler/           # Claude API integration
│   ├── compiler/             # g++ wrapper + feedback loop
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
flow run hello.flow      # Transpile + compile + run
flow build hello.flow    # Transpile + compile (creates binary)
flow show hello.flow     # Show generated C++
```

---

## TL;DR for Transpilation

Flow v0.3 uses natural English:
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

**Output ONLY valid C++17. No markdown. Just code.**

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
