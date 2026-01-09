// Package lexer provides tokenization for the Flow programming language.
// Flow is a natural-English syntax language that compiles to C++20.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	NEWLINE
	INDENT
	DEDENT

	// Literals
	INT
	FLOAT
	STRING

	// Identifier
	IDENT

	// ==================== KEYWORDS ====================

	// Core keywords
	TO        // function definition
	IS        // assignment
	BECOMES   // reassignment
	RETURN    // return statement
	IF        // conditional
	OTHERWISE // else
	A         // struct definition

	// Struct/Method keywords
	HAS // struct fields
	CAN // method definition
	AS  // type annotation
	MY  // self reference

	// Loop keywords
	FOR    // for loop
	EACH   // for each
	IN     // in collection
	REPEAT // repeat n times
	TIMES  // times keyword
	WHILE  // while loop
	SKIP   // continue
	STOP   // break

	// Boolean keywords
	AND // logical and
	OR  // logical or
	NOT // logical not
	YES // true
	NO  // false

	// I/O keywords
	SAY    // print with newline
	PRINT  // print without newline
	ASK    // read from stdin
	NOW    // current datetime
	TODAY  // current date
	CLOCK  // current time
	PAUSE  // sleep milliseconds
	READ   // read file
	WRITE  // write file
	APPEND // append file
	ENV    // environment variable
	RUN    // shell command
	OPEN   // open file

	// Advanced keywords
	WHERE  // filter condition
	THEN   // then keyword
	USING  // context manager
	FROM   // slice start
	YIELD  // generator yield
	CHANGE // mutable marker

	// v1.0 Networking
	FETCH   // HTTP GET
	CONNECT // WebSocket connect
	SEND    // WebSocket send

	// v1.0 Data Processing
	PARSE     // JSON parse
	STRINGIFY // JSON stringify
	MATCH     // regex match
	FIND      // regex find all
	REPLACE   // regex replace
	HASH      // cryptographic hash

	// v1.0 Concurrency
	WAIT     // async await
	DO       // do together
	TOGETHER // together keyword

	// v1.0 Logging
	LOG   // log statement
	INFO  // log level
	WARN  // log level
	ERROR // log level

	// v1.0 Testing
	TEST   // test block
	ASSERT // assertion

	// v1.0 Error Handling
	TRY   // try block
	CATCH // catch block
	THROW // throw exception

	// ==================== OPERATORS ====================

	// Arithmetic
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %

	// Comparison
	LT // <
	GT // >
	LE // <=
	GE // >=
	EQ // ==
	NE // !=

	// Special operators
	POSSESSIVE // 's
	AT         // at (index)
	AT_SIGN    // @ (decorator)
	PIPE       // | (piping)

	// ==================== PUNCTUATION ====================

	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	LBRACE   // {
	RBRACE   // }
	COLON    // :
	COMMA    // ,
)

// keywords maps Flow keywords to their token types.
var keywords = map[string]TokenType{
	// Core
	"to": TO, "is": IS, "becomes": BECOMES, "return": RETURN,
	"if": IF, "otherwise": OTHERWISE, "a": A,
	// Struct/Method
	"has": HAS, "can": CAN, "as": AS, "my": MY,
	// Loops
	"for": FOR, "each": EACH, "in": IN, "repeat": REPEAT,
	"times": TIMES, "while": WHILE, "skip": SKIP, "stop": STOP,
	// Boolean
	"and": AND, "or": OR, "not": NOT, "yes": YES, "no": NO,
	// I/O
	"say": SAY, "print": PRINT, "ask": ASK, "pause": PAUSE,
	"now": NOW, "today": TODAY, "clock": CLOCK,
	"read": READ, "write": WRITE, "append": APPEND,
	"env": ENV, "run": RUN, "open": OPEN,
	// Advanced
	"where": WHERE, "then": THEN, "using": USING, "from": FROM,
	"yield": YIELD, "change": CHANGE,
	// v1.0 Networking
	"fetch": FETCH, "connect": CONNECT, "send": SEND,
	// v1.0 Data Processing
	"parse": PARSE, "stringify": STRINGIFY, "match": MATCH,
	"find": FIND, "replace": REPLACE, "hash": HASH,
	// v1.0 Concurrency
	"wait": WAIT, "do": DO, "together": TOGETHER,
	// v1.0 Logging
	"log": LOG, "info": INFO, "warn": WARN, "error": ERROR,
	// v1.0 Testing
	"test": TEST, "assert": ASSERT,
	// v1.0 Error Handling
	"try": TRY, "catch": CATCH, "throw": THROW,
}

