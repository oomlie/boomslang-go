package core

import (
	"bufio"
	"strings"
	"os"
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
func (o *Opts) AddDebug (debug DebugTarget) { o.debug |= debug; }


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

func RunFile(opts *Opts, filePath string) int {
	// Open the file in read-only mode
	if !strings.HasSuffix(filePath, ".bs") {
		fmt.Fprintf(opts.estr, "Bad file extension, '%s' does not look like a boomslang file.\n", filePath)
		return (EXIT_BAD_FILE)
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
	if err != nil {
		fmt.Fprintf(opts.estr, "Error opening file '%s': %s\n", filePath, err)
		return (EXIT_BAD_FILE)
	}
	defer file.Close()
	buf := bufio.NewReader(file)

	source := FileSource{filePath,buf}


	// evaluate the program
	env := MakeEnv(opts)
	LoadBuiltins(env)

	rc, _ := Run(opts, source, env)
	return rc
}

func Run(opts *Opts, source Source, env *BsEnv) (int, BsValue) {
	log.Printf("running, std out = %p, std err = %p", opts.ostr, opts.estr);

	log.Printf("entering lexer...");
	lexer := MakeLexer(opts, source)
	tokens, err := lexer.Lex()
	if err != nil {
		log.Printf("ERROR: failed to lex: %s", err);
		fmt.Fprintf(opts.estr, "I am very sorry, but I could not understand this file due to: %v", err)
		return EXIT_LEX_FAILURE, nil
	}

	log.Printf("entering parser...");
	parser := MakeParser(opts, tokens)
	ast, err := parser.Parse()
	if err != nil {
		log.Printf("ERROR: failed to parse: %s", err);
		fmt.Fprintf(opts.estr, "I am sorry, but I simply could not understand the file you gave me: %v", err)
		return EXIT_PARSE_FAILURE, nil
	}

	if opts.debug != 0 {
		log.Printf("============================ BEGIN EVAL ===========================\n")
	}

	log.Printf("entering eval all...");
	val := EvalAll(env, ast)
	
	if val.ShouldUnwind() {
		log.Printf("ERROR: runtime error: %s", val.PrettyPrint());
		fmt.Fprintf(opts.estr, "Failure occured during runtime:\n%v\n", val.PrettyPrint())
		return EXIT_RUNTIME_FAILURE, val
	}

	return 0, val
}
