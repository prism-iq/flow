// Package parser provides AST construction for the Flow programming language.
// It implements a recursive descent parser for Flow's natural-English syntax.
package parser

import (
	"fmt"
	"strconv"

	"flow/internal/lexer"
)

// ==================== AST TYPES ====================

// Program is the root AST node containing all top-level statements.
type Program struct {
	Statements []Statement
}

// Statement is the interface for all statement nodes.
type Statement interface{ stmt() }

// Expression is the interface for all expression nodes.
type Expression interface{ expr() }

// -------------------- Core Statements --------------------

// Function represents a function definition.
type Function struct {
	Name   string
	Params []string
	Body   []Statement
}

func (Function) stmt() {}

// Return represents a return statement.
type Return struct {
	Value Expression
}

func (Return) stmt() {}

// Assignment represents variable assignment (name is value).
type Assignment struct {
	Name    string
	Value   Expression
	Mutable bool
}

func (Assignment) stmt() {}

// Reassign represents variable reassignment (name becomes value).
type Reassign struct {
	Name  string
	Value Expression
}

func (Reassign) stmt() {}

// ExprStmt wraps an expression as a statement.
type ExprStmt struct {
	Expr Expression
}

func (ExprStmt) stmt() {}

// -------------------- Struct/Method Statements --------------------

// Struct represents a struct definition.
type Struct struct {
	Name   string
	Fields []Field
}

func (Struct) stmt() {}

// Field represents a struct field.
type Field struct {
	Name string
	Type string
}

// Method represents a method definition.
type Method struct {
	Struct string
	Name   string
	Body   []Statement
}

func (Method) stmt() {}

// -------------------- Control Flow Statements --------------------

// If represents an if statement with optional else-if and else.
type If struct {
	Condition Expression
	Then      []Statement
	ElseIfs   []ElseIf
	Else      []Statement
}

func (If) stmt() {}

// ElseIf represents an else-if clause.
type ElseIf struct {
	Condition Expression
	Then      []Statement
}

// ForEach represents a for-each loop.
type ForEach struct {
	Var   string
	Start Expression
	End   Expression // nil for collection iteration
	Body  []Statement
}

func (ForEach) stmt() {}

// Repeat represents a repeat-n-times loop.
type Repeat struct {
	Count int
	Body  []Statement
}

func (Repeat) stmt() {}

// While represents a while loop.
type While struct {
	Condition Expression
	Body      []Statement
}

func (While) stmt() {}

// Skip represents a continue statement.
type Skip struct{}

func (Skip) stmt() {}

// Stop represents a break statement.
type Stop struct{}

func (Stop) stmt() {}

// -------------------- I/O Statements --------------------

// Say represents a print statement with newline.
type Say struct {
	Value Expression
}

func (Say) stmt() {}

// Print represents a print statement without newline.
type Print struct {
	Value Expression
}

func (Print) stmt() {}

// Pause represents a sleep/delay statement.
type Pause struct {
	Milliseconds Expression
}

func (Pause) stmt() {}

// WriteFile represents file writing.
type WriteFile struct {
	Content Expression
	Path    Expression
	Append  bool
}

func (WriteFile) stmt() {}

// -------------------- Advanced Statements --------------------

// UnpackAssign represents unpacking assignment (a, b is func).
type UnpackAssign struct {
	Names   []string
	Value   Expression
	Mutable bool
}

func (UnpackAssign) stmt() {}

// Using represents a context manager block.
type Using struct {
	Name string
	Expr Expression
	Body []Statement
}

func (Using) stmt() {}

// Yield represents a generator yield.
type Yield struct {
	Value Expression
}

func (Yield) stmt() {}

// Decorator represents a decorated function.
type Decorator struct {
	Name     string
	Function Function
}

func (Decorator) stmt() {}

// -------------------- v1.0 Statements --------------------

// DoTogether represents concurrent execution.
type DoTogether struct {
	Body []Statement
}

func (DoTogether) stmt() {}

// WebSocketSend represents sending to a WebSocket.
type WebSocketSend struct {
	Socket  Expression
	Message Expression
}

func (WebSocketSend) stmt() {}

// Log represents structured logging.
type Log struct {
	Level   string // info, warn, error
	Message Expression
}

func (Log) stmt() {}

