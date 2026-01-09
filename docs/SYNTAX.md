# Flow Language Syntax

**Version:** 0.2.0
**Philosophy:** English words, not computer jargon.

---

## Core Principle

```
Code should read like you're explaining it to a human.
Power of C++. Clarity of Python. Words that make sense.
```

---

## Vocabulary

| Flow | Traditional | Meaning |
|------|-------------|---------|
| `do` | fn, func, def | "do this action" |
| `thing` | struct, class | "a thing with properties" |
| `kind` | enum | "a kind of X" |
| `when` | match, switch | "when X, do Y" |
| `need` | import, use | "I need this" |
| `give` | return | "give back result" |
| `start` | spawn, async | "start this task" |
| `wait` | await | "wait for it" |
| `pipe` | channel | "data flows through" |
| `later` | defer | "do this later" |
| `try`/`fail` | Result/Error | "try, might fail" |
| `good` | Ok, Some | "good result" |
| `nothing` | None, nil, null | "nothing here" |
| `can` | interface, trait | "can do X" |
| `each` | for | "for each item" |
| `vary` | mut, var | "this can vary" |
| `me` | self, this | "myself" |
| `say` | print | "say this" |
| `yes`/`no` | true/false | natural booleans |
| `and`/`or`/`not` | &&/\|\|/! | natural logic |

---

## Basics

### Hello World

```flow
do main:
    say "Hello, World!"
```

### Variables

```flow
// Immutable by default (just assign)
name = "Flow"
age = 25
pi = 3.14159

// Mutable (can vary)
vary count = 0
count = count + 1

// Type annotation (optional, usually inferred)
score: int = 100
```

### Comments

```flow
// Single line

/*
   Multi-line
   comment
*/
```

---

## Functions

### Basic

```flow
do greet(name):
    say "Hello, " + name

do add(a, b):
    give a + b
```

### One-liners

```flow
do double(x) = x * 2
do square(x) = x * x
do adult?(age) = age >= 18
```

### With Types

```flow
do divide(a: int, b: int) -> try int:
    if b == 0:
        give fail("division by zero")
    give good(a / b)
```

### Default Arguments

```flow
do greet(name, greeting = "Hello"):
    say "#{greeting}, #{name}!"

greet("World")           // Hello, World!
greet("World", "Hi")     // Hi, World!
```

### Multiple Returns

```flow
do minmax(items):
    give (items.min(), items.max())

(lo, hi) = minmax([3, 1, 4, 1, 5])
```

---

## Types

### Primitives

| Flow | Description |
|------|-------------|
| `int` | Integer |
| `float` | Floating point |
| `str` | String |
| `bool` | yes/no |
| `byte` | Single byte |

### Collections

```flow
// List
items = [1, 2, 3, 4, 5]
items: [int] = []

// Map
scores = {"alice": 100, "bob": 85}
scores: {str: int} = {}

// Set
uniques = {1, 2, 3}
```

### Nothing and Maybe

```flow
// Nothing = absence of value
result = nothing

// Maybe = might have value
do find(items, target) -> maybe int:
    each i, item in items:
        if item == target:
            give some(i)
    give nothing
```

---

## Things (Structs)

### Definition

```flow
thing Point:
    x: int
    y: int

thing Person:
    name: str
    age: int
    email: maybe str    // optional field
```

### Short Form

```flow
thing Point(x: int, y: int)
thing Color(r: byte, g: byte, b: byte)
```

### Methods

```flow
thing Circle:
    x: int
    y: int
    radius: float

    do area(me) = 3.14159 * me.radius * me.radius

    do contains?(me, px, py):
        dx = px - me.x
        dy = py - me.y
        give (dx*dx + dy*dy) <= me.radius * me.radius

    do grow(me, factor):
        me.radius = me.radius * factor
```

### Creation

```flow
p = Point(10, 20)
c = Circle(0, 0, 5.0)

// Named arguments
person = Person(
    name: "Alice",
    age: 30,
    email: some("alice@example.com")
)
```

---

## Kinds (Enums)

### Simple

```flow
kind Color:
    Red
    Green
    Blue

kind Direction:
    Up
    Down
    Left
    Right
```

### With Data

```flow
kind Status:
    Pending
    Running(progress: int)
    Done(result: str)
    Failed(error: str)

kind Shape:
    Circle(radius: float)
    Rectangle(width: float, height: float)
    Triangle(a: float, b: float, c: float)
```

### Usage

