# Flow Language v0.3 - Human-First Syntax

Flow reads like English, compiles to C++.

---

## Axioms

1. **Explicit > Implicit** - No hidden behavior
2. **Errors are values** - No exceptions, must handle
3. **Null doesn't exist** - Use `maybe`, compiler enforces
4. **Immutable by default** - Say `can change` for mutable
5. **No hidden allocations** - Stack default, `on heap` explicit
6. **One way to do things** - No overloading
7. **Composition > Inheritance** - No classes, embed don't extend
8. **Zero cost abstractions** - Generics compile to concrete types
9. **Fail at compile time** - Strong typing, no implicit conversions
10. **Readable > Clever** - Code reads like prose

---

## Variables

```flow
// Immutable (default)
name is "Flow"
age is 25
numbers are [1, 2, 3]
pi is 3.14159

// Mutable
count is 0, can change
count becomes 5

// Type inference, or explicit
score is 100              // inferred as number
name is "Bob" as text     // explicit
```

---

## Functions

```flow
// Simple function
to greet someone:
    say "Hello, {someone}!"

// With return
to add a and b:
    return a + b

// Multiple parameters
to calculate tax on amount at rate:
    return amount * rate

// Entry point
to start:
    greet "World"
    result is add 2 and 3
    say result
```

---

## Conditions

```flow
if age >= 18:
    say "Adult"
otherwise:
    say "Minor"

// Chained
if score > 90:
    say "Excellent"
otherwise if score > 70:
    say "Good"
otherwise:
    say "Keep trying"

// Boolean logic (natural words)
if logged in and is admin:
    say "Full access"

if not verified or expired:
    say "Access denied"
```

---

## Loops

```flow
// For each
for each item in list:
    say item

for each i in 1 to 10:
    say i

// Repeat
repeat 5 times:
    say "hello"

// While
while count < 10:
    count becomes count + 1

// Control
for each n in numbers:
    if n is 0:
        skip            // continue
    if n > 100:
        stop            // break
    say n
```

---

## Structs (Things)

```flow
// Define a thing
a Person has:
    name as text
    age as number

// Create
bob is a Person with name "Bob", age 30

// Access
say bob's name
say bob's age

// Update (if mutable)
bob is a Person with name "Bob", age 30, can change
bob's age becomes 31
```

---

## Methods

```flow
a Person has:
    name as text
    age as number

a Person can introduce:
    say "I'm {my name}, {my age} years old"

a Person can have birthday:
    my age becomes my age + 1

// Usage
alice is a Person with name "Alice", age 25
alice introduce
alice have birthday
```

---

## Error Handling

```flow
// Functions that can fail return result
to open file at path:
    // ... returns result or error

// Handle errors naturally
result is try open file at "data.txt"

if result failed:
    say "Could not open: {result's error}"
    return

// Use the value
file is result's value
process file

// Or ignore error explicitly
_ is try delete file at "temp.txt"
```

---

## Optionals (Maybe)

```flow
// Maybe there's a value, maybe not
user is maybe find user by id 42

if user exists:
    say user's name
otherwise:
    say "Not found"

// With default
name is user's name or "Anonymous"

// Chain maybes
city is maybe user's address's city
```

---

## Collections

```flow
// Lists
numbers are [1, 2, 3, 4, 5]
names are ["Alice", "Bob", "Carol"]

// Access
first is numbers at 0
last is numbers at -1

// Add/remove (mutable)
items are [], can change
add "apple" to items
remove "apple" from items

// Maps
scores are {
    "Alice": 95,
    "Bob": 87
}

alice score is scores at "Alice"
```

---

## Async

```flow
// Wait for async operation
data is wait fetch "https://api.com/data"

// Do multiple things together
do together:
    users is fetch "/users"
    posts is fetch "/posts"
    comments is fetch "/comments"
then:
    combine users, posts, comments
```

---

## Composition

```flow
// Embed, don't inherit
a Worker has:
    person as Person     // embedded
    job as text
    salary as number

// Worker gets Person's fields
w is a Worker with person (name "Bob", age 30), job "Engineer", salary 75000
say w's person's name    // "Bob"

// Interfaces (contracts)
a Printable can:
    print

a Person can print:
    say "{my name}, {my age}"

a Worker can print:
    say "{my person's name}: {my job}"

// Use interface
to display thing as Printable:
    thing print
```

---

## Memory

```flow
// Stack by default (fast, automatic cleanup)
point is a Point with x 10, y 20

// Heap when needed (explicit)
buffer is a Buffer with size 1024, on heap

// Ownership is clear
data is load file                    // owns data
process data                         // borrows data
save data to "output.txt"           // borrows data
// data freed here automatically
```

---

## Modules

```flow
// file: math.flow
to add a and b:
    return a + b

to multiply a and b:
    return a * b

// file: main.flow
use math

to start:
    result is math add 2 and 3
    say result
```

---

## Complete Example

```flow
// A simple task manager

a Task has:
    title as text
    done as yes/no

a Task can toggle:
    my done becomes not my done

a Task can display:
    status is "x" if my done otherwise " "
    say "[{status}] {my title}"

to start:
    tasks are [], can change

    add (a Task with title "Learn Flow", done no) to tasks
    add (a Task with title "Build app", done no) to tasks
    add (a Task with title "Deploy", done no) to tasks

    // Complete first task
    tasks at 0 toggle

    say "My Tasks:"
    for each task in tasks:
        task display
```

Output:
```
My Tasks:
[x] Learn Flow
[ ] Build app
[ ] Deploy
```

---

## Vocabulary Reference

| Flow | Meaning | C++ |
|------|---------|-----|
| `is` | assignment | `=` |
| `becomes` | reassignment | `=` |
| `can change` | mutable | `mut` |
| `to ... :` | function | `fn` |
| `return` | return value | `return` |
| `a X has:` | struct | `struct` |
| `a X can Y:` | method | member fn |
| `if/otherwise` | conditional | `if/else` |
| `for each` | iteration | `for` |
| `repeat N times` | counted loop | `for` |
| `while` | while loop | `while` |
| `skip` | continue | `continue` |
| `stop` | break | `break` |
| `and/or/not` | logic | `&&/\|\|/!` |
| `yes/no` | booleans | `true/false` |
| `try` | fallible call | result type |
| `maybe` | optional | `optional` |
| `wait` | async | `co_await` |
| `do together` | concurrent | async spawn |
| `say` | print | `cout` |
| `my` | self reference | `this->` |
| `'s` | member access | `.` |
| `at` | index access | `[]` |
| `on heap` | heap alloc | `new` |