// Test represents a test block.
type Test struct {
	Name string
	Body []Statement
}

func (Test) stmt() {}

// Assert represents an assertion.
type Assert struct {
	Condition Expression
	Message   Expression
}

func (Assert) stmt() {}

// Try represents error handling.
type Try struct {
	Body    []Statement
	Catch   []Statement
	ErrName string
}

func (Try) stmt() {}

// Throw represents throwing an exception.
type Throw struct {
	Message Expression
}

func (Throw) stmt() {}

// ==================== EXPRESSIONS ====================

// -------------------- Literals --------------------

// IntLit represents an integer literal.
type IntLit struct {
	Value int
}

func (IntLit) expr() {}

// FloatLit represents a float literal.
type FloatLit struct {
	Value float64
}

func (FloatLit) expr() {}

// StringLit represents a string literal.
type StringLit struct {
	Value string
}

func (StringLit) expr() {}

// BoolLit represents a boolean literal.
type BoolLit struct {
	Value bool
}

func (BoolLit) expr() {}

// -------------------- Identifiers and Access --------------------

// Ident represents an identifier.
type Ident struct {
	Name string
}

func (Ident) expr() {}

// MyAccess represents self field access (my field).
type MyAccess struct {
	Field string
}

func (MyAccess) expr() {}

// Access represents field access (obj's field).
type Access struct {
	Object Expression
	Field  string
}

func (Access) expr() {}

// Index represents array indexing.
type Index struct {
	Object Expression
	Index  Expression
}

func (Index) expr() {}

// -------------------- Operations --------------------

// BinaryOp represents a binary operation.
type BinaryOp struct {
	Left  Expression
	Op    string
	Right Expression
}

func (BinaryOp) expr() {}

// UnaryOp represents a unary operation.
type UnaryOp struct {
	Op    string
	Value Expression
}

func (UnaryOp) expr() {}

// -------------------- Collections --------------------

// List represents a list literal.
type List struct {
	Elements []Expression
}

func (List) expr() {}

// ListComprehension represents a list comprehension.
type ListComprehension struct {
	Expr      Expression
	Var       string
	Start     Expression
	End       Expression
	Condition Expression
}

func (ListComprehension) expr() {}

// Slice represents list slicing.
type Slice struct {
	Object Expression
	Start  Expression
	End    Expression
}

func (Slice) expr() {}

// TupleExpr represents a tuple (for multiple returns).
type TupleExpr struct {
	Elements []Expression
}

func (TupleExpr) expr() {}

// -------------------- Function Calls --------------------

// Call represents a function call.
type Call struct {
	Func Expression
	Args []Expression
}

func (Call) expr() {}

// Pipe represents piping (value | func).
type Pipe struct {
	Left  Expression
	Right Expression
}

func (Pipe) expr() {}

// -------------------- I/O Expressions --------------------

// ReadFile represents reading a file.
type ReadFile struct {
	Path Expression
}

func (ReadFile) expr() {}

// Ask represents reading from stdin with optional prompt.
type Ask struct {
	Prompt Expression // optional prompt to display
}

func (Ask) expr() {}

// Now represents current datetime.
type Now struct{}

func (Now) expr() {}

// Today represents current date.
type Today struct{}

func (Today) expr() {}

// Clock represents current time.
type Clock struct{}

func (Clock) expr() {}

// EnvVar represents environment variable access.
type EnvVar struct {
	Name Expression
}

func (EnvVar) expr() {}

// RunCommand represents shell command execution.
type RunCommand struct {
	Command Expression
}

func (RunCommand) expr() {}

// OpenFile represents opening a file.
type OpenFile struct {
	Path Expression
	Mode string
}

func (OpenFile) expr() {}

// -------------------- v1.0 Expressions --------------------

// Fetch represents an HTTP GET request.
type Fetch struct {
	URL     Expression
	Method  string
	Body    Expression
	Headers map[string]string
}

func (Fetch) expr() {}

// ParseJSON represents JSON parsing.
type ParseJSON struct {
	Value Expression
}

func (ParseJSON) expr() {}

// StringifyJSON represents JSON stringification.
type StringifyJSON struct {
	Value Expression
}

func (StringifyJSON) expr() {}

