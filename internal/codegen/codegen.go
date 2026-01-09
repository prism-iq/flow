package codegen

import (
	"fmt"
	"strings"

	"flow/internal/parser"
)

type Generator struct {
	indent  int
	output  strings.Builder
	structs map[string]parser.Struct
	methods map[string][]parser.Method
}

func New() *Generator {
	return &Generator{
		structs: make(map[string]parser.Struct),
		methods: make(map[string][]parser.Method),
	}
}

func GenerateCode(prog *parser.Program) (string, error) {
	gen := New()
	return gen.Generate(prog)
}

func (g *Generator) Generate(prog *parser.Program) (string, error) {
	// First pass: collect structs and methods
	for _, stmt := range prog.Statements {
		switch s := stmt.(type) {
		case parser.Struct:
			g.structs[s.Name] = s
		case parser.Method:
			g.methods[s.Struct] = append(g.methods[s.Struct], s)
		}
	}

	// Headers
	g.writeln("#include <iostream>")
	g.writeln("#include <string>")
	g.writeln("#include <vector>")
	g.writeln("#include <type_traits>")
	g.writeln("#include <fstream>")
	g.writeln("#include <sstream>")
	g.writeln("#include <cstdlib>")
	g.writeln("")

	// Generate structs with methods
	for _, stmt := range prog.Statements {
		if s, ok := stmt.(parser.Struct); ok {
			g.genStruct(s)
		}
	}

	// Generate standalone functions
	for _, stmt := range prog.Statements {
		if f, ok := stmt.(parser.Function); ok {
			g.genFunction(f)
		}
	}

	return g.output.String(), nil
}

func (g *Generator) genStruct(s parser.Struct) {
	g.writeln("struct %s {", s.Name)
	g.indent++

	// Fields
	for _, f := range s.Fields {
		cppType := g.flowTypeToCpp(f.Type)
		g.writeln("%s %s;", cppType, f.Name)
	}

	// Methods
	if methods, ok := g.methods[s.Name]; ok {
		g.writeln("")
		for _, m := range methods {
			g.genMethod(m)
		}
	}

	g.indent--
	g.writeln("};")
	g.writeln("")
}

func (g *Generator) genMethod(m parser.Method) {
	g.writeln("void %s() {", m.Name)
	g.indent++
	for _, stmt := range m.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) genFunction(f parser.Function) {
	if f.Name == "start" {
		g.writeln("int main() {")
	} else {
		params := g.genParams(f.Params)
		g.writeln("auto %s(%s) {", f.Name, params)
	}
	g.indent++

	for _, stmt := range f.Body {
		g.genStatement(stmt)
	}

	if f.Name == "start" {
		g.writeln("return 0;")
	}
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) genParams(params []string) string {
	var parts []string
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("auto %s", p))
	}
	return strings.Join(parts, ", ")
}

func (g *Generator) genStatement(stmt parser.Statement) {
	switch s := stmt.(type) {
	case parser.If:
		g.genIf(s)
	case parser.ForEach:
		g.genForEach(s)
	case parser.Repeat:
		g.genRepeat(s)
	case parser.While:
		g.genWhile(s)
	case parser.Return:
		g.genReturn(s)
	case parser.Say:
		g.genSay(s)
	case parser.Assignment:
		g.genAssignment(s)
	case parser.Reassign:
		g.genReassign(s)
	case parser.Skip:
		g.writeln("continue;")
	case parser.Stop:
		g.writeln("break;")
	case parser.WriteFile:
		g.genWriteFile(s)
	case parser.ExprStmt:
		g.writeln("%s;", g.genExpr(s.Expr))
	}
}

func (g *Generator) genIf(s parser.If) {
	g.writeln("if (%s) {", g.genExpr(s.Condition))
	g.indent++
	for _, stmt := range s.Then {
		g.genStatement(stmt)
	}
	g.indent--

	for _, elif := range s.ElseIfs {
		g.writeln("} else if (%s) {", g.genExpr(elif.Condition))
		g.indent++
		for _, stmt := range elif.Then {
			g.genStatement(stmt)
		}
		g.indent--
	}

	if len(s.Else) > 0 {
		g.writeln("} else {")
		g.indent++
		for _, stmt := range s.Else {
			g.genStatement(stmt)
		}
		g.indent--
	}
	g.writeln("}")
}

func (g *Generator) genForEach(s parser.ForEach) {
	if s.End != nil {
		// Range loop: for each i in 1 to 10
		start := g.genExpr(s.Start)
		end := g.genExpr(s.End)
		g.writeln("for (int %s = %s; %s <= %s; %s++) {", s.Var, start, s.Var, end, s.Var)
	} else {
		// Collection loop: for each item in list
		collection := g.genExpr(s.Start)
		g.writeln("for (const auto& %s : %s) {", s.Var, collection)
	}
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) genRepeat(s parser.Repeat) {
	g.writeln("for (int _i = 0; _i < %d; _i++) {", s.Count)
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) genWhile(s parser.While) {
	g.writeln("while (%s) {", g.genExpr(s.Condition))
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
}

func (g *Generator) genReturn(s parser.Return) {
	if s.Value != nil {
		g.writeln("return %s;", g.genExpr(s.Value))
	} else {
		g.writeln("return;")
	}
}

func (g *Generator) genSay(s parser.Say) {
	g.writeln("std::cout << %s << std::endl;", g.genExpr(s.Value))
}

func (g *Generator) genAssignment(s parser.Assignment) {
	if s.Mutable {
		g.writeln("auto %s = %s;", s.Name, g.genExpr(s.Value))
	} else {
		g.writeln("const auto %s = %s;", s.Name, g.genExpr(s.Value))
	}
}

