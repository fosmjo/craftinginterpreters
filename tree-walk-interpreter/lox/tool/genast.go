package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func defineAst(w io.Writer, baseName string, types []string) {
	fmt.Fprintf(w, "// Code generated by `go run ../tool/genast.go`; DO NOT EDIT.\n\n")
	fmt.Fprintf(w, "package parser\n\n")

	fmt.Fprintf(w, "import \"github.com/fosmjo/lox/scanner\"\n\n")

	fmt.Fprintf(w, "type %s interface {\n", baseName)
	fmt.Fprintf(w, "    Accept(%sVisitor) interface{}\n", baseName)
	fmt.Fprintf(w, "}\n\n")

	defineVisitor(w, baseName, types)

	for _, s := range types {
		defineType(w, baseName, s)
	}
}

func defineVisitor(w io.Writer, baseName string, types []string) {
	fmt.Fprintf(w, "type %sVisitor interface {\n", baseName)

	for _, s := range types {
		exprType := strings.Split(s, ":")[0]
		exprStructName := strings.TrimSpace(exprType) + baseName
		fmt.Fprintf(w, "    Visit%s(%s) interface{}\n", exprStructName, exprStructName)
	}

	fmt.Fprintf(w, "}\n\n")
}

func defineType(w io.Writer, baseName, s string) {
	tokens := strings.Split(s, ":")
	exprType := strings.TrimSpace(tokens[0])
	structName := exprType + baseName
	fmt.Fprintf(w, "type %s struct {\n", structName)

	params := strings.TrimSpace(tokens[1])
	fields := strings.Split(params, ",")

	for _, field := range fields {
		tokens := strings.Split(strings.TrimSpace(field), " ")
		fieldName, fieldType := strings.Title(tokens[0]), tokens[1]
		fmt.Fprintf(w, "    %s %s\n", fieldName, fieldType)
	}

	fmt.Fprintf(w, "}\n\n")

	// factory
	fmt.Fprintf(w, "func New%s(%s) %s {\n", structName, params, structName)
	fmt.Fprintf(w, "    return %s{\n", structName)

	for _, field := range fields {
		tokens := strings.Split(strings.TrimSpace(field), " ")
		paramName := tokens[0]
		fieldName := strings.Title(paramName)
		fmt.Fprintf(w, "        %s: %s,\n", fieldName, paramName)
	}

	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "}\n\n")

	// implement Expr
	fmt.Fprintf(w, "func (%s %s) Accept(visitor %sVisitor) interface{} {\n", strings.ToLower(baseName), structName, baseName)
	fmt.Fprintf(w, "    return visitor.Visit%s(%s)\n", structName, strings.ToLower(baseName))
	fmt.Fprintf(w, "}\n\n")
}

func gen(fileName, baseName string, types []string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	defineAst(file, baseName, types)
}

func main() {
	gen(
		"expr.go",
		"Expr",
		[]string{
			"Assign   : name scanner.Token, value Expr",
			"Binary   : left Expr, operator scanner.Token, right Expr",
			"Call     : callee Expr, paren scanner.Token, arguments []Expr",
			"Get      : object Expr, name scanner.Token",
			"Grouping : expression Expr",
			"Literal  : value interface{}",
			"Logical  : left Expr, operator scanner.Token, right Expr",
			"Set      : object Expr, name scanner.Token, value Expr",
			"This     : keyword scanner.Token",
			"Unary    : operator scanner.Token, right Expr",
			"Variable : name scanner.Token",
		},
	)
	gen(
		"stmt.go",
		"Stmt",
		[]string{
			"Block      : statements []Stmt",
			"Expression : expression Expr",
			"Class      : name scanner.Token, methods []FunctionStmt",
			"Function   : name scanner.Token, params []scanner.Token, body []Stmt",
			"If         : condition Expr, thenBranch Stmt, elseBranch Stmt",
			"Print      : expression Expr",
			"Return     : keyword scanner.Token, value Expr",
			"Var        : name scanner.Token, initializer Expr",
			"While      : condition Expr, body Stmt",
		},
	)
}