// RegexMatch represents regex matching.
type RegexMatch struct {
	Pattern Expression
	Text    Expression
}

func (RegexMatch) expr() {}

// RegexFindAll represents finding all regex matches.
type RegexFindAll struct {
	Pattern Expression
	Text    Expression
}

func (RegexFindAll) expr() {}

// RegexReplace represents regex replacement.
type RegexReplace struct {
	Pattern     Expression
	Text        Expression
	Replacement Expression
}

func (RegexReplace) expr() {}

// Hash represents cryptographic hashing.
type Hash struct {
	Algorithm string
	Value     Expression
}

func (Hash) expr() {}

// Wait represents async await.
type Wait struct {
	Expr Expression
}

func (Wait) expr() {}

// WebSocketConnect represents WebSocket connection.
type WebSocketConnect struct {
	URL Expression
}

func (WebSocketConnect) expr() {}

// ==================== PARSER ====================

// Parser parses Flow source code into an AST.
type Parser struct {
	tokens []lexer.Token
	pos    int
}

// New creates a new Parser for the given tokens.
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses Flow source code and returns an AST.
func Parse(source string) (*Program, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}
	return New(tokens).Parse()
}

// Parse parses tokens into a Program AST.
func (p *Parser) Parse() (*Program, error) {
	prog := &Program{}

	for !p.atEnd() {
		p.skipNewlines()
		if p.atEnd() {
			break
		}

		stmt, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}

	return prog, nil
}

// -------------------- Top-Level Parsing --------------------

func (p *Parser) parseTopLevel() (Statement, error) {
	// Decorator
	if p.match(lexer.AT_SIGN) {
		return p.parseDecorator()
	}

	// Function
	if p.match(lexer.TO) {
		return p.parseFunction()
	}

	// Struct or Method
	if p.match(lexer.A) {
		return p.parseStructOrMethod()
	}

	// Test block
	if p.match(lexer.TEST) {
		return p.parseTest()
	}

	return nil, p.errorf("expected 'to', 'a', 'test', or '@' at top level")
}

func (p *Parser) parseDecorator() (Statement, error) {
	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected decorator name after '@'")
	}
	p.skipNewlines()

	if !p.match(lexer.TO) {
		return nil, p.errorf("expected 'to' after decorator")
	}

	fn, err := p.parseFunction()
	if err != nil {
		return nil, err
	}

	return Decorator{Name: name, Function: fn.(Function)}, nil
}

func (p *Parser) parseFunction() (Statement, error) {
	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected function name")
	}

	// Parse parameters
	var params []string
	for p.current().Type == lexer.IDENT || p.current().Type == lexer.A {
		params = append(params, p.current().Value)
		p.advance()
		p.match(lexer.AND) // Skip 'and' between parameters
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after function signature")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Function{Name: name, Params: params, Body: body}, nil
}

func (p *Parser) parseStructOrMethod() (Statement, error) {
	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected struct name")
	}

	if p.match(lexer.HAS) {
		return p.parseStructFields(name)
	}
	if p.match(lexer.CAN) {
		return p.parseMethod(name)
	}

	return nil, p.errorf("expected 'has' or 'can' after struct name")
}

func (p *Parser) parseStructFields(name string) (Statement, error) {
	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after 'has'")
	}

	p.skipNewlines()
	var fields []Field

	if p.match(lexer.INDENT) {
		for !p.check(lexer.DEDENT) && !p.atEnd() {
			p.skipNewlines()
			if p.check(lexer.DEDENT) {
				break
			}

			fieldName := p.current().Value
			if !p.match(lexer.IDENT) {
				return nil, p.errorf("expected field name")
			}
			if !p.match(lexer.AS) {
				return nil, p.errorf("expected 'as' after field name")
			}
			fieldType := p.current().Value
			if !p.match(lexer.IDENT) {
				return nil, p.errorf("expected field type")
			}

			fields = append(fields, Field{Name: fieldName, Type: fieldType})
			p.skipNewlines()
		}
		p.match(lexer.DEDENT)
	}

	return Struct{Name: name, Fields: fields}, nil
}

func (p *Parser) parseMethod(structName string) (Statement, error) {
	methodName := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected method name")
	}
	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after method name")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Method{Struct: structName, Name: methodName, Body: body}, nil
}

