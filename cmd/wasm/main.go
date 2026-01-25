package main;

import (
    "bufio"
    "syscall/js"
    "strings"
    "log"
    "github.com/arbaregni/boomslang-go/core"
)

type StringSource struct {
    name string
	buf *bufio.Reader
}
func (s StringSource) Name() string {
    return s.name;
}
func (s StringSource) ReadLine() (string, error) {
	return s.buf.ReadString('\n')
}
func MakeStringSource(name string, programText string) StringSource {
    reader := strings.NewReader(programText);
    buf := bufio.NewReader(reader);
    return StringSource { name, buf };
}


// Define the function to be exposed to javascript
func executeJSBoomslangSource(this js.Value, p []js.Value) interface{} {
    // retrive the value of the argument passed from javascript
    programText := p[0].String();
    log.Printf("retrieved program text: '%s'", programText)
    source := MakeStringSource("web browser", programText);

    // to be passed back
    programStdout := new(strings.Builder);
    programStderr := new(strings.Builder);

    // boomslang initialization
    opts := new(core.Opts);
    opts.SetDebug(core.DBG_ALL)
    opts.SetOstr(programStdout);
    opts.SetEstr(programStderr);

	env := core.MakeEnv(opts)
	core.LoadBuiltins(env)

    // evaluate the program
	rc, val := core.Run(opts, source, env)

    result := map[string]interface{} {
        "programStdout": programStdout.String(),
        "programStderr": programStderr.String(),
        "programReturncode": rc,
        "programReturnvalue": val.PrettyPrint(),
        "programShouldUnwind": val.ShouldUnwind(),
    };

	return js.ValueOf(result)
}

var keepAlive = make(chan struct{});

func main() {
    // setup wasm logging
    log.SetOutput(MakeConsoleLogger())
    log.SetFlags(log.LstdFlags | log.Lshortfile)

    // set up wasm hooks
    api := js.Global().Get("Object").New();
    api.Set("run", js.FuncOf(executeJSBoomslangSource))
    js.Global().Set("boomslang_wasm", api);

    <-keepAlive
}