```flow
status = Status.Running(45)

when status:
    Pending => say "Waiting..."
    Running(p) => say "Progress: #{p}%"
    Done(r) => say "Result: #{r}"
    Failed(e) => say "Error: #{e}"
```

---

## Control Flow

### If / Else

```flow
if age >= 18:
    say "adult"
else:
    say "minor"

// Chained
if score >= 90:
    say "A"
else if score >= 80:
    say "B"
else if score >= 70:
    say "C"
else:
    say "F"

// As expression
status = if online: "here" else: "away"
```

### When (Pattern Matching)

```flow
when value:
    0 => say "zero"
    1 => say "one"
    2..10 => say "small"
    _ => say "big"

// With destructuring
when point:
    Point(0, 0) => say "origin"
    Point(x, 0) => say "on x-axis at #{x}"
    Point(0, y) => say "on y-axis at #{y}"
    Point(x, y) => say "at #{x}, #{y}"

// With guards
when age:
    n if n < 0 => say "invalid"
    n if n < 18 => say "minor"
    n if n < 65 => say "adult"
    _ => say "senior"
```

### Pattern Matching in Functions

```flow
do factorial(0) = 1
do factorial(n) = n * factorial(n - 1)

do fib(0) = 0
do fib(1) = 1
do fib(n) = fib(n-1) + fib(n-2)

do len([]) = 0
do len([_, ...rest]) = 1 + len(rest)
```

---

## Loops

### Each

```flow
each item in items:
    say item

each i in 1..10:
    say i

each i in 1..=10:    // inclusive
    say i

// With index
each i, item in items:
    say "#{i}: #{item}"

// With step
each i in 0..100 by 10:
    say i
```

### While

```flow
vary x = 10
while x > 0:
    say x
    x = x - 1
say "Boom!"
```

### Loop Control

```flow
each i in 1..100:
    if i == 5:
        skip        // continue
    if i == 10:
        stop        // break
    say i
```

### Comprehensions

```flow
// List
squares = [x*x each x in 1..10]
evens = [x each x in items if x % 2 == 0]

// Map
scores = {name: score each (name, score) in results}

// With transform
names = [person.name.upper() each person in people if person.active]
```

---

## Pipes

```flow
// Chain operations left to right
result = data
    |> parse
    |> validate
    |> transform
    |> save

// vs nested calls
result = save(transform(validate(parse(data))))

// With lambdas
result = numbers
    |> filter(x => x > 0)
    |> map(x => x * 2)
    |> sum()

// UFCS - these are equivalent
items.filter(x => x > 0).map(double).sum()
sum(map(filter(items, x => x > 0), double))
```

---

## Error Handling

### Try and Fail

```flow
do read_file(path) -> try str:
    if not exists(path):
        give fail("file not found: #{path}")
    give good(read_contents(path))

do parse_int(s) -> try int:
    // ... parsing logic
    if invalid:
        give fail("not a number")
    give good(number)
```

### Using Results

```flow
when read_file("data.txt"):
    good(content) => process(content)
    fail(error) => say "Error: #{error}"

// Or with ? propagation
do load_config() -> try Config:
    text = read_file("config.json")?    // propagates fail
    data = parse_json(text)?
    give good(Config.from(data))

// Default values
port = parse_int(env("PORT")) or 8080
name = user.nickname or user.name or "Anonymous"
```

### Check (Guard Clauses)

```flow
do process(data):
    check data.valid else give fail("invalid data")
    check data.size < MAX else give fail("too large")
    check user.authorized else give fail("not allowed")

    // continue with valid data...
```

---

## Concurrency

### Start (Spawn)

```flow
// Start a task in background
start expensive_calculation()

// Start and get handle
task = start fetch(url)
result = wait task
```

### Wait

```flow
// Wait for single task
data = wait fetch(url)

// Wait for all
results = wait all(
    fetch(url1),
    fetch(url2),
    fetch(url3)
)

// Wait for first
first = wait any(
    fetch(url1),
    fetch(url2),
    timeout(5s)
)
```

### Pipes (Channels)

```flow
// Create pipe
p = pipe<int>()

// Send
p <- 42

// Receive
value = <-p

// In practice
do producer(p):
    each i in 1..100:
        p <- i
    close(p)

do consumer(p):
    each value in p:
        say "Got: #{value}"

do main:
    p = pipe<int>()
    start producer(p)
    start consumer(p)
    wait all_done()
```

### Parallel Operations

```flow
// Parallel map
results = items |> parallel(process)

// Parallel each
parallel each url in urls:
    fetch_and_save(url)
```

