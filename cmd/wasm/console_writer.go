package main

import "syscall/js"

type consoleWriter struct {
    method string // "log", "warn", "error"
}

func (w consoleWriter) Write(p []byte) (int, error) {
	console := js.Global().Get("console");
	if console.Truthy() {
		console.Call(w.method, string(p))
	}
	return len(p), nil
}

func MakeConsoleLogger() consoleWriter {
	return consoleWriter{ method: "log" }
}
