package main

import (
	"flag"
	"fmt"
	"os"

	"language-betawi/internal/evaluator"
	"language-betawi/internal/lexer"
	"language-betawi/internal/object"
	"language-betawi/internal/parser"
)

func main() {
	debug := flag.Bool("debug", false, "show token stream and AST alongside execution")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: betawi [-debug] <file.bwi>")
		os.Exit(1)
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Amsyong bre! File-nya kagak ketemu/kebaca: %v\n", err)
		os.Exit(1)
	}
	src := string(data)

	if *debug {
		printTokens(src)
	}

	p := parser.New(lexer.New(src))
	program := p.ParseProgram()

	if len(p.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "%d masalah ketemu pas parsing, tong:\n\n", len(p.Errors))
		for _, e := range p.Errors {
			fmt.Fprintln(os.Stderr, "  "+e.String())
		}
		os.Exit(1)
	}

	if *debug {
		fmt.Println("=== AST ===")
		fmt.Println(program.String())
		fmt.Println("=== OUTPUT ===")
	}

	env := object.NewEnvironment()
	eval := evaluator.New()
	result := eval.Eval(program, env)

	if errObj, ok := result.(*object.Error); ok {
		fmt.Fprintln(os.Stderr, errObj.Message)
		os.Exit(1)
	}
}

func printTokens(src string) {
	l := lexer.New(src)
	fmt.Println("=== TOKEN STREAM ===")
	for {
		tok := l.NextToken()
		if tok.Type == lexer.EOF {
			break
		}
		marker := ""
		if tok.FuzzyCorrected {
			marker = fmt.Sprintf("  <- fuzzy: '%s' (%.0f%% match)", tok.OriginalWord, tok.MatchScore*100)
		}
		fmt.Printf("%-14s %-22q line:%-3d%s\n", tok.Type, tok.Literal, tok.Line, marker)
	}
	if n := len(l.Warnings); n > 0 {
		fmt.Printf("%d fuzzy correction(s) applied during lexing.\n\n", n)
	}
}
