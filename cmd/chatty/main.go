package main

import (
	"fmt"
	"os"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		mustWrite(os.Stdout, "chatty\n\nUsage:\n  chatty serve\n  chatty version\n")
		os.Exit(0)
	}

	switch args[0] {
	case "serve":
		mustWrite(os.Stdout, "chatty serve is not implemented yet\n")
	case "version":
		mustWrite(os.Stdout, fmt.Sprintf("chatty %s (commit=%s, date=%s)\n", Version, Commit, Date))
	default:
		mustWrite(os.Stderr, fmt.Sprintf("unknown command: %s\n", args[0]))
		os.Exit(1)
	}
}

func mustWrite(f *os.File, s string) {
	if _, err := f.WriteString(s); err != nil {
		panic(err)
	}
}
