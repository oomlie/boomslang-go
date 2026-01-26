package main

import (
	"fmt"
	"github.com/arbaregni/boomslang-go/core"
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

func repl(opts *core.Opts) {
	env := core.MakeEnv(opts)
	core.LoadBuiltins(env)

	fmt.Fprintf(opts.Ostr(), "boomslang 0.1.0 >>>>\n")

	source := replsource{}

	for {
		rc, val := runrepl(opts, source, env)
		if val != nil {
			fmt.Fprintf(opts.Ostr(), "(%d) => %s\n", rc, val.PrettyPrint())
		}
	}
}
func runrepl(opts *core.Opts, source core.Source, env *core.BsEnv) (int, core.BsValue) {
	lexer := core.MakeLexer(opts, source)
	tokens, err := lexer.LexLine()
	if err != nil {
		fmt.Fprintf(opts.Estr(), "\033[0;31m I am very sorry, but I could not understand this file due to: %v\n\033[0m ", err)
		return core.EXIT_LEX_FAILURE, nil
	}

	parser := core.MakeParser(opts, tokens)
	ast, err := parser.ParseStmnt(tokens)
	if err != nil {
		fmt.Fprintf(opts.Estr(), "\033[0;31m I am sorry, but I simply could not understand the file you gave me: %v\n\033[0m ", err)
		return core.EXIT_PARSE_FAILURE, nil
	}

	if opts.Debug() != 0 {
		fmt.Fprintf(opts.Ostr(), "============================ BEGIN EVAL ===========================\n")
	}

	val := ast.Eval(env)
	if val.ShouldUnwind() {
		fmt.Fprintf(opts.Estr(), "\033[0;31m Failure occured during runtime:\n%v\033[0m\n", val.PrettyPrint())
		return core.EXIT_RUNTIME_FAILURE, val
	}

	return 0, val

}


