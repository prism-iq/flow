package parser

import (
	"fmt"
	"strconv"

	"flow/internal/lexer"
)

// AST Types

type Program struct {
	Statements []Statement
}

type Statement interface {
	stmt()
}

type Function struct {
	Name   string
	Params []string
	Body   []Statement
}

func (Function) stmt() {}

type Struct struct {
	Name   string
	Fields []Field
}

func (Struct) stmt() {}

type Field struct {
	Name string
	Type string
}

type Method struct {
	Struct string
	Name   string
	Body   []Statement
}

func (Method) stmt() {}

type If struct {
	Condition Expression
	Then      []Statement
	ElseIfs   []ElseIf
	Else      []Statement
}

func (If) stmt() {}

type ElseIf struct {
	Condition Expression
	Then      []Statement
}

type ForEach struct {
	Var   string
	Start Expression
	End   Expression // nil for collection iteration
	Body  []Statement
}

func (ForEach) stmt() {}

type Repeat struct {
	Count int
	Body  []Statement
}

func (Repeat) stmt() {}

type While struct {
	Condition Expression
	Body      []Statement
}

func (While) stmt() {}

type Return struct {
	Value Expression
}

func (Return) stmt() {}

type Say struct {
	Value Expression
}

func (Say) stmt() {}

type Assignment struct {
	Name    string
	Value   Expression
	Mutable bool
}

func (Assignment) stmt() {}

type Reassign struct {
	Name  string
	Value Expression
}

func (Reassign) stmt() {}

type Skip struct{}

func (Skip) stmt() {}

type Stop struct{}

func (Stop) stmt() {}

type ExprStmt struct {
	Expr Expression
}

func (ExprStmt) stmt() {}

// Expressions

type Expression interface {
	expr()
}

type BinaryOp struct {
	Left  Expression
	Op    string
	Right Expression
}

func (BinaryOp) expr() {}

type UnaryOp struct {
	Op    string
	Value Expression
}

func (UnaryOp) expr() {}

type IntLit struct {
	Value int
}

func (IntLit) expr() {}

type FloatLit struct {
	Value float64
}

func (FloatLit) expr() {}

type StringLit struct {
	Value string
}

func (StringLit) expr() {}

type BoolLit struct {
	Value bool
}

func (BoolLit) expr() {}

type Ident struct {
	Name string
}

func (Ident) expr() {}

type MyAccess struct {
	Field string
}

func (MyAccess) expr() {}

type Access struct {
	Object Expression
	Field  string
}

func (Access) expr() {}

type Index struct {
	Object Expression
	Index  Expression
}

func (Index) expr() {}

type Call struct {
	Func Expression
	Args []Expression
}

func (Call) expr() {}

type List struct {
	Elements []Expression
}

func (List) expr() {}

type ListComprehension struct {
	Expr      Expression // The expression to evaluate for each item
	Var       string     // Loop variable name
	Start     Expression // Collection or range start
	End       Expression // Range end (nil for collection iteration)
	Condition Expression // Optional filter condition (where clause)
}

func (ListComprehension) expr() {}

type Pipe struct {
	Left  Expression
	Right Expression // Function name to apply
}

func (Pipe) expr() {}

type ReadFile struct {
	Path Expression
}

func (ReadFile) expr() {}

type EnvVar struct {
	Name Expression
}

func (EnvVar) expr() {}

type WriteFile struct {
	Content Expression
	Path    Expression
	Append  bool
}

func (WriteFile) stmt() {}

type RunCommand struct {
	Command Expression
}

func (RunCommand) expr() {}

type TupleExpr struct {
	Elements []Expression
}

func (TupleExpr) expr() {}

type UnpackAssign struct {
	Names   []string
	Value   Expression
	Mutable bool
}

func (UnpackAssign) stmt() {}

type Using struct {
	Name string
	Expr Expression
	Body []Statement
}

func (Using) stmt() {}

type OpenFile struct {
	Path Expression
	Mode string // "read", "write", "append"
}

func (OpenFile) expr() {}

type Slice struct {
	Object Expression
	Start  Expression // nil means from beginning
	End    Expression // nil means to end
}

func (Slice) expr() {}

