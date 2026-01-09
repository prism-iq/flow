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
	g.writeln("#include <tuple>")
	g.writeln("#include <array>")
	g.writeln("#include <memory>")
	g.writeln("#include <algorithm>")
	g.writeln("#include <functional>")
	g.writeln("")

	// Generate structs with methods
	for _, stmt := range prog.Statements {
		if s, ok := stmt.(parser.Struct); ok {
			g.genStruct(s)
		}
	}

	// Generate standalone functions and decorated functions
	for _, stmt := range prog.Statements {
		switch s := stmt.(type) {
		case parser.Function:
			g.genFunction(s)
		case parser.Decorator:
			g.genDecorator(s)
		}
	}

	return g.output.String(), nil
}

func (g *Generator) genDecorator(d parser.Decorator) {
	// Generate the original function with a different name
	origName := d.Function.Name
	implName := "_" + origName + "_impl"

	// Generate implementation function
	implFunc := d.Function
	implFunc.Name = implName
	g.genFunction(implFunc)

	// Generate wrapper that applies decorator
	// Pass the actual result to the decorator, not a thunk
	params := g.genParams(d.Function.Params)
	paramNames := strings.Join(d.Function.Params, ", ")

	g.writeln("auto %s(%s) {", origName, params)
	g.indent++
	if len(d.Function.Params) > 0 {
		g.writeln("return %s(%s(%s));", d.Name, implName, paramNames)
	} else {
		g.writeln("return %s(%s());", d.Name, implName)
	}
	g.indent--
	g.writeln("}")
	g.writeln("")
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
	isGenerator := g.hasYield(f.Body)

	if f.Name == "start" {
		g.writeln("int main() {")
	} else if isGenerator {
		// Generator function - returns vector
		params := g.genParams(f.Params)
		g.writeln("auto %s(%s) {", f.Name, params)
		g.indent++
		g.writeln("std::vector<int> _result;")
		g.writeln("auto _yield = [&](auto v) { _result.push_back(v); };")
		g.indent--
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
	} else if isGenerator {
		g.writeln("return _result;")
	}
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) hasYield(stmts []parser.Statement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case parser.Yield:
			return true
		case parser.If:
			if g.hasYield(s.Then) || g.hasYield(s.Else) {
				return true
			}
			for _, elif := range s.ElseIfs {
				if g.hasYield(elif.Then) {
					return true
				}
			}
		case parser.ForEach:
			if g.hasYield(s.Body) {
				return true
			}
		case parser.Repeat:
			if g.hasYield(s.Body) {
				return true
			}
		case parser.While:
			if g.hasYield(s.Body) {
				return true
			}
		case parser.Using:
			if g.hasYield(s.Body) {
				return true
			}
		}
	}
	return false
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
	case parser.UnpackAssign:
		g.genUnpackAssign(s)
	case parser.Using:
		g.genUsing(s)
	case parser.Yield:
		g.writeln("_yield(%s);", g.genExpr(s.Value))
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

func (g *Generator) genUnpackAssign(s parser.UnpackAssign) {
	// Use C++17 structured bindings: auto [a, b] = expr;
	names := strings.Join(s.Names, ", ")
	val := g.genExpr(s.Value)
	if s.Mutable {
		g.writeln("auto [%s] = %s;", names, val)
	} else {
		g.writeln("const auto [%s] = %s;", names, val)
	}
}

func (g *Generator) genUsing(s parser.Using) {
	// Context manager using RAII - create scoped block
	val := g.genExpr(s.Expr)
	g.writeln("{ auto %s = %s;", s.Name, val)
	g.indent++
	for _, stmt := range s.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
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
	case parser.RunCommand:
		return g.genRunCommand(e)
	case parser.TupleExpr:
		return g.genTupleExpr(e)
	case parser.OpenFile:
		return g.genOpenFile(e)
	case parser.Slice:
		return g.genSlice(e)
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

func (g *Generator) genRunCommand(rc parser.RunCommand) string {
	cmd := g.genExpr(rc.Command)
	// Execute command and capture stdout using popen
	return fmt.Sprintf("[&]() { std::string _result; std::array<char, 128> _buf; std::unique_ptr<FILE, decltype(&pclose)> _pipe(popen(%s, \"r\"), pclose); if (_pipe) { while (fgets(_buf.data(), _buf.size(), _pipe.get()) != nullptr) { _result += _buf.data(); } } return _result; }()", cmd)
}

func (g *Generator) genTupleExpr(te parser.TupleExpr) string {
	var elems []string
	for _, elem := range te.Elements {
		elems = append(elems, g.genExpr(elem))
	}
	return fmt.Sprintf("std::make_tuple(%s)", strings.Join(elems, ", "))
}

func (g *Generator) genOpenFile(of parser.OpenFile) string {
	path := g.genExpr(of.Path)
	// Return an fstream that can be used in a using block
	return fmt.Sprintf("std::fstream(%s)", path)
}

func (g *Generator) genSlice(sl parser.Slice) string {
	obj := g.genExpr(sl.Object)
	// Generate vector slice using iterators
	if sl.Start == nil && sl.End != nil {
		// items to 5 → vector from begin to begin+5
		end := g.genExpr(sl.End)
		return fmt.Sprintf("std::vector(%s.begin(), %s.begin() + %s)", obj, obj, end)
	} else if sl.Start != nil && sl.End == nil {
		// items from 2 → vector from begin+2 to end
		start := g.genExpr(sl.Start)
		return fmt.Sprintf("std::vector(%s.begin() + %s, %s.end())", obj, start, obj)
	} else if sl.Start != nil && sl.End != nil {
		// items from 1 to 5 → vector from begin+1 to begin+5
		start := g.genExpr(sl.Start)
		end := g.genExpr(sl.End)
		return fmt.Sprintf("std::vector(%s.begin() + %s, %s.begin() + %s)", obj, start, obj, end)
	}
	// No slice, just return the object
	return obj
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