// Token represents a lexical token with its type, value, and position.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%v(%q)@%d:%d", t.Type, t.Value, t.Line, t.Column)
}

// Lexer tokenizes Flow source code.
type Lexer struct {
	input   string
	pos     int
	line    int
	column  int
	tokens  []Token
	indents []int
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		input:   input,
		line:    1,
		column:  1,
		indents: []int{0},
	}
}

// Tokenize processes the input and returns all tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for !l.atEnd() {
		if l.column == 1 {
			l.handleIndentation()
		}
		if err := l.scanToken(); err != nil {
			return nil, err
		}
	}

	// Emit remaining DEDENTs
	for len(l.indents) > 1 {
		l.indents = l.indents[:len(l.indents)-1]
		l.emit(DEDENT, "")
	}
	l.emit(EOF, "")

	return l.tokens, nil
}

// scanToken scans and emits a single token.
func (l *Lexer) scanToken() error {
	ch := l.current()

	switch {
	case ch == '\n':
		l.emit(NEWLINE, "\n")
		l.advance()
		l.line++
		l.column = 1

	case ch == ' ' || ch == '\t':
		l.skipWhitespace()

	case ch == '/' && l.peek() == '/':
		l.skipComment()

	case ch == '"':
		return l.scanString()

	case unicode.IsDigit(rune(ch)):
		l.scanNumber()

	case unicode.IsLetter(rune(ch)) || ch == '_':
		l.scanIdentifier()

	default:
		return l.scanOperator()
	}

	return nil
}

// scanOperator scans operators and punctuation.
func (l *Lexer) scanOperator() error {
	ch := l.current()

	switch ch {
	case '+':
		l.emitAdvance(PLUS, "+")
	case '-':
		l.emitAdvance(MINUS, "-")
	case '*':
		l.emitAdvance(STAR, "*")
	case '/':
		l.emitAdvance(SLASH, "/")
	case '%':
		l.emitAdvance(PERCENT, "%")
	case '<':
		l.scanCompare(LT, "<", LE, "<=")
	case '>':
		l.scanCompare(GT, ">", GE, ">=")
	case '=':
		if l.peek() == '=' {
			l.emit(EQ, "==")
			l.advance()
			l.advance()
		} else {
			return l.unexpectedChar()
		}
	case '!':
		if l.peek() == '=' {
			l.emit(NE, "!=")
			l.advance()
			l.advance()
		} else {
			return l.unexpectedChar()
		}
	case '\'':
		if l.peek() == 's' {
			l.emit(POSSESSIVE, "'s")
			l.advance()
			l.advance()
		} else {
			return l.unexpectedChar()
		}
	case '(':
		l.emitAdvance(LPAREN, "(")
	case ')':
		l.emitAdvance(RPAREN, ")")
	case '[':
		l.emitAdvance(LBRACKET, "[")
	case ']':
		l.emitAdvance(RBRACKET, "]")
	case '{':
		l.emitAdvance(LBRACE, "{")
	case '}':
		l.emitAdvance(RBRACE, "}")
	case ':':
		l.emitAdvance(COLON, ":")
	case ',':
		l.emitAdvance(COMMA, ",")
	case '|':
		l.emitAdvance(PIPE, "|")
	case '@':
		l.emitAdvance(AT_SIGN, "@")
	default:
		return l.unexpectedChar()
	}

	return nil
}

