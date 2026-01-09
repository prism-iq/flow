# Claude Context - Flow Programming Language

**Version:** 1.0.0 (Full Local - No API)
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
| `ask "prompt"` | read line from stdin |
| `ask` | read line from stdin (no prompt) |
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
| `yield x` | generator yield |
| `@decorator` | function decorator |
| `fetch "url"` | HTTP GET request |
| `parse json` | parse JSON string |
| `stringify value` | convert to string |
| `match "pattern" in text` | regex match (bool) |
| `find "pattern" in text` | regex find all |
| `replace "pattern" in text with "new"` | regex replace |
| `hash sha256 value` | SHA256 hash |
| `hash md5 value` | MD5 hash |
| `hash sha1 value` | SHA1 hash |
| `wait expr` | async await |
| `do together:` | concurrent block |
| `connect "ws://..."` | WebSocket connect |
| `send msg to socket` | WebSocket send |
| `log info "msg"` | info logging |
| `log warn "msg"` | warning logging |
| `log error "msg"` | error logging |
| `test "name":` | test block |
| `assert cond, "msg"` | assertion |
| `try: ... catch e:` | error handling |
| `throw "error"` | throw exception |
| `to_int x` | `std::stoi(x)` |
| `to_float x` | `std::stod(x)` |
| `to_string x` | `std::to_string(x)` |
| `length x` | `x.size()` |

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

### Generators

```flow
to count_up limit:
    for each i in 1 to limit:
        yield i

for each n in count_up 5:
    say n
```
→
```cpp
auto count_up(auto limit) {
    std::vector<int> _result;
    auto _yield = [&](auto v) { _result.push_back(v); };
    for (int i = 1; i <= limit; i++) { _yield(i); }
    return _result;
}
```

### Decorators

```flow
to doubled value:
    return value * 2

@doubled
to get_value x:
    return x + 1
```
→
```cpp
auto doubled(auto value) { return value * 2; }
auto _get_value_impl(auto x) { return x + 1; }
auto get_value(auto x) { return doubled(_get_value_impl(x)); }
```

### HTTP Client

```flow
response is fetch "http://api.example.com/data"
say response
```
→
```cpp
// Uses POSIX sockets for HTTP GET request
auto response = /* socket-based HTTP fetch */;
```

### Regex

```flow
if match "[0-9]+" in text:
    say "Found numbers"

matches is find "[a-z]+" in text
cleaned is replace "[0-9]+" in text with "X"
```
→
```cpp
if (std::regex_search(text, std::regex("[0-9]+"))) { ... }
auto matches = /* sregex_iterator loop */;
auto cleaned = std::regex_replace(text, std::regex("[0-9]+"), "X");
```

### Hashing

```flow
sha is hash sha256 "hello"
md is hash md5 "hello"
```
→
```cpp
// Uses OpenSSL
auto sha = /* SHA256 computation */;
auto md = /* MD5 computation */;
```

### Concurrent Execution

```flow
do together:
    task1
    task2
    task3
```
→
```cpp
{
    std::vector<std::thread> _threads;
    _threads.emplace_back([&]() { task1; });
    _threads.emplace_back([&]() { task2; });
    _threads.emplace_back([&]() { task3; });
    for (auto& t : _threads) t.join();
}
```

### Logging

```flow
log info "Application started"
log warn "Low memory"
log error "Connection failed"
```
→
```cpp
// Outputs: 2024-01-15 10:30:45 [INFO] Application started
std::cerr << timestamp << " [INFO] " << msg << std::endl;
```

### Testing

```flow
test "addition works":
    result is add 2 and 3
    assert result == 5, "2 + 3 should be 5"
```
→
```cpp
void test_addition_works() {
    auto result = add(2, 3);
    if (!(result == 5)) {
        std::cerr << "Assertion failed: 2 + 3 should be 5";
        std::abort();
    }
}
```

### Error Handling

```flow
try:
    result is risky_operation
catch e:
    say "Error occurred"
```
→
```cpp
try {
    auto result = risky_operation();
} catch (const std::exception& e) {
    std::cout << "Error occurred" << std::endl;
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

Flow v1.0 uses natural English:

**Core:**
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
- `ask` = read input
- `{x}` = interpolation

**Collections:**
- `[expr for each x in items]` = list comprehension
- `value | func` = piping
- `items from 1 to 5` = slicing

**I/O:**
- `read/write/append` = file I/O
- `env "VAR"` = environment variable
- `run "cmd"` = shell command

**Functions:**
- `return a and b` = multiple returns
- `x, y is func` = unpacking
- `yield x` = generator
- `@decorator` = function decorator
- `using x is open:` = context manager

**Networking (v1.0):**
- `fetch "url"` = HTTP GET
- `connect "ws://..."` = WebSocket
- `send msg to socket` = WebSocket send

**Regex (v1.0):**
- `match "pattern" in text` = regex test
- `find "pattern" in text` = find all matches
- `replace "pattern" in text with "new"` = replace

**Crypto (v1.0):**
- `hash sha256/md5/sha1 value` = cryptographic hash

**Concurrency (v1.0):**
- `wait expr` = async await
- `do together:` = parallel execution

**Testing (v1.0):**
- `test "name":` = test block
- `assert cond, "msg"` = assertion

**Error Handling (v1.0):**
- `try: ... catch e:` = exception handling
- `throw "error"` = raise exception

**Logging (v1.0):**
- `log info/warn/error "msg"` = structured logging

**Type Conversion:**
- `to_int x` = convert to integer
- `to_float x` = convert to float
- `to_string x` = convert to string
- `length x` = get length/size

**Generates valid C++20. Links with OpenSSL and pthreads.**

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
