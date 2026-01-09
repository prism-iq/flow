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

// escapeCppString escapes a string for C++ string literal output
func escapeCppString(s string) string {
	var result strings.Builder
	for _, c := range s {
		switch c {
		case '\n':
			result.WriteString("\\n")
		case '\t':
			result.WriteString("\\t")
		case '\r':
			result.WriteString("\\r")
		case '\\':
			result.WriteString("\\\\")
		case '"':
			result.WriteString("\\\"")
		case 0:
			result.WriteString("\\0")
		default:
			result.WriteRune(c)
		}
	}
	return result.String()
}

// isValidInterpolation checks if the content inside {} looks like a valid identifier/expression
func isValidInterpolation(content string) bool {
	if len(content) == 0 {
		return false
	}
	// Must start with a letter or underscore (identifier)
	first := content[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}
	// Regex quantifiers like {2,4} or {3} are NOT valid interpolation
	// They contain only digits and optional comma
	allDigitsOrComma := true
	for _, c := range content {
		if !((c >= '0' && c <= '9') || c == ',' || c == ' ') {
			allDigitsOrComma = false
			break
		}
	}
	if allDigitsOrComma {
		return false
	}
	return true
}

// genStringWithInterpolation handles string interpolation like "Hello {name}!"
func (g *Generator) genStringWithInterpolation(s string) string {
	// Check if string contains interpolation
	if !strings.Contains(s, "{") {
		return fmt.Sprintf("\"%s\"", escapeCppString(s))
	}

	// Parse and build string stream expression
	var parts []string
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			// Find closing brace
			j := i + 1
			braceCount := 1
			for j < len(s) && braceCount > 0 {
				if s[j] == '{' {
					braceCount++
				} else if s[j] == '}' {
					braceCount--
				}
				j++
			}
			if braceCount == 0 {
				// Extract content inside braces
				content := s[i+1 : j-1]
				// Check if this looks like valid interpolation
				if isValidInterpolation(content) {
					parts = append(parts, content)
				} else {
					// Not valid interpolation, keep as literal (including braces)
					parts = append(parts, fmt.Sprintf("\"%s\"", escapeCppString(s[i:j])))
				}
				i = j
			} else {
				// Unmatched brace, treat as literal
				parts = append(parts, fmt.Sprintf("\"%s\"", escapeCppString(string(s[i]))))
				i++
			}
		} else {
			// Regular text until next { or end
			j := i
			for j < len(s) && s[j] != '{' {
				j++
			}
			text := s[i:j]
			if text != "" {
				parts = append(parts, fmt.Sprintf("\"%s\"", escapeCppString(text)))
			}
			i = j
		}
	}

	// Generate concatenation with ostringstream
	if len(parts) == 1 {
		return parts[0]
	}

	// Use lambda with ostringstream for proper concatenation
	var result strings.Builder
	result.WriteString("[&]() { std::ostringstream _ss; _ss")
	for _, part := range parts {
		result.WriteString(" << ")
		result.WriteString(part)
	}
	result.WriteString("; return _ss.str(); }()")
	return result.String()
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
	g.writeln("#include <regex>")
	g.writeln("#include <thread>")
	g.writeln("#include <future>")
	g.writeln("#include <mutex>")
	g.writeln("#include <chrono>")
	g.writeln("#include <iomanip>")
	g.writeln("#include <stdexcept>")
	g.writeln("#include <ctime>")
	g.writeln("#include <cmath>")
	g.writeln("#include <numeric>")
	g.writeln("#include <random>")
	g.writeln("#include <filesystem>")
	// For networking - use POSIX sockets (available on Linux)
	g.writeln("#include <sys/socket.h>")
	g.writeln("#include <netinet/in.h>")
	g.writeln("#include <arpa/inet.h>")
	g.writeln("#include <netdb.h>")
	g.writeln("#include <unistd.h>")
	// OpenSSL for HTTPS and hashing
	g.writeln("#include <openssl/sha.h>")
	g.writeln("#include <openssl/md5.h>")
	g.writeln("#include <openssl/ssl.h>")
	g.writeln("#include <openssl/err.h>")
	g.writeln("")

	// Generate structs with methods
	for _, stmt := range prog.Statements {
		if s, ok := stmt.(parser.Struct); ok {
			g.genStruct(s)
		}
	}

	// Generate standalone functions, decorated functions, and tests
	for _, stmt := range prog.Statements {
		switch s := stmt.(type) {
		case parser.Function:
			g.genFunction(s)
		case parser.Decorator:
			g.genDecorator(s)
		case parser.Test:
			g.genTest(s)
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
	case parser.Print:
		g.genPrint(s)
	case parser.Pause:
		g.genPause(s)
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
	case parser.Log:
		g.genLog(s)
	case parser.Assert:
		g.genAssert(s)
	case parser.Try:
		g.genTry(s)
	case parser.Throw:
		g.writeln("throw std::runtime_error(%s);", g.genExpr(s.Message))
	case parser.DoTogether:
		g.genDoTogether(s)
	case parser.WebSocketSend:
		g.genWebSocketSend(s)
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

func (g *Generator) genPrint(s parser.Print) {
	g.writeln("std::cout << %s << std::flush;", g.genExpr(s.Value))
}

func (g *Generator) genPause(s parser.Pause) {
	g.writeln("std::this_thread::sleep_for(std::chrono::milliseconds(%s));", g.genExpr(s.Milliseconds))
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
		// Ensure float is formatted with decimal point for C++
		s := fmt.Sprintf("%v", e.Value)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	case parser.StringLit:
		return g.genStringWithInterpolation(e.Value)
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
		// Handle builtin functions
		if ident, ok := e.Func.(parser.Ident); ok {
			switch ident.Name {
			case "to_int":
				if len(args) == 1 {
					return fmt.Sprintf("std::stoi(%s)", args[0])
				}
			case "to_float":
				if len(args) == 1 {
					return fmt.Sprintf("std::stod(%s)", args[0])
				}
			case "to_string":
				if len(args) == 1 {
					return fmt.Sprintf("std::to_string(%s)", args[0])
				}
			case "length":
				if len(args) == 1 {
					return fmt.Sprintf("static_cast<int>(std::string(%s).size())", args[0])
				}
			case "upper":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::string s = %s; std::transform(s.begin(), s.end(), s.begin(), ::toupper); return s; }()`, args[0])
				}
			case "lower":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::string s = %s; std::transform(s.begin(), s.end(), s.begin(), ::tolower); return s; }()`, args[0])
				}
			case "trim":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::string s = %s; s.erase(0, s.find_first_not_of(" \t\n\r")); s.erase(s.find_last_not_of(" \t\n\r") + 1); return s; }()`, args[0])
				}
			case "split":
				if len(args) == 2 {
					return fmt.Sprintf(`[&]() { std::vector<std::string> result; std::string s = %s; std::string delim = %s; size_t pos = 0; while ((pos = s.find(delim)) != std::string::npos) { result.push_back(s.substr(0, pos)); s.erase(0, pos + delim.length()); } result.push_back(s); return result; }()`, args[0], args[1])
				}
			case "join":
				if len(args) == 2 {
					return fmt.Sprintf(`[&]() { std::string result; auto& items = %s; std::string sep = %s; for (size_t i = 0; i < items.size(); ++i) { if (i > 0) result += sep; result += items[i]; } return result; }()`, args[0], args[1])
				}
			case "contains":
				if len(args) == 2 {
					return fmt.Sprintf(`(std::string(%s).find(%s) != std::string::npos)`, args[0], args[1])
				}
			case "starts_with":
				if len(args) == 2 {
					return fmt.Sprintf(`(std::string(%s).rfind(%s, 0) == 0)`, args[0], args[1])
				}
			case "ends_with":
				if len(args) == 2 {
					return fmt.Sprintf(`[&]() { std::string s = %s; std::string suffix = %s; return s.size() >= suffix.size() && s.compare(s.size() - suffix.size(), suffix.size(), suffix) == 0; }()`, args[0], args[1])
				}
			case "replace_all":
				if len(args) == 3 {
					return fmt.Sprintf(`[&]() { std::string s = %s; std::string from = %s; std::string to = %s; size_t pos = 0; while ((pos = s.find(from, pos)) != std::string::npos) { s.replace(pos, from.length(), to); pos += to.length(); } return s; }()`, args[0], args[1], args[2])
				}
			// Math functions
			case "abs":
				if len(args) == 1 {
					return fmt.Sprintf("std::abs(%s)", args[0])
				}
			case "min":
				if len(args) == 2 {
					return fmt.Sprintf("std::min(%s, %s)", args[0], args[1])
				}
			case "max":
				if len(args) == 2 {
					return fmt.Sprintf("std::max(%s, %s)", args[0], args[1])
				}
			case "floor":
				if len(args) == 1 {
					return fmt.Sprintf("std::floor(%s)", args[0])
				}
			case "ceil":
				if len(args) == 1 {
					return fmt.Sprintf("std::ceil(%s)", args[0])
				}
			case "round":
				if len(args) == 1 {
					return fmt.Sprintf("std::round(%s)", args[0])
				}
			case "sqrt":
				if len(args) == 1 {
					return fmt.Sprintf("std::sqrt(%s)", args[0])
				}
			case "pow":
				if len(args) == 2 {
					return fmt.Sprintf("std::pow(%s, %s)", args[0], args[1])
				}
			case "sin":
				if len(args) == 1 {
					return fmt.Sprintf("std::sin(%s)", args[0])
				}
			case "cos":
				if len(args) == 1 {
					return fmt.Sprintf("std::cos(%s)", args[0])
				}
			case "tan":
				if len(args) == 1 {
					return fmt.Sprintf("std::tan(%s)", args[0])
				}
			case "log":
				if len(args) == 1 {
					return fmt.Sprintf("std::log(%s)", args[0])
				}
			case "log10":
				if len(args) == 1 {
					return fmt.Sprintf("std::log10(%s)", args[0])
				}
			case "exp":
				if len(args) == 1 {
					return fmt.Sprintf("std::exp(%s)", args[0])
				}
			// List functions
			case "sum":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { auto v = %s; return std::accumulate(std::begin(v), std::end(v), 0); }()`, args[0])
				}
			case "product":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { auto v = %s; return std::accumulate(std::begin(v), std::end(v), 1, std::multiplies<int>()); }()`, args[0])
				}
			case "reverse":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::vector<int> v(%s); std::reverse(v.begin(), v.end()); return v; }()`, args[0])
				}
			case "sort":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::vector<int> v(%s); std::sort(v.begin(), v.end()); return v; }()`, args[0])
				}
			case "unique":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::vector<int> v(%s); std::sort(v.begin(), v.end()); v.erase(std::unique(v.begin(), v.end()), v.end()); return v; }()`, args[0])
				}
			case "first":
				if len(args) == 1 {
					return fmt.Sprintf("*std::begin(%s)", args[0])
				}
			case "last":
				if len(args) == 1 {
					return fmt.Sprintf("*std::prev(std::end(%s))", args[0])
				}
			case "empty":
				if len(args) == 1 {
					return fmt.Sprintf("(std::begin(%s) == std::end(%s))", args[0], args[0])
				}
			// Time functions
			case "now":
				if len(args) == 0 {
					return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%Y-%m-%d %H:%M:%S"); return oss.str(); }()`
				}
			case "timestamp":
				if len(args) == 0 {
					return `std::chrono::duration_cast<std::chrono::seconds>(std::chrono::system_clock::now().time_since_epoch()).count()`
				}
			case "sleep":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::this_thread::sleep_for(std::chrono::milliseconds(%s)); return 0; }()`, args[0])
				}
			case "date":
				if len(args) == 0 {
					return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%Y-%m-%d"); return oss.str(); }()`
				}
			case "time":
				if len(args) == 0 {
					return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%H:%M:%S"); return oss.str(); }()`
				}
			// Random
			case "random":
				if len(args) == 0 {
					return `[&]() { static std::mt19937 gen(std::random_device{}()); static std::uniform_real_distribution<> dis(0.0, 1.0); return dis(gen); }()`
				}
				if len(args) == 2 {
					return fmt.Sprintf(`[&]() { static std::mt19937 gen(std::random_device{}()); std::uniform_int_distribution<> dis(%s, %s); return dis(gen); }()`, args[0], args[1])
				}
			// File system functions
			case "exists":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::exists(%s)`, args[0])
				}
			case "isfile":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::is_regular_file(%s)`, args[0])
				}
			case "isdir":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::is_directory(%s)`, args[0])
				}
			case "filesize":
				if len(args) == 1 {
					return fmt.Sprintf(`static_cast<long>(std::filesystem::file_size(%s))`, args[0])
				}
			case "listdir":
				if len(args) == 1 {
					return fmt.Sprintf(`[&]() { std::vector<std::string> result; for (const auto& entry : std::filesystem::directory_iterator(%s)) { result.push_back(entry.path().filename().string()); } return result; }()`, args[0])
				}
			case "basename":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::path(%s).filename().string()`, args[0])
				}
			case "dirname":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::path(%s).parent_path().string()`, args[0])
				}
			case "extension":
				if len(args) == 1 {
					return fmt.Sprintf(`std::filesystem::path(%s).extension().string()`, args[0])
				}
			}
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
	case parser.Ask:
		return g.genAsk(e)
	case parser.Now:
		return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%Y-%m-%d %H:%M:%S"); return oss.str(); }()`
	case parser.Today:
		return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%Y-%m-%d"); return oss.str(); }()`
	case parser.Clock:
		return `[&]() { auto now = std::chrono::system_clock::now(); auto time = std::chrono::system_clock::to_time_t(now); std::ostringstream oss; oss << std::put_time(std::localtime(&time), "%H:%M:%S"); return oss.str(); }()`
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
	case parser.Fetch:
		return g.genFetch(e)
	case parser.ParseJSON:
		return g.genParseJSON(e)
	case parser.StringifyJSON:
		return g.genStringifyJSON(e)
	case parser.RegexMatch:
		return g.genRegexMatch(e)
	case parser.RegexFindAll:
		return g.genRegexFindAll(e)
	case parser.RegexReplace:
		return g.genRegexReplace(e)
	case parser.Hash:
		return g.genHash(e)
	case parser.Wait:
		return g.genWait(e)
	case parser.WebSocketConnect:
		return g.genWebSocketConnect(e)
	default:
		return "/* unknown expr */"
	}
}

