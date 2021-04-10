package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/fosmjo/lox/scanner"
)

type Lox struct {
	hadError bool
}

func NewLox() *Lox {
	return &Lox{hadError: false}
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
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func (lox *Lox) run(source string) {
	s := scanner.New(source, lox)
	tokens := s.ScanTokens()
	for i := range tokens {
		fmt.Println(tokens[i])
	}
}

func (lox *Lox) Error(line int, msg string) {
	lox.report(line, "", msg)
}

func (lox *Lox) report(line int, where, msg string) {
	fmt.Fprintf(os.Stderr, "line [%d], Error %s : %s\n", line, where, msg)
	lox.hadError = true
}

func main() {
	lox := NewLox()

	if len(os.Args) > 2 {
		fmt.Println("Usage: lox [script]")
		os.Exit(65)
	} else if len(os.Args) == 2 {
		lox.RunFile(os.Args[1])
	} else {
		lox.RunPromt()
	}
}