// Parser

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func Parse(source string) (*Program, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}

	p := New(tokens)
	return p.Parse()
}

func (p *Parser) Parse() (*Program, error) {
	prog := &Program{}

	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
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

func (p *Parser) parseTopLevel() (Statement, error) {
	if p.match(lexer.TO) {
		return p.parseFunction()
	}
	if p.match(lexer.A) {
		return p.parseStructOrMethod()
	}
	return nil, p.error("expected 'to' or 'a' at top level")
}

func (p *Parser) parseFunction() (Statement, error) {
	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.error("expected function name")
	}

	// Parse parameters (can be separated by "and")
	// Note: "a" is a keyword but can also be a parameter name
	var params []string
	for p.current().Type == lexer.IDENT || p.current().Type == lexer.A {
		params = append(params, p.current().Value)
		p.advance()
		// Skip "and" between parameters
		p.match(lexer.AND)
	}

	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after function signature")
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
		return nil, p.error("expected struct name")
	}

	if p.match(lexer.HAS) {
		return p.parseStructFields(name)
	}
	if p.match(lexer.CAN) {
		return p.parseMethod(name)
	}

	return nil, p.error("expected 'has' or 'can' after struct name")
}

func (p *Parser) parseStructFields(name string) (Statement, error) {
	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after 'has'")
	}

	p.skipNewlines()

	var fields []Field
	if p.match(lexer.INDENT) {
		for !p.check(lexer.DEDENT) && !p.isAtEnd() {
			p.skipNewlines()
			if p.check(lexer.DEDENT) {
				break
			}

			fieldName := p.current().Value
			if !p.match(lexer.IDENT) {
				return nil, p.error("expected field name")
			}
			if !p.match(lexer.AS) {
				return nil, p.error("expected 'as' after field name")
			}
			fieldType := p.current().Value
			if !p.match(lexer.IDENT) {
				return nil, p.error("expected field type")
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
		return nil, p.error("expected method name")
	}
	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after method name")
	}

	p.skipNewlines()

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Method{Struct: structName, Name: methodName, Body: body}, nil
}