func (g *Generator) genReadFile(rf parser.ReadFile) string {
	// Read entire file into string using stringstream
	path := g.genExpr(rf.Path)
	return fmt.Sprintf("[&]() { std::ifstream _f(%s); std::stringstream _ss; _ss << _f.rdbuf(); return _ss.str(); }()", path)
}

func (g *Generator) genAsk(a parser.Ask) string {
	// Read a line from stdin, optionally displaying a prompt first
	if a.Prompt != nil {
		prompt := g.genExpr(a.Prompt)
		return fmt.Sprintf(`[&]() { std::cout << %s << std::flush; std::string _input; std::getline(std::cin, _input); return _input; }()`, prompt)
	}
	return `[&]() { std::string _input; std::getline(std::cin, _input); return _input; }()`
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

// ============= v1.0 Code Generation =============

// HTTP Fetch - simple HTTP GET using sockets
func (g *Generator) genFetch(f parser.Fetch) string {
	url := g.genExpr(f.URL)
	return fmt.Sprintf(`[&]() -> std::string {
    std::string url = %s;
    // Parse URL
    std::string host, path = "/";
    size_t pos = url.find("://");
    if (pos != std::string::npos) url = url.substr(pos + 3);
    pos = url.find("/");
    if (pos != std::string::npos) { host = url.substr(0, pos); path = url.substr(pos); }
    else { host = url; }
    // Resolve host
    struct hostent* server = gethostbyname(host.c_str());
    if (!server) return "";
    // Create socket
    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) return "";
    struct sockaddr_in serv_addr;
    memset(&serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_port = htons(80);
    memcpy(&serv_addr.sin_addr.s_addr, server->h_addr, server->h_length);
    if (connect(sockfd, (struct sockaddr*)&serv_addr, sizeof(serv_addr)) < 0) { close(sockfd); return ""; }
    // Send HTTP request
    std::string request = "GET " + path + " HTTP/1.1\r\nHost: " + host + "\r\nConnection: close\r\n\r\n";
    send(sockfd, request.c_str(), request.length(), 0);
    // Read response
    std::string response;
    char buffer[4096];
    ssize_t n;
    while ((n = recv(sockfd, buffer, sizeof(buffer)-1, 0)) > 0) { buffer[n] = 0; response += buffer; }
    close(sockfd);
    // Extract body after headers
    pos = response.find("\r\n\r\n");
    if (pos != std::string::npos) return response.substr(pos + 4);
    return response;
}()`, url)
}

// JSON parsing - simple key-value extraction (basic implementation)
func (g *Generator) genParseJSON(pj parser.ParseJSON) string {
	val := g.genExpr(pj.Value)
	// Return a simple map-like structure using vector of pairs
	return fmt.Sprintf(`[&]() -> std::vector<std::pair<std::string, std::string>> {
    std::vector<std::pair<std::string, std::string>> result;
    std::string json = %s;
    std::regex pattern(R"(\"([^\"]+)\"\s*:\s*(?:\"([^\"]*)\"|(\d+(?:\.\d+)?)|(\w+)))");
    std::smatch match;
    std::string::const_iterator searchStart(json.cbegin());
    while (std::regex_search(searchStart, json.cend(), match, pattern)) {
        std::string key = match[1];
        std::string value = match[2].matched ? match[2] : (match[3].matched ? match[3] : match[4]);
        result.push_back({key, value});
        searchStart = match.suffix().first;
    }
    return result;
}()`, val)
}

// JSON stringify - convert value to JSON string
func (g *Generator) genStringifyJSON(sj parser.StringifyJSON) string {
	val := g.genExpr(sj.Value)
	return fmt.Sprintf(`[&]() -> std::string {
    std::ostringstream oss;
    oss << %s;
    return oss.str();
}()`, val)
}

// Regex match - returns true/false
func (g *Generator) genRegexMatch(rm parser.RegexMatch) string {
	pattern := g.genExpr(rm.Pattern)
	text := g.genExpr(rm.Text)
	return fmt.Sprintf("std::regex_search(%s, std::regex(%s))", text, pattern)
}

// Regex find all - returns vector of matches
func (g *Generator) genRegexFindAll(rf parser.RegexFindAll) string {
	pattern := g.genExpr(rf.Pattern)
	text := g.genExpr(rf.Text)
	return fmt.Sprintf(`[&]() -> std::vector<std::string> {
    std::vector<std::string> results;
    std::string s = %s;
    std::regex r(%s);
    std::sregex_iterator it(s.begin(), s.end(), r), end;
    for (; it != end; ++it) results.push_back((*it)[0]);
    return results;
}()`, text, pattern)
}

// Regex replace - returns new string with replacements
func (g *Generator) genRegexReplace(rr parser.RegexReplace) string {
	pattern := g.genExpr(rr.Pattern)
	text := g.genExpr(rr.Text)
	replacement := g.genExpr(rr.Replacement)
	return fmt.Sprintf("std::regex_replace(%s, std::regex(%s), %s)", text, pattern, replacement)
}

// Hash functions using OpenSSL
func (g *Generator) genHash(h parser.Hash) string {
	val := g.genExpr(h.Value)
	switch h.Algorithm {
	case "md5":
		return fmt.Sprintf(`[&]() -> std::string {
    std::string input = %s;
    unsigned char digest[MD5_DIGEST_LENGTH];
    MD5((unsigned char*)input.c_str(), input.length(), digest);
    std::ostringstream oss;
    for (int i = 0; i < MD5_DIGEST_LENGTH; i++) oss << std::hex << std::setfill('0') << std::setw(2) << (int)digest[i];
    return oss.str();
}()`, val)
	case "sha1":
		return fmt.Sprintf(`[&]() -> std::string {
    std::string input = %s;
    unsigned char digest[SHA_DIGEST_LENGTH];
    SHA1((unsigned char*)input.c_str(), input.length(), digest);
    std::ostringstream oss;
    for (int i = 0; i < SHA_DIGEST_LENGTH; i++) oss << std::hex << std::setfill('0') << std::setw(2) << (int)digest[i];
    return oss.str();
}()`, val)
	default: // sha256
		return fmt.Sprintf(`[&]() -> std::string {
    std::string input = %s;
    unsigned char digest[SHA256_DIGEST_LENGTH];
    SHA256((unsigned char*)input.c_str(), input.length(), digest);
    std::ostringstream oss;
    for (int i = 0; i < SHA256_DIGEST_LENGTH; i++) oss << std::hex << std::setfill('0') << std::setw(2) << (int)digest[i];
    return oss.str();
}()`, val)
	}
}

// Wait - async await using std::future
func (g *Generator) genWait(w parser.Wait) string {
	expr := g.genExpr(w.Expr)
	return fmt.Sprintf("std::async(std::launch::async, [&]() { return %s; }).get()", expr)
}

// WebSocket connect - returns socket file descriptor
func (g *Generator) genWebSocketConnect(ws parser.WebSocketConnect) string {
	url := g.genExpr(ws.URL)
	return fmt.Sprintf(`[&]() -> int {
    std::string url = %s;
    std::string host;
    int port = 80;
    // Parse ws://host:port/path
    size_t pos = url.find("://");
    if (pos != std::string::npos) url = url.substr(pos + 3);
    pos = url.find(":");
    size_t pathPos = url.find("/");
    if (pos != std::string::npos && pos < pathPos) {
        host = url.substr(0, pos);
        port = std::stoi(url.substr(pos + 1, pathPos - pos - 1));
    } else {
        host = url.substr(0, pathPos);
    }
    struct hostent* server = gethostbyname(host.c_str());
    if (!server) return -1;
    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) return -1;
    struct sockaddr_in serv_addr;
    memset(&serv_addr, 0, sizeof(serv_addr));
    serv_addr.sin_family = AF_INET;
    serv_addr.sin_port = htons(port);
    memcpy(&serv_addr.sin_addr.s_addr, server->h_addr, server->h_length);
    if (connect(sockfd, (struct sockaddr*)&serv_addr, sizeof(serv_addr)) < 0) { close(sockfd); return -1; }
    return sockfd;
}()`, url)
}

// WebSocket send
func (g *Generator) genWebSocketSend(ws parser.WebSocketSend) {
	socket := g.genExpr(ws.Socket)
	msg := g.genExpr(ws.Message)
	g.writeln("{ std::string _msg = %s; send(%s, _msg.c_str(), _msg.length(), 0); }", msg, socket)
}

// Logging with levels and timestamps
func (g *Generator) genLog(l parser.Log) {
	msg := g.genExpr(l.Message)
	levelPrefix := ""
	switch l.Level {
	case "info":
		levelPrefix = "[INFO]"
	case "warn":
		levelPrefix = "[WARN]"
	case "error":
		levelPrefix = "[ERROR]"
	}
	g.writeln(`{ auto _now = std::chrono::system_clock::now(); auto _time = std::chrono::system_clock::to_time_t(_now); std::cerr << std::put_time(std::localtime(&_time), "%%Y-%%m-%%d %%H:%%M:%%S") << " %s " << %s << std::endl; }`, levelPrefix, msg)
}

// Assert with optional message
func (g *Generator) genAssert(a parser.Assert) {
	cond := g.genExpr(a.Condition)
	if a.Message != nil {
		msg := g.genExpr(a.Message)
		g.writeln(`if (!(%s)) { std::cerr << "Assertion failed: " << %s << std::endl; std::abort(); }`, cond, msg)
	} else {
		g.writeln(`if (!(%s)) { std::cerr << "Assertion failed: %s" << std::endl; std::abort(); }`, cond, cond)
	}
}

// Try/catch with exception handling
func (g *Generator) genTry(t parser.Try) {
	g.writeln("try {")
	g.indent++
	for _, stmt := range t.Body {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("} catch (const std::exception& %s) {", t.ErrName)
	g.indent++
	for _, stmt := range t.Catch {
		g.genStatement(stmt)
	}
	g.indent--
	g.writeln("}")
}

// Do together - concurrent execution using threads
func (g *Generator) genDoTogether(dt parser.DoTogether) {
	g.writeln("{")
	g.indent++
	g.writeln("std::vector<std::thread> _threads;")
	for i, stmt := range dt.Body {
		g.writeln("_threads.emplace_back([&]() {")
		g.indent++
		g.genStatement(stmt)
		g.indent--
		g.writeln("});")
		_ = i
	}
	g.writeln("for (auto& t : _threads) t.join();")
	g.indent--
	g.writeln("}")
}

// Test block generation
func (g *Generator) genTest(t parser.Test) {
	testName := t.Name
	if testName == "" {
		testName = "unnamed_test"
	}
	g.writeln("void test_%s() {", strings.ReplaceAll(testName, " ", "_"))
	g.indent++
	g.writeln(`std::cout << "Running test: %s" << std::endl;`, testName)
	for _, stmt := range t.Body {
		g.genStatement(stmt)
	}
	g.writeln(`std::cout << "Test passed: %s" << std::endl;`, testName)
	g.indent--
	g.writeln("}")
	g.writeln("")
}
