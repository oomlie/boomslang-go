package main;

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
    "github.com/arbaregni/boomslang-go/core"
)

func parse_opts() *core.Opts {
	opts := new(core.Opts)
	flag.Parse()

	positional := make([]string, 0, len(flag.Args()))

	for _, arg := range flag.Args() {
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}

		// start: parsing the arg
		if arg == "--debug" {
			opts.SetDebug(core.DBG_ALL)
		} else if strings.HasPrefix(arg, "--debug=") {
			elems := strings.Split(strings.TrimPrefix(arg, "--debug="), ",")
			for _, elem := range elems {
				if elem == "lex" {
					opts.AddDebug(core.DBG_LEX)
				} else if elem == "parse" {
					opts.AddDebug(core.DBG_PARSE)
				} else if elem == "eval" {
					opts.AddDebug(core.DBG_EVAL)
				} else {
					fmt.Printf("Bad choice for --debug, '%s' not supported\n", elem)
					os.Exit(core.EXIT_BAD_OPTS)
				}
			}
		} else {
			fmt.Printf("Bad flag: I do not recognize %s\n", arg)
			os.Exit(core.EXIT_BAD_OPTS)
		}

		//end: parsing the flag args

	}

	// use the postional args

	if len(positional) > 1 {
		fmt.Printf("usage: boomslang <filename> [flags], got: %v\n", positional)
		os.Exit(core.EXIT_BAD_OPTS)
	}
	if len(positional) == 1 {
		opts.SetFilePath(positional[0])
	}

	// set good defaults for other args
	opts.SetIstr(os.Stdin)
	opts.SetOstr(os.Stdout)
	opts.SetEstr(os.Stderr)

	return opts
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	opts := parse_opts()
	if opts.Debug() > 0 {
		log.Printf("debug mode, good choice...\n")
	}
	if opts.FilePath() != "" {
		rc := core.RunFile(opts, opts.FilePath())
		os.Exit(rc)
	}
    log.Printf("no repl yet\n")
}


