package main

import (
	"fmt"
)

type replsource struct{
	line string
}

func (s replsource) Name() string {
	return "<repl>"
}
func (s replsource) ReadLine() (string, error) {
	fmt.Printf("> ")
	_	, err := fmt.Scanln( &s.line)
	return s.line, err
}

func repl(opts *Opts) {
	env := MakeEnv(opts)
	LoadBuiltins(env)

	fmt.Fprintf(opts.ostr, "boomslang 0.1.0 >>>>\n")

	source := replsource{}

	for {
		rc, val := runrepl(opts, source, env)
		if val != nil {
			fmt.Fprintf(opts.ostr, "(%d) => %s\n", rc, val.PrettyPrint())
		}
	}
}
func runrepl(opts *Opts, source Source, env *BsEnv) (int, BsValue) {
	lexer := MakeLexer(opts, source)
	tokens, err := lexer.LexLine()
	if err != nil {
		fmt.Fprintf(opts.estr, "\033[0;31m I am very sorry, but I could not understand this file due to: %v\n\033[0m ", err)
		return EXIT_LEX_FAILURE, nil
	}

	parser := MakeParser(opts, tokens)
	ast, err := parser.parseStmnt(tokens)
	if err != nil {
		fmt.Fprintf(opts.estr, "\033[0;31m I am sorry, but I simply could not understand the file you gave me: %v\n\033[0m ", err)
		return EXIT_PARSE_FAILURE, nil
	}

	if opts.debug != 0 {
		fmt.Fprintf(opts.ostr, "============================ BEGIN EVAL ===========================\n")
	}

	val := ast.Eval(env)
	if val.ShouldUnwind() {
		fmt.Fprintf(opts.estr, "\033[0;31m Failure occured during runtime:\n%v\033[0m\n", val.PrettyPrint())
		return EXIT_RUNTIME_FAILURE, val
	}

	return 0, val

}


