package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	EOF TokenType = iota
	NEWLINE
	INDENT
	DEDENT

	// Literals
	INT
	FLOAT
	STRING

	// Identifiers and keywords
	IDENT

	// Keywords
	TO
	IS
	BECOMES
	CAN
	CHANGE
	IF
	OTHERWISE
	FOR
	EACH
	IN
	REPEAT
	TIMES
	WHILE
	RETURN
	SAY
	AND
	OR
	NOT
	YES
	NO
	SKIP
	STOP
	A
	HAS
	AS
	MY
	AT
	WHERE
	THEN
	READ
	WRITE
	APPEND
	ENV

	// Operators
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	LT
	GT
	LE
	GE
	EQ
	NE
	POSSESSIVE // 's

	// Punctuation
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	LBRACE
	RBRACE
	COLON
	COMMA
	PIPE
)

var keywords = map[string]TokenType{
	"to":        TO,
	"is":        IS,
	"becomes":   BECOMES,
	"can":       CAN,
	"change":    CHANGE,
	"if":        IF,
	"otherwise": OTHERWISE,
	"for":       FOR,
	"each":      EACH,
	"in":        IN,
	"repeat":    REPEAT,
	"times":     TIMES,
	"while":     WHILE,
	"return":    RETURN,
	"say":       SAY,
	"and":       AND,
	"or":        OR,
	"not":       NOT,
	"yes":       YES,
	"no":        NO,
	"skip":      SKIP,
	"stop":      STOP,
	"a":         A,
	"has":       HAS,
	"as":        AS,
	"my":        MY,
	"at":        AT,
	"where":     WHERE,
	"then":      THEN,
	"read":      READ,
	"write":     WRITE,
	"append":    APPEND,
	"env":       ENV,
}

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%v(%q) at %d:%d", t.Type, t.Value, t.Line, t.Column)
}

type Lexer struct {
	input   string
	pos     int
	line    int
	column  int
	tokens  []Token
	indents []int
}

func New(input string) *Lexer {
	return &Lexer{
		input:   input,
		pos:     0,
		line:    1,
		column:  1,
		indents: []int{0},
	}
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		// Handle start of line (indentation)
		if l.column == 1 {
			l.handleIndentation()
		}

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
			if err := l.readString(); err != nil {
				return nil, err
			}

		case unicode.IsDigit(rune(ch)):
			l.readNumber()

		case unicode.IsLetter(rune(ch)) || ch == '_':
			l.readIdent()

		case ch == '+':
			l.emit(PLUS, "+")
			l.advance()
		case ch == '-':
			l.emit(MINUS, "-")
			l.advance()
		case ch == '*':
			l.emit(STAR, "*")
			l.advance()
		case ch == '/':
			l.emit(SLASH, "/")
			l.advance()
		case ch == '%':
			l.emit(PERCENT, "%")
			l.advance()
		case ch == '<':
			if l.peek() == '=' {
				l.emit(LE, "<=")
				l.advance()
				l.advance()
			} else {
				l.emit(LT, "<")
				l.advance()
			}
		case ch == '>':
			if l.peek() == '=' {
				l.emit(GE, ">=")
				l.advance()
				l.advance()
			} else {
				l.emit(GT, ">")
				l.advance()
			}
		case ch == '=' && l.peek() == '=':
			l.emit(EQ, "==")
			l.advance()
			l.advance()
		case ch == '!' && l.peek() == '=':
			l.emit(NE, "!=")
			l.advance()
			l.advance()
		case ch == '\'':
			if l.peek() == 's' {
				l.emit(POSSESSIVE, "'s")
				l.advance()
				l.advance()
			} else {
				return nil, fmt.Errorf("unexpected character '%c' at %d:%d", ch, l.line, l.column)
			}
		case ch == '(':
			l.emit(LPAREN, "(")
			l.advance()
		case ch == ')':
			l.emit(RPAREN, ")")
			l.advance()
		case ch == '[':
			l.emit(LBRACKET, "[")
			l.advance()
		case ch == ']':
			l.emit(RBRACKET, "]")
			l.advance()
		case ch == '{':
			l.emit(LBRACE, "{")
			l.advance()
		case ch == '}':
			l.emit(RBRACE, "}")
			l.advance()
		case ch == ':':
			l.emit(COLON, ":")
			l.advance()
		case ch == ',':
			l.emit(COMMA, ",")
			l.advance()
		case ch == '|':
			l.emit(PIPE, "|")
			l.advance()

		default:
			return nil, fmt.Errorf("unexpected character '%c' at %d:%d", ch, l.line, l.column)
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

func (l *Lexer) current() byte {
	if l.pos >= len(l.input) {
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

func (l *Lexer) advance() {
	l.pos++
	l.column++
}

func (l *Lexer) emit(typ TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:   typ,
		Value:  value,
		Line:   l.line,
		Column: l.column,
	})
}

func (l *Lexer) handleIndentation() {
	indent := 0
	for l.current() == ' ' {
		indent++
		l.advance()
	}
	for l.current() == '\t' {
		indent += 4 // Treat tabs as 4 spaces
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

func (l *Lexer) skipWhitespace() {
	for l.current() == ' ' || l.current() == '\t' {
		l.advance()
	}
}

func (l *Lexer) skipComment() {
	for l.current() != '\n' && l.current() != 0 {
		l.advance()
	}
}

func (l *Lexer) readString() error {
	l.advance() // skip opening quote
	start := l.pos

	for l.current() != '"' && l.current() != 0 {
		if l.current() == '\n' {
			return fmt.Errorf("unterminated string at %d:%d", l.line, l.column)
		}
		l.advance()
	}

	if l.current() == 0 {
		return fmt.Errorf("unterminated string at %d:%d", l.line, l.column)
	}

	value := l.input[start:l.pos]
	l.emit(STRING, value)
	l.advance() // skip closing quote
	return nil
}

func (l *Lexer) readNumber() {
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

func (l *Lexer) readIdent() {
	start := l.pos

	for unicode.IsLetter(rune(l.current())) || unicode.IsDigit(rune(l.current())) || l.current() == '_' {
		l.advance()
	}

	value := l.input[start:l.pos]

	// Check if it's a keyword
	if typ, ok := keywords[strings.ToLower(value)]; ok {
		l.emit(typ, value)
	} else {
		l.emit(IDENT, value)
	}
}