func (p *Parser) parseTest() (Statement, error) {
	name := ""
	if p.current().Type == lexer.STRING {
		name = p.current().Value
		p.advance()
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after test name")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Test{Name: name, Body: body}, nil
}

// -------------------- Block Parsing --------------------

func (p *Parser) parseBlock() ([]Statement, error) {
	var stmts []Statement

	if !p.match(lexer.INDENT) {
		if p.check(lexer.NEWLINE) || p.atEnd() {
			return stmts, nil
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return []Statement{stmt}, nil
	}

	for !p.check(lexer.DEDENT) && !p.atEnd() {
		p.skipNewlines()
		if p.check(lexer.DEDENT) {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
		p.skipNewlines()
	}

	p.match(lexer.DEDENT)
	return stmts, nil
}

// -------------------- Statement Parsing --------------------

func (p *Parser) parseStatement() (Statement, error) {
	switch p.current().Type {
	// Control flow
	case lexer.IF:
		return p.parseIf()
	case lexer.FOR:
		return p.parseForEach()
	case lexer.REPEAT:
		return p.parseRepeat()
	case lexer.WHILE:
		return p.parseWhile()
	case lexer.RETURN:
		return p.parseReturn()
	case lexer.SKIP:
		p.advance()
		return Skip{}, nil
	case lexer.STOP:
		p.advance()
		return Stop{}, nil

	// I/O
	case lexer.SAY:
		return p.parseSay()
	case lexer.PRINT:
		return p.parsePrint()
	case lexer.PAUSE:
		return p.parsePause()
	case lexer.WRITE:
		return p.parseWriteFile(false)
	case lexer.APPEND:
		return p.parseWriteFile(true)

	// Advanced
	case lexer.USING:
		return p.parseUsing()
	case lexer.YIELD:
		return p.parseYield()

	// v1.0
	case lexer.LOG:
		return p.parseLog()
	case lexer.ASSERT:
		return p.parseAssert()
	case lexer.TRY:
		return p.parseTry()
	case lexer.THROW:
		return p.parseThrow()
	case lexer.DO:
		return p.parseDoTogether()
	case lexer.SEND:
		return p.parseSend()

	// Identifier-based statements (including 'a' which can be a variable)
	case lexer.IDENT, lexer.A:
		return p.parseIdentStatement()

	default:
		return nil, p.errorf("unexpected token %v", p.current())
	}
}

func (p *Parser) parseIf() (Statement, error) {
	p.advance()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after if condition")
	}

	p.skipNewlines()
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	var elseifs []ElseIf
	var elseBody []Statement

	for p.match(lexer.OTHERWISE) {
		if p.match(lexer.IF) {
			elifCond, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if !p.match(lexer.COLON) {
				return nil, p.errorf("expected ':' after elif condition")
			}
			p.skipNewlines()
			elifBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseifs = append(elseifs, ElseIf{Condition: elifCond, Then: elifBody})
		} else {
			if !p.match(lexer.COLON) {
				return nil, p.errorf("expected ':' after otherwise")
			}
			p.skipNewlines()
			elseBody, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
			break
		}
	}

	return If{Condition: cond, Then: then, ElseIfs: elseifs, Else: elseBody}, nil
}

func (p *Parser) parseForEach() (Statement, error) {
	p.advance()

	if !p.match(lexer.EACH) {
		return nil, p.errorf("expected 'each' after 'for'")
	}

	varName := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected variable name")
	}

	if !p.match(lexer.IN) {
		return nil, p.errorf("expected 'in' after variable name")
	}

	start, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var end Expression
	if p.match(lexer.TO) {
		end, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after for each")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return ForEach{Var: varName, Start: start, End: end, Body: body}, nil
}

func (p *Parser) parseRepeat() (Statement, error) {
	p.advance()

	if p.current().Type != lexer.INT {
		return nil, p.errorf("expected integer after 'repeat'")
	}
	count, _ := strconv.Atoi(p.current().Value)
	p.advance()

	if !p.match(lexer.TIMES) {
		return nil, p.errorf("expected 'times' after count")
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after 'times'")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Repeat{Count: count, Body: body}, nil
}

func (p *Parser) parseWhile() (Statement, error) {
	p.advance()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after while condition")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return While{Condition: cond, Body: body}, nil
}

func (p *Parser) parseReturn() (Statement, error) {
	p.advance()

	if p.check(lexer.NEWLINE) || p.check(lexer.DEDENT) || p.atEnd() {
		return Return{}, nil
	}

	val, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	// Multiple returns: return a and b and c
	if p.check(lexer.AND) {
		elements := []Expression{val}
		for p.match(lexer.AND) {
			elem, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
		}
		return Return{Value: TupleExpr{Elements: elements}}, nil
	}

	return Return{Value: val}, nil
}

func (p *Parser) parseSay() (Statement, error) {
	p.advance()
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Say{Value: val}, nil
}

func (p *Parser) parsePrint() (Statement, error) {
	p.advance()
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Print{Value: val}, nil
}

func (p *Parser) parsePause() (Statement, error) {
	p.advance()
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Pause{Milliseconds: val}, nil
}

func (p *Parser) parseWriteFile(appendMode bool) (Statement, error) {
	p.advance()

	// Special handling: if we have "ident to", don't interpret as slice
	var content Expression
	var err error

	if p.current().Type == lexer.IDENT || p.current().Type == lexer.A {
		// Check if next token is TO (meaning simple variable, not slice)
		if p.peek().Type == lexer.TO {
			content = Ident{Name: p.current().Value}
			p.advance()
		} else {
			content, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
	} else {
		content, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	if !p.match(lexer.TO) {
		return nil, p.errorf("expected 'to' after content")
	}

	path, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return WriteFile{Content: content, Path: path, Append: appendMode}, nil
}

func (p *Parser) parseUsing() (Statement, error) {
	p.advance()

	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected variable name after 'using'")
	}

	if !p.match(lexer.IS) {
		return nil, p.errorf("expected 'is' after variable name")
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after using expression")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Using{Name: name, Expr: expr, Body: body}, nil
}

func (p *Parser) parseYield() (Statement, error) {
	p.advance()
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Yield{Value: val}, nil
}

// v1.0 statement parsers

func (p *Parser) parseLog() (Statement, error) {
	p.advance()

	level := "info"
	switch {
	case p.match(lexer.INFO):
		level = "info"
	case p.match(lexer.WARN):
		level = "warn"
	case p.match(lexer.ERROR):
		level = "error"
	}

	msg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return Log{Level: level, Message: msg}, nil
}

func (p *Parser) parseAssert() (Statement, error) {
	p.advance()

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var msg Expression
	if p.match(lexer.COMMA) {
		msg, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	return Assert{Condition: cond, Message: msg}, nil
}

func (p *Parser) parseTry() (Statement, error) {
	p.advance()

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after 'try'")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.CATCH) {
		return nil, p.errorf("expected 'catch' after try block")
	}

	errName := "err"
	if p.current().Type == lexer.IDENT {
		errName = p.current().Value
		p.advance()
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after 'catch'")
	}

	p.skipNewlines()
	catchBody, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Try{Body: body, Catch: catchBody, ErrName: errName}, nil
}

func (p *Parser) parseThrow() (Statement, error) {
	p.advance()
	msg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return Throw{Message: msg}, nil
}

func (p *Parser) parseDoTogether() (Statement, error) {
	p.advance()

	if !p.match(lexer.TOGETHER) {
		return nil, p.errorf("expected 'together' after 'do'")
	}

	if !p.match(lexer.COLON) {
		return nil, p.errorf("expected ':' after 'do together'")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return DoTogether{Body: body}, nil
}

func (p *Parser) parseSend() (Statement, error) {
	p.advance()

	msg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.TO) {
		return nil, p.errorf("expected 'to' after message")
	}

	socket, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return WebSocketSend{Message: msg, Socket: socket}, nil
}

func (p *Parser) parseIdentStatement() (Statement, error) {
	name := p.current().Value
	p.advance()

	// Unpacking: a, b is value
	if p.match(lexer.COMMA) {
		return p.parseUnpacking(name)
	}

	// Assignment: name is value
	if p.match(lexer.IS) {
		return p.parseAssignment(name)
	}

	// Reassignment: name becomes value
	if p.match(lexer.BECOMES) {
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return Reassign{Name: name, Value: val}, nil
	}

	// Expression statement
	p.pos--
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ExprStmt{Expr: expr}, nil
}

func (p *Parser) parseUnpacking(firstName string) (Statement, error) {
	names := []string{firstName}
	for {
		if p.current().Type != lexer.IDENT {
			return nil, p.errorf("expected identifier in unpacking")
		}
		names = append(names, p.current().Value)
		p.advance()
		if !p.match(lexer.COMMA) {
			break
		}
	}

	if !p.match(lexer.IS) {
		return nil, p.errorf("expected 'is' after unpacking names")
	}

	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	mutable := p.parseMutableMarker()
	return UnpackAssign{Names: names, Value: val, Mutable: mutable}, nil
}

func (p *Parser) parseAssignment(name string) (Statement, error) {
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	mutable := p.parseMutableMarker()
	return Assignment{Name: name, Value: val, Mutable: mutable}, nil
}

func (p *Parser) parseMutableMarker() bool {
	if p.match(lexer.COMMA) {
		if p.match(lexer.CAN) && p.match(lexer.CHANGE) {
			return true
		}
	}
	return false
}

// -------------------- Expression Parsing --------------------

func (p *Parser) parseExpression() (Expression, error) {
	return p.parsePipe()
}

func (p *Parser) parsePipe() (Expression, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.PIPE) {
		right, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		left = Pipe{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.OR) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = BinaryOp{Left: left, Op: "||", Right: right}
	}

	return left, nil
}

func (p *Parser) parseAnd() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.AND) {
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = BinaryOp{Left: left, Op: "&&", Right: right}
	}

	return left, nil
}

func (p *Parser) parseComparison() (Expression, error) {
	left, err := p.parseAddition()
	if err != nil {
		return nil, err
	}

	for {
		var op string
		switch p.current().Type {
		case lexer.LT:
			op = "<"
		case lexer.GT:
			op = ">"
		case lexer.LE:
			op = "<="
		case lexer.GE:
			op = ">="
		case lexer.EQ:
			op = "=="
		case lexer.NE:
			op = "!="
		case lexer.IS:
			op = "=="
		default:
			return left, nil
		}
		p.advance()

		right, err := p.parseAddition()
		if err != nil {
			return nil, err
		}
		left = BinaryOp{Left: left, Op: op, Right: right}
	}
}

func (p *Parser) parseAddition() (Expression, error) {
	left, err := p.parseMultiplication()
	if err != nil {
		return nil, err
	}

	for {
		var op string
		switch p.current().Type {
		case lexer.PLUS:
			op = "+"
		case lexer.MINUS:
			op = "-"
		default:
			return left, nil
		}
		p.advance()

		right, err := p.parseMultiplication()
		if err != nil {
			return nil, err
		}
		left = BinaryOp{Left: left, Op: op, Right: right}
	}
}

func (p *Parser) parseMultiplication() (Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		var op string
		switch p.current().Type {
		case lexer.STAR:
			op = "*"
		case lexer.SLASH:
			op = "/"
		case lexer.PERCENT:
			op = "%"
		default:
			return left, nil
		}
		p.advance()

		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = BinaryOp{Left: left, Op: op, Right: right}
	}
}

func (p *Parser) parseUnary() (Expression, error) {
	if p.match(lexer.NOT) {
		val, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return UnaryOp{Op: "!", Value: val}, nil
	}

	if p.match(lexer.MINUS) {
		val, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return UnaryOp{Op: "-", Value: val}, nil
	}

	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(lexer.POSSESSIVE) {
			field := p.current().Value
			if !p.match(lexer.IDENT) {
				return nil, p.errorf("expected field name after 's")
			}
			expr = Access{Object: expr, Field: field}
		} else if p.match(lexer.AT) {
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			expr = Index{Object: expr, Index: idx}
		} else if p.match(lexer.FROM) {
			start, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			var end Expression
			if p.match(lexer.TO) {
				end, err = p.parsePrimary()
				if err != nil {
					return nil, err
				}
			}
			expr = Slice{Object: expr, Start: start, End: end}
		} else if p.match(lexer.TO) {
			if _, isIdent := expr.(Ident); isIdent {
				end, err := p.parsePrimary()
				if err != nil {
					return nil, err
				}
				expr = Slice{Object: expr, Start: nil, End: end}
			} else {
				p.pos--
				break
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) parsePrimary() (Expression, error) {
	switch p.current().Type {
	// Literals
	case lexer.INT:
		val, _ := strconv.Atoi(p.current().Value)
		p.advance()
		return IntLit{Value: val}, nil

	case lexer.FLOAT:
		val, _ := strconv.ParseFloat(p.current().Value, 64)
		p.advance()
		return FloatLit{Value: val}, nil

	case lexer.STRING:
		val := p.current().Value
		p.advance()
		return StringLit{Value: val}, nil

	case lexer.YES:
		p.advance()
		return BoolLit{Value: true}, nil

	case lexer.NO:
		p.advance()
		return BoolLit{Value: false}, nil

	// Self access
	case lexer.MY:
		p.advance()
		field := p.current().Value
		if !p.match(lexer.IDENT) {
			return nil, p.errorf("expected field name after 'my'")
		}
		return MyAccess{Field: field}, nil

	// I/O expressions
	case lexer.READ:
		p.advance()
		path, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return ReadFile{Path: path}, nil

	case lexer.ASK:
		p.advance()
		// Optional prompt
		var prompt Expression
		if p.isExprStart() {
			var err error
			prompt, err = p.parseArgument()
			if err != nil {
				return nil, err
			}
		}
		return Ask{Prompt: prompt}, nil

	case lexer.NOW:
		p.advance()
		return Now{}, nil

	case lexer.TODAY:
		p.advance()
		return Today{}, nil

	case lexer.CLOCK:
		p.advance()
		return Clock{}, nil

	case lexer.ENV:
		p.advance()
		name, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return EnvVar{Name: name}, nil

	case lexer.RUN:
		p.advance()
		cmd, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return RunCommand{Command: cmd}, nil

	case lexer.OPEN:
		p.advance()
		path, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return OpenFile{Path: path, Mode: "read"}, nil

	// v1.0 expressions
	case lexer.FETCH:
		p.advance()
		url, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return Fetch{URL: url, Method: "GET"}, nil

	case lexer.PARSE:
		p.advance()
		val, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return ParseJSON{Value: val}, nil

	case lexer.STRINGIFY:
		p.advance()
		val, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return StringifyJSON{Value: val}, nil

	case lexer.MATCH:
		return p.parseRegexMatch()

	case lexer.FIND:
		return p.parseRegexFindAll()

	case lexer.REPLACE:
		return p.parseRegexReplace()

	case lexer.HASH:
		return p.parseHash()

	case lexer.WAIT:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return Wait{Expr: expr}, nil

	case lexer.CONNECT:
		p.advance()
		url, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return WebSocketConnect{URL: url}, nil

	// Identifiers and function calls
	case lexer.IDENT, lexer.A:
		return p.parseIdentOrCall()

	// Grouping
	case lexer.LPAREN:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RPAREN) {
			return nil, p.errorf("expected ')'")
		}
		return expr, nil

	// List
	case lexer.LBRACKET:
		return p.parseList()

	default:
		return nil, p.errorf("unexpected token: %v", p.current())
	}
}

func (p *Parser) parseRegexMatch() (Expression, error) {
	p.advance()
	pattern, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.IN) {
		return nil, p.errorf("expected 'in' after pattern")
	}
	text, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	return RegexMatch{Pattern: pattern, Text: text}, nil
}

func (p *Parser) parseRegexFindAll() (Expression, error) {
	p.advance()
	pattern, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.IN) {
		return nil, p.errorf("expected 'in' after pattern")
	}
	text, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	return RegexFindAll{Pattern: pattern, Text: text}, nil
}