// scanCompare handles two-character comparison operators.
func (l *Lexer) scanCompare(singleType TokenType, singleVal string, doubleType TokenType, doubleVal string) {
	if l.peek() == '=' {
		l.emit(doubleType, doubleVal)
		l.advance()
		l.advance()
	} else {
		l.emitAdvance(singleType, singleVal)
	}
}

// scanString scans a string literal with escape sequence support.
func (l *Lexer) scanString() error {
	l.advance() // skip opening quote
	var result strings.Builder

	for l.current() != '"' && !l.atEnd() {
		if l.current() == '\n' {
			return fmt.Errorf("unterminated string at %d:%d", l.line, l.column)
		}
		if l.current() == '\\' {
			l.advance() // skip backslash
			switch l.current() {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case '0':
				result.WriteByte(0)
			default:
				// Unknown escape, keep as-is
				result.WriteByte('\\')
				result.WriteByte(l.current())
			}
		} else {
			result.WriteByte(l.current())
		}
		l.advance()
	}

	if l.atEnd() {
		return fmt.Errorf("unterminated string at %d:%d", l.line, l.column)
	}

	l.emit(STRING, result.String())
	l.advance() // skip closing quote
	return nil
}

// scanNumber scans an integer or float literal.
func (l *Lexer) scanNumber() {
	start := l.pos
	isFloat := false

	for unicode.IsDigit(rune(l.current())) {
		l.advance()
	}

	if l.current() == '.' && unicode.IsDigit(rune(l.peek())) {
		isFloat = true
		l.advance() // skip dot
		for unicode.IsDigit(rune(l.current())) {
			l.advance()
		}
	}

	value := l.input[start:l.pos]
	if isFloat {
		l.emit(FLOAT, value)
	} else {
		l.emit(INT, value)
	}
}

// scanIdentifier scans an identifier or keyword.
func (l *Lexer) scanIdentifier() {
	start := l.pos

	for unicode.IsLetter(rune(l.current())) || unicode.IsDigit(rune(l.current())) || l.current() == '_' {
		l.advance()
	}

	value := l.input[start:l.pos]
	if typ, ok := keywords[strings.ToLower(value)]; ok {
		l.emit(typ, value)
	} else {
		l.emit(IDENT, value)
	}
}

// handleIndentation processes indentation at the start of a line.
func (l *Lexer) handleIndentation() {
	indent := 0
	for l.current() == ' ' {
		indent++
		l.advance()
	}
	for l.current() == '\t' {
		indent += 4
		l.advance()
	}

	// Skip empty lines and comment-only lines
	if l.current() == '\n' || (l.current() == '/' && l.peek() == '/') {
		return
	}

	currentIndent := l.indents[len(l.indents)-1]

	if indent > currentIndent {
		l.indents = append(l.indents, indent)
		l.emit(INDENT, "")
	} else if indent < currentIndent {
		for len(l.indents) > 1 && l.indents[len(l.indents)-1] > indent {
			l.indents = l.indents[:len(l.indents)-1]
			l.emit(DEDENT, "")
		}
	}
}

// Helper methods

func (l *Lexer) current() byte {
	if l.atEnd() {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peek() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) advance() {
	l.pos++
	l.column++
}

func (l *Lexer) emit(typ TokenType, value string) {
	l.tokens = append(l.tokens, Token{Type: typ, Value: value, Line: l.line, Column: l.column})
}

func (l *Lexer) emitAdvance(typ TokenType, value string) {
	l.emit(typ, value)
	l.advance()
}

func (l *Lexer) skipWhitespace() {
	for l.current() == ' ' || l.current() == '\t' {
		l.advance()
	}
}

func (l *Lexer) skipComment() {
	for l.current() != '\n' && !l.atEnd() {
		l.advance()
	}
}

func (l *Lexer) unexpectedChar() error {
	return fmt.Errorf("unexpected character '%c' at %d:%d", l.current(), l.line, l.column)
}