func (g *Generator) genReassign(s parser.Reassign) {
	g.writeln("%s = %s;", s.Name, g.genExpr(s.Value))
}

func (g *Generator) genWriteFile(s parser.WriteFile) {
	path := g.genExpr(s.Path)
	content := g.genExpr(s.Content)
	if s.Append {
		g.writeln("{ std::ofstream _f(%s, std::ios::app); _f << %s; }", path, content)
	} else {
		g.writeln("{ std::ofstream _f(%s); _f << %s; }", path, content)
	}
}

func (g *Generator) genExpr(expr parser.Expression) string {
	switch e := expr.(type) {
	case parser.BinaryOp:
		return fmt.Sprintf("(%s %s %s)", g.genExpr(e.Left), e.Op, g.genExpr(e.Right))
	case parser.UnaryOp:
		return fmt.Sprintf("(%s%s)", e.Op, g.genExpr(e.Value))
	case parser.IntLit:
		return fmt.Sprintf("%d", e.Value)
	case parser.FloatLit:
		return fmt.Sprintf("%v", e.Value)
	case parser.StringLit:
		return fmt.Sprintf("\"%s\"", e.Value)
	case parser.BoolLit:
		if e.Value {
			return "true"
		}
		return "false"
	case parser.Ident:
		return e.Name
	case parser.MyAccess:
		return e.Field
	case parser.Access:
		return fmt.Sprintf("%s.%s", g.genExpr(e.Object), e.Field)
	case parser.Index:
		return fmt.Sprintf("%s[%s]", g.genExpr(e.Object), g.genExpr(e.Index))
	case parser.Call:
		var args []string
		for _, arg := range e.Args {
			args = append(args, g.genExpr(arg))
		}
		return fmt.Sprintf("%s(%s)", g.genExpr(e.Func), strings.Join(args, ", "))
	case parser.List:
		var elems []string
		for _, elem := range e.Elements {
			elems = append(elems, g.genExpr(elem))
		}
		return fmt.Sprintf("{%s}", strings.Join(elems, ", "))
	case parser.ListComprehension:
		return g.genListComprehension(e)
	case parser.Pipe:
		return g.genPipe(e)
	case parser.ReadFile:
		return g.genReadFile(e)
	case parser.EnvVar:
		return g.genEnvVar(e)
	default:
		return "/* unknown expr */"
	}
}

func (g *Generator) genReadFile(rf parser.ReadFile) string {
	// Read entire file into string using stringstream
	path := g.genExpr(rf.Path)
	return fmt.Sprintf("[&]() { std::ifstream _f(%s); std::stringstream _ss; _ss << _f.rdbuf(); return _ss.str(); }()", path)
}

func (g *Generator) genEnvVar(ev parser.EnvVar) string {
	name := g.genExpr(ev.Name)
	// std::getenv returns nullptr if not found, so handle that
	return fmt.Sprintf("[&]() { const char* _v = std::getenv(%s); return _v ? std::string(_v) : std::string(); }()", name)
}

func (g *Generator) genListComprehension(lc parser.ListComprehension) string {
	// Generate an immediately-invoked lambda that builds a vector
	// Use a two-pass approach: first create empty vector, loop to fill it, return
	// For range-based (1 to 10): use int
	// For collection-based: deduce from collection's value_type

	var sb strings.Builder
	sb.WriteString("[&]() { ")

	exprStr := g.genExpr(lc.Expr)

	if lc.End != nil {
		// Range-based: for each x in 1 to 10
		// We know x is int, so we can deduce result type after first iteration
		start := g.genExpr(lc.Start)
		end := g.genExpr(lc.End)
		// Use int as loop variable, auto for result element type
		sb.WriteString("std::vector<int> _result; ")
		sb.WriteString(fmt.Sprintf("for (int %s = %s; %s <= %s; %s++) { ", lc.Var, start, lc.Var, end, lc.Var))
	} else {
		// Collection-based: for each x in items
		collection := g.genExpr(lc.Start)
		// Use auto-deduction
		sb.WriteString("std::vector<std::decay_t<decltype(*std::begin(")
		sb.WriteString(collection)
		sb.WriteString("))>> _result; ")
		sb.WriteString(fmt.Sprintf("for (const auto& %s : %s) { ", lc.Var, collection))
	}

	if lc.Condition != nil {
		sb.WriteString(fmt.Sprintf("if (%s) { ", g.genExpr(lc.Condition)))
		sb.WriteString(fmt.Sprintf("_result.push_back(%s); ", exprStr))
		sb.WriteString("} ")
	} else {
		sb.WriteString(fmt.Sprintf("_result.push_back(%s); ", exprStr))
	}

	sb.WriteString("} return _result; }()")

	return sb.String()
}

func (g *Generator) genPipe(pipe parser.Pipe) string {
	// items | double | sum → sum(double(items))
	// Recursively handle nested pipes
	left := g.genExpr(pipe.Left)
	right := g.genExpr(pipe.Right)

	// If right is just a function name (Ident), call it with left as argument
	// If right is already a call, this gets more complex - for now assume it's a simple function name
	return fmt.Sprintf("%s(%s)", right, left)
}

func (g *Generator) flowTypeToCpp(t string) string {
	switch t {
	case "text":
		return "std::string"
	case "number":
		return "int"
	case "decimal":
		return "double"
	case "bool", "yes/no":
		return "bool"
	default:
		return t
	}
}

func (g *Generator) writeln(format string, args ...interface{}) {
	indent := strings.Repeat("    ", g.indent)
	g.output.WriteString(indent)
	fmt.Fprintf(&g.output, format, args...)
	g.output.WriteString("\n")
}
