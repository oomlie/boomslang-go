package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

const (
	EXIT_BAD_OPTS int = 10 + iota
	EXIT_BAD_FILE
	EXIT_LEX_FAILURE
	EXIT_PARSE_FAILURE
	EXIT_RUNTIME_FAILURE
)

type DebugTarget int

const (
	DBG_PRE DebugTarget = 1 << iota
	DBG_LEX
	DBG_PARSE
	DBG_EVAL
)
const DBG_ALL DebugTarget = ^0

type Opts struct {
	debug    DebugTarget
	istr     io.Reader
	ostr     io.Writer
	estr     io.Writer
	filePath string
}
func (o Opts) Debug    () DebugTarget { return o.debug   ; }
func (o Opts) Istr     () io.Reader { return o.istr    ; }
func (o Opts) Ostr     () io.Writer { return o.ostr    ; }
func (o Opts) Estr     () io.Writer { return o.estr    ; }
func (o Opts) FilePath () string { return o.filePath; }
func (o *Opts) SetDebug    (debug    DebugTarget) { o.debug   = debug   ; }
func (o *Opts) SetIstr     (istr     io.Reader) { o.istr    = istr    ; }
func (o *Opts) SetOstr     (ostr     io.Writer) { o.ostr    = ostr    ; }
func (o *Opts) SetEstr     (estr     io.Writer) { o.estr    = estr    ; }
func (o *Opts) SetFilePath (filePath string) { o.filePath= filePath; }


type FileSource struct {
	filePath string
	buf *bufio.Reader
}
func (s FileSource) Name() string {
	return s.filePath
}
func (s FileSource) ReadLine() (string, error) {
	return s.buf.ReadString('\n')
}

func Run(opts *Opts, source Source, env *BsEnv) (int, BsValue) {
    fmt.Fprintf(env.ostr, "=========================== begin stdout ===================\n");
    fmt.Fprintf(env.estr, "=========================== begin stderr ===================\n");

	log.Printf("entering lexer...");
	lexer := MakeLexer(opts, source)
	tokens, err := lexer.Lex()
	if err != nil {
		log.Printf("ERROR: failed to lex: %s", err);
		fmt.Fprintf(opts.estr, "\033[0;31m I am very sorry, but I could not understand this file due to: %v\n\033[0m ", err)
		return EXIT_LEX_FAILURE, nil
	}

	log.Printf("entering parser...");
	parser := MakeParser(opts, tokens)
	ast, err := parser.Parse()
	if err != nil {
		log.Printf("ERROR: failed to parse: %s", err);
		fmt.Fprintf(opts.estr, "\033[0;31m I am sorry, but I simply could not understand the file you gave me: %v\n\033[0m ", err)
		return EXIT_PARSE_FAILURE, nil
	}

	if opts.debug != 0 {
		fmt.Fprintf(opts.ostr, "============================ BEGIN EVAL ===========================\n")
	}

	log.Printf("entering eval all...");
	val := EvalAll(env, ast)
	
	fmt.Fprintf(env.ostr, "=========================== end stdout ===================\n");
    fmt.Fprintf(env.estr, "=========================== end stderr ===================\n");

	if val.ShouldUnwind() {
		log.Printf("ERROR: runtime error: %s", val.PrettyPrint());
		fmt.Fprintf(opts.estr, "\033[0;31m Failure occured during runtime:\n%v\033[0m\n", val.PrettyPrint())
		return EXIT_RUNTIME_FAILURE, val
	}

	return 0, val
}