func (p *Parser) parseBlock() ([]Statement, error) {
	var stmts []Statement

	if !p.match(lexer.INDENT) {
		// Single statement on same line or empty
		if p.check(lexer.NEWLINE) || p.isAtEnd() {
			return stmts, nil
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		return []Statement{stmt}, nil
	}

	for !p.check(lexer.DEDENT) && !p.isAtEnd() {
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

func (p *Parser) parseStatement() (Statement, error) {
	switch p.current().Type {
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
	case lexer.SAY:
		return p.parseSay()
	case lexer.WRITE:
		return p.parseWriteFile(false)
	case lexer.APPEND:
		return p.parseWriteFile(true)
	case lexer.USING:
		return p.parseUsing()
	case lexer.SKIP:
		p.advance()
		return Skip{}, nil
	case lexer.STOP:
		p.advance()
		return Stop{}, nil
	case lexer.IDENT:
		return p.parseIdentStatement()
	default:
		return nil, p.error(fmt.Sprintf("unexpected token %v", p.current()))
	}
}

func (p *Parser) parseIf() (Statement, error) {
	p.advance() // consume 'if'

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after if condition")
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
				return nil, p.error("expected ':' after elif condition")
			}
			p.skipNewlines()
			elifBody, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			elseifs = append(elseifs, ElseIf{Condition: elifCond, Then: elifBody})
		} else {
			if !p.match(lexer.COLON) {
				return nil, p.error("expected ':' after otherwise")
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
	p.advance() // consume 'for'

	if !p.match(lexer.EACH) {
		return nil, p.error("expected 'each' after 'for'")
	}

	varName := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.error("expected variable name")
	}

	if !p.match(lexer.IN) {
		return nil, p.error("expected 'in' after variable name")
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
		return nil, p.error("expected ':' after for each")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return ForEach{Var: varName, Start: start, End: end, Body: body}, nil
}

func (p *Parser) parseRepeat() (Statement, error) {
	p.advance() // consume 'repeat'

	if p.current().Type != lexer.INT {
		return nil, p.error("expected integer after 'repeat'")
	}
	count, _ := strconv.Atoi(p.current().Value)
	p.advance()

	if !p.match(lexer.TIMES) {
		return nil, p.error("expected 'times' after count")
	}

	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after 'times'")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Repeat{Count: count, Body: body}, nil
}

func (p *Parser) parseWhile() (Statement, error) {
	p.advance() // consume 'while'

	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after while condition")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return While{Condition: cond, Body: body}, nil
}

func (p *Parser) parseReturn() (Statement, error) {
	p.advance() // consume 'return'

	if p.check(lexer.NEWLINE) || p.check(lexer.DEDENT) || p.isAtEnd() {
		return Return{}, nil
	}

	// Parse first value using parseComparison to avoid consuming 'and'
	val, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	// Check for multiple returns: return a and b and c
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
	p.advance() // consume 'say'

	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return Say{Value: val}, nil
}

func (p *Parser) parseWriteFile(appendMode bool) (Statement, error) {
	p.advance() // consume 'write' or 'append'

	content, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.TO) {
		return nil, p.error("expected 'to' after content in write/append")
	}

	path, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return WriteFile{Content: content, Path: path, Append: appendMode}, nil
}

func (p *Parser) parseUsing() (Statement, error) {
	p.advance() // consume 'using'

	name := p.current().Value
	if !p.match(lexer.IDENT) {
		return nil, p.error("expected variable name after 'using'")
	}

	if !p.match(lexer.IS) {
		return nil, p.error("expected 'is' after variable name in using")
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.COLON) {
		return nil, p.error("expected ':' after using expression")
	}

	p.skipNewlines()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return Using{Name: name, Expr: expr, Body: body}, nil
}

func (p *Parser) parseIdentStatement() (Statement, error) {
	name := p.current().Value
	p.advance()

	// Check for unpacking: a, b is value
	if p.match(lexer.COMMA) {
		names := []string{name}
		for {
			if p.current().Type != lexer.IDENT {
				return nil, p.error("expected identifier in unpacking")
			}
			names = append(names, p.current().Value)
			p.advance()
			if !p.match(lexer.COMMA) {
				break
			}
		}

		if !p.match(lexer.IS) {
			return nil, p.error("expected 'is' after unpacking names")
		}

		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		mutable := false
		if p.match(lexer.COMMA) {
			if !p.match(lexer.CAN) {
				return nil, p.error("expected 'can' after ','")
			}
			if !p.match(lexer.CHANGE) {
				return nil, p.error("expected 'change' after 'can'")
			}
			mutable = true
		}

		return UnpackAssign{Names: names, Value: val, Mutable: mutable}, nil
	}

	if p.match(lexer.IS) {
		// Assignment
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		mutable := false
		if p.match(lexer.COMMA) {
			if !p.match(lexer.CAN) {
				return nil, p.error("expected 'can' after ','")
			}
			if !p.match(lexer.CHANGE) {
				return nil, p.error("expected 'change' after 'can'")
			}
			mutable = true
		}

		return Assignment{Name: name, Value: val, Mutable: mutable}, nil
	}

	if p.match(lexer.BECOMES) {
		// Reassignment
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return Reassign{Name: name, Value: val}, nil
	}

	// Function call or expression statement
	// Put back the identifier and parse as expression
	p.pos--
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return ExprStmt{Expr: expr}, nil
}

// Expression parsing with precedence

func (p *Parser) parseExpression() (Expression, error) {
	return p.parsePipe()
}

func (p *Parser) parsePipe() (Expression, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	for p.match(lexer.PIPE) {
		// The right side should be a function name (identifier)
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
				return nil, p.error("expected field name after 's")
			}
			expr = Access{Object: expr, Field: field}
		} else if p.match(lexer.AT) {
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			expr = Index{Object: expr, Index: idx}
		} else if p.match(lexer.FROM) {
			// Slice: items from start to end, items from start
			// Use parsePrimary for start to avoid it consuming 'to'
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
			// Slice: items to end (from beginning)
			// Only applies if expr is an identifier (not a literal)
			if _, isIdent := expr.(Ident); isIdent {
				end, err := p.parsePrimary()
				if err != nil {
					return nil, err
				}
				expr = Slice{Object: expr, Start: nil, End: end}
			} else {
				// Put back the TO and break - it's not a slice
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

	case lexer.MY:
		p.advance()
		field := p.current().Value
		if !p.match(lexer.IDENT) {
			return nil, p.error("expected field name after 'my'")
		}
		return MyAccess{Field: field}, nil

	case lexer.READ:
		p.advance() // consume 'read'
		path, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return ReadFile{Path: path}, nil

	case lexer.ENV:
		p.advance() // consume 'env'
		name, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return EnvVar{Name: name}, nil

	case lexer.RUN:
		p.advance() // consume 'run'
		cmd, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return RunCommand{Command: cmd}, nil

	case lexer.OPEN:
		p.advance() // consume 'open'
		path, err := p.parseArgument()
		if err != nil {
			return nil, err
		}
		return OpenFile{Path: path, Mode: "read"}, nil

	case lexer.IDENT, lexer.A:
		name := p.current().Value
		p.advance()

		// Check for function call with arguments
		var args []Expression
		for p.isArgStart() {
			arg, err := p.parseArgument()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			// Skip "and" between arguments
			p.match(lexer.AND)
		}

		if len(args) > 0 {
			return Call{Func: Ident{Name: name}, Args: args}, nil
		}

		return Ident{Name: name}, nil

	case lexer.LPAREN:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !p.match(lexer.RPAREN) {
			return nil, p.error("expected ')'")
		}
		return expr, nil

	case lexer.LBRACKET:
		return p.parseList()

	default:
		return nil, p.error(fmt.Sprintf("unexpected token in expression: %v", p.current()))
	}
}

func (p *Parser) parseList() (Expression, error) {
	p.advance() // consume '['

	// Empty list
	if p.match(lexer.RBRACKET) {
		return List{Elements: []Expression{}}, nil
	}

	// Parse first expression
	first, err := p.parseOr() // Use parseOr to avoid consuming 'for' keyword
	if err != nil {
		return nil, err
	}

	// Check for list comprehension: [expr for each var in ...]
	if p.match(lexer.FOR) {
		if !p.match(lexer.EACH) {
			return nil, p.error("expected 'each' after 'for' in list comprehension")
		}

		varName := p.current().Value
		if !p.match(lexer.IDENT) {
			return nil, p.error("expected variable name in list comprehension")
		}

		if !p.match(lexer.IN) {
			return nil, p.error("expected 'in' after variable in list comprehension")
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
			return nil, p.error("expected ']' at end of list comprehension")
		}

		return ListComprehension{
			Expr:      first,
			Var:       varName,
			Start:     start,
			End:       end,
			Condition: condition,
		}, nil
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
		return nil, p.error("expected ']'")
	}

	return List{Elements: elements}, nil
}

func (p *Parser) isExprStart() bool {
	switch p.current().Type {
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.YES, lexer.NO,
		lexer.IDENT, lexer.LPAREN, lexer.LBRACKET, lexer.MY, lexer.A:
		return true
	default:
		return false
	}
}

func (p *Parser) isArgStart() bool {
	switch p.current().Type {
	case lexer.INT, lexer.FLOAT, lexer.STRING, lexer.YES, lexer.NO,
		lexer.LPAREN, lexer.LBRACKET, lexer.IDENT, lexer.A:
		return true
	default:
		return false
	}
}

func (p *Parser) parseArgument() (Expression, error) {
	// Parse a single argument (not a full expression to avoid consuming operators)
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
			return nil, p.error("expected ')'")
		}
		return expr, nil
	case lexer.LBRACKET:
		return p.parseList()
	default:
		return nil, p.error("expected argument")
	}
}

// Helper methods

func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) check(t lexer.TokenType) bool {
	return p.current().Type == t
}

func (p *Parser) match(t lexer.TokenType) bool {
	if p.check(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == lexer.EOF
}

func (p *Parser) skipNewlines() {
	for p.match(lexer.NEWLINE) {
	}
}

func (p *Parser) error(msg string) error {
	tok := p.current()
	return fmt.Errorf("%d:%d: %s (got %v)", tok.Line, tok.Column, msg, tok)
}