---

## Later (Defer)

```flow
do process_file(path):
    f = open(path)
    later close(f)      // always runs at end

    // work with f...
    // close(f) called automatically

do transaction():
    begin()
    later if fail: rollback()
    later if good: commit()

    // operations...
```

---

## Interfaces (Can)

```flow
// Define capability
can Show:
    do show(me) -> str

can Compare:
    do compare(me, other) -> int

can Iterate<T>:
    do next(me) -> maybe T

// Implement by having the methods
thing Point:
    x: int
    y: int

    do show(me) = "(#{me.x}, #{me.y})"

// Point automatically "can Show"

// Use in function
do print_all(items: [can Show]):
    each item in items:
        say item.show()
```

---

## Modules

### Importing

```flow
need json                    // standard library
need http                    // standard library
need myproject/utils         // local module
need github.com/user/lib     // external

// Selective
need json: {parse, stringify}
need http: {get, post}

// Aliased
need very_long_name as short
```

### Exporting

```flow
// In utils.flow

// Public (exported)
pub do helper():
    // ...

// Private (not exported)
do internal():
    // ...

pub thing Config:
    // ...
```

---

## Memory & Performance

### Stack vs Heap

```flow
// Stack by default (fast)
point = Point(1, 2)

// Heap when needed (box)
big_data = box [1, 2, 3, ... millions]

// Reference
do modify(ref point):
    point.x = 10
```

### Inline Hints

```flow
@inline
do tiny_function(x) = x + 1

@noinline
do big_function():
    // ...
```

### Comptime

```flow
// Computed at compile time
@comptime VERSION = "1.0.0"
@comptime MAX_SIZE = 1024 * 1024

@comptime do factorial(n):
    if n <= 1: give 1
    give n * factorial(n - 1)

// Used at compile time
LOOKUP_TABLE = @comptime [factorial(i) each i in 0..20]
```

---

## Complete Example

```flow
need http
need json

thing Todo:
    id: int
    title: str
    done: bool

    do toggle(me):
        me.done = not me.done

    do show(me):
        mark = if me.done: "✓" else: "○"
        give "#{mark} #{me.title}"

do fetch_todos(url) -> try [Todo]:
    response = wait http.get(url)?
    data = json.parse(response.body)?
    give good(data.as([Todo]))

do main:
    say "Fetching todos..."

    when fetch_todos("https://api.example.com/todos"):
        good(todos) =>
            say "Found #{todos.len()} items:\n"
            each todo in todos:
                say todo.show()

            // Count done
            done = todos |> filter(t => t.done) |> len()
            say "\n#{done}/#{todos.len()} completed"

        fail(e) =>
            say "Failed: #{e}"
```

---

## Quick Reference

```
╔══════════════════════════════════════════════════════════════╗
║  FLOW CHEAT SHEET                                            ║
╠══════════════════════════════════════════════════════════════╣
║  do name(args):          define function                     ║
║  do name(x) = expr       one-liner function                  ║
║  thing Name: ...         define struct                       ║
║  kind Name: ...          define enum                         ║
║  when x: ...             pattern match                       ║
║  each x in items:        loop                                ║
║  if/else                 condition                           ║
║  give value              return                              ║
║  need module             import                              ║
║  start task()            spawn async                         ║
║  wait task               await result                        ║
║  pipe<T>()               create channel                      ║
║  later action()          defer to end                        ║
║  try/good/fail           error handling                      ║
║  x?                      propagate error                     ║
║  x or default            default if nothing/fail             ║
║  can Name: ...           interface                           ║
║  x |> f |> g             pipe chain                          ║
║  [x each x in l]         list comprehension                  ║
║  yes/no                  booleans                            ║
║  and/or/not              logic operators                     ║
║  vary x = val            mutable variable                    ║
║  say "text"              print                               ║
║  check cond else ...     guard clause                        ║
║  me.field                self reference                      ║
║  nothing/some(x)         optional values                     ║
╚══════════════════════════════════════════════════════════════╝
```

---

## Design Principles

1. **Words over symbols** - `and` not `&&`, `or` not `||`
2. **Obvious names** - `thing` not `struct`, `when` not `match`
3. **Inference by default** - types optional unless ambiguous
4. **Immutable by default** - explicit `vary` for mutation
5. **Expression everything** - if/when/blocks return values
6. **Errors are values** - no exceptions, explicit handling
7. **Concurrency is easy** - `start`/`wait`/`pipe`
8. **Zero cost abstractions** - compiles to efficient C++