func (p *Parser) parseRegexReplace() (Expression, error) {
	p.advance()
	pattern, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.IN) {
		return nil, p.errorf("expected 'in' after pattern")
	}
	text, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	if p.current().Type != lexer.IDENT || p.current().Value != "with" {
		return nil, p.errorf("expected 'with' in replace")
	}
	p.advance()
	replacement, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	return RegexReplace{Pattern: pattern, Text: text, Replacement: replacement}, nil
}

func (p *Parser) parseHash() (Expression, error) {
	p.advance()
	algo := "sha256"
	if p.current().Type == lexer.IDENT {
		name := p.current().Value
		if name == "sha256" || name == "md5" || name == "sha1" {
			algo = name
			p.advance()
		}
	}
	val, err := p.parseArgument()
	if err != nil {
		return nil, err
	}
	return Hash{Algorithm: algo, Value: val}, nil
}

func (p *Parser) parseIdentOrCall() (Expression, error) {
	name := p.current().Value
	p.advance()

	var args []Expression
	for p.isArgStart() {
		arg, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		p.match(lexer.AND)
	}

	if len(args) > 0 {
		return Call{Func: Ident{Name: name}, Args: args}, nil
	}

	return Ident{Name: name}, nil
}

func (p *Parser) parseList() (Expression, error) {
	p.advance()

	if p.match(lexer.RBRACKET) {
		return List{Elements: []Expression{}}, nil
	}

	first, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	// List comprehension
	if p.match(lexer.FOR) {
		return p.parseListComprehension(first)
	}

	// Regular list
	elements := []Expression{first}
	for p.match(lexer.COMMA) {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
	}

	if !p.match(lexer.RBRACKET) {
		return nil, p.errorf("expected ']'")
	}

	return List{Elements: elements}, nil
}

