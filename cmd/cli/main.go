package main;

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
    "github.com/arbaregi/boomslang-go/core"
)

func parse_opts() *Opts {
	opts := new(Opts)
	flag.Parse()

	positional := make([]string, 0, len(flag.Args()))

	for _, arg := range flag.Args()[1:] {
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}

		// start: parsing the arg
		if arg == "--debug" {
			opts.debug = DBG_ALL
		} else if strings.HasPrefix(arg, "--debug=") {
			elems := strings.Split(strings.TrimPrefix(arg, "--debug="), ",")
			for _, elem := range elems {
				if elem == "lex" {
					opts.debug |= DBG_LEX
				} else if elem == "parse" {
					opts.debug |= DBG_PARSE
				} else if elem == "eval" {
					opts.debug |= DBG_EVAL
				} else {
					fmt.Printf("Bad choice for --debug, '%s' not supported\n", elem)
					os.Exit(EXIT_BAD_OPTS)
				}
			}
		} else {
			fmt.Printf("Bad flag: I do not recognize %s\n", arg)
			os.Exit(EXIT_BAD_OPTS)
		}

		//end: parsing the flag args

	}

	// use the postional args

	if len(positional) > 1 {
		fmt.Printf("usage: boomslang <filename> [flags], got: %v\n", positional)
		os.Exit(EXIT_BAD_OPTS)
	}
	if len(positional) == 1 {
		opts.filePath = positional[0]
	}

	// set good defaults for other args
	opts.istr = os.Stdin
	opts.ostr = os.Stdout
	opts.estr = os.Stderr

	return opts
}


func execute(opts *Opts, filePath string) int {
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

	rc, _ := run(opts, source, env)
	return rc
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	opts := parse_opts()
	if opts.debug > 0 {
		log.Printf("debug mode, good choice...\n")
	}
	if opts.filePath != "" {
		rc := execute(opts, opts.filePath)
		os.Exit(rc)
	}
    log.Printf("no repl yet\n")
}


