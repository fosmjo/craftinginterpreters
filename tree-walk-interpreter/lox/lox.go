package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fosmjo/lox/interpreter"
	"github.com/fosmjo/lox/parser"
	"github.com/fosmjo/lox/resolver"
	"github.com/fosmjo/lox/scanner"
)

type Lox struct {
	hadError        bool
	hadRuntimeError bool
	interpreter     *interpreter.Interpreter
}

func NewLox() *Lox {
	return &Lox{
		hadError:        false,
		hadRuntimeError: false,
	}
}

func (lox *Lox) RunFile(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	lox.run(string(data))
	if lox.hadError {
		os.Exit(65)
	}
	if lox.hadRuntimeError {
		os.Exit(70)
	}
}

func (lox *Lox) RunPromt() {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		lox.run(line)
		lox.hadError = false
		lox.hadRuntimeError = false
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func (lox *Lox) run(source string) {
	s := scanner.New(source, lox)
	tokens := s.ScanTokens()
	parser := parser.NewParser(tokens, lox)

	statements, err := parser.Parse()
	if err != nil {
		return
	}

	resolver := resolver.NewResolver(lox.interpreter, lox)
	resolver.Resolve(statements)

	// Stop if there was a resolution error.
	if lox.hadError {
		return
	}

	lox.interpreter.Interpret(statements)
}

func (lox *Lox) Error(line int, msg string) {
	lox.report(line, "", msg)
}

func (lox *Lox) ErrorWithToken(token scanner.Token, msg string) {
	if token.Type == scanner.EOF {
		lox.report(token.Line, " at end", msg)
	} else {
		lox.report(token.Line, " at '"+token.Lexeme+"'", msg)
	}
}

func (lox *Lox) RuntimeError(err interpreter.RuntimeError) {
	fmt.Fprintln(os.Stderr, err.Error()+"\n[line "+strconv.Itoa(err.Token().Line)+"]")
	lox.hadRuntimeError = true
}

func (lox *Lox) report(line int, where, msg string) {
	fmt.Fprintf(os.Stderr, "line [%d], Error %s : %s\n", line, where, msg)
	lox.hadError = true
}

func main() {
	lox := NewLox()
	lox.interpreter = interpreter.NewInterpreter(lox)

	if len(os.Args) > 2 {
		fmt.Println("Usage: lox [script]")
		os.Exit(65)
	} else if len(os.Args) == 2 {
		lox.RunFile(os.Args[1])
	} else {
		lox.RunPromt()
	}
}