func (p *Parser) parseListComprehension(expr Expression) (Expression, error) {
	if !p.match(lexer.EACH) {
		return nil, p.errorf("expected 'each' after 'for'")
	}

	varName := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.errorf("expected variable name")
	}

	if !p.match(lexer.IN) {
		return nil, p.errorf("expected 'in' after variable")
	}

	start, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	var end Expression
	if p.match(lexer.TO) {
		end, err = p.parseOr()
		if err != nil {
			return nil, err
		}
	}

	var condition Expression
	if p.match(lexer.WHERE) {
		condition, err = p.parseOr()
		if err != nil {
			return nil, err
		}
	}

	if !p.match(lexer.RBRACKET) {
		return nil, p.errorf("expected ']'")
	}

	return ListComprehension{
		Expr:      expr,
		Var:       varName,
		Start:     start,
		End:       end,
		Condition: condition,
	}, nil
}

func (p *Parser) parseArgument() (Expression, error) {
	switch p.current().Type {
	case lexer.INT:
		val, _ := strconv.Atoi(p.current().Value)
		p.advance()
		return IntLit{Value: val}, nil
	case lexer.FLOAT:
		val, _ := strconv.ParseFloat(p.current().Value, 64)
		p.advance()
		return FloatLit{Value: val}, nil
	case lexer.STRING:
		val := p.current().Value
		p.advance()
		return StringLit{Value: val}, nil
	case lexer.YES:
		p.advance()
		return BoolLit{Value: true}, nil
	case lexer.NO:
		p.advance()
		return BoolLit{Value: false}, nil
	case lexer.IDENT, lexer.A:
		name := p.current().Value
		p.advance()
		return Ident{Name: name}, nil
	case lexer.LPAREN:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RPAREN) {
			return nil, p.errorf("expected ')'")
		}
		return expr, nil
	case lexer.LBRACKET:
		return p.parseList()
	default:
		return nil, p.errorf("expected argument")
	}
}

func (p *Parser) isArgStart() bool {
	switch p.current().Type {
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.YES, lexer.NO,
		lexer.LPAREN, lexer.LBRACKET, lexer.IDENT, lexer.A:
		return true
	}
	return false
}

// -------------------- Helper Methods --------------------

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peek() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) check(t lexer.TokenType) bool {
	return p.current().Type == t
}

// isExprStart returns true if the current token could start an expression
func (p *Parser) isExprStart() bool {
	switch p.current().Type {
	case lexer.IDENT, lexer.INT, lexer.FLOAT, lexer.STRING,
		lexer.LPAREN, lexer.LBRACKET, lexer.LBRACE,
		lexer.YES, lexer.NO, lexer.NOT, lexer.MINUS, lexer.A:
		return true
	}
	return false
}

func (p *Parser) match(t lexer.TokenType) bool {
	if p.check(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) atEnd() bool {
	return p.current().Type == lexer.EOF
}

func (p *Parser) skipNewlines() {
	for p.match(lexer.NEWLINE) {
	}
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	tok := p.current()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%d:%d: %s (got %v)", tok.Line, tok.Column, msg, tok)
}
