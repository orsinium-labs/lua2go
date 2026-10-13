package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/yuin/gopher-lua/parse"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	stmts, err := parse.Parse(os.Stdin, "main.lua")
	if err != nil {
		return fmt.Errorf("parse AST: %v", err)
	}
	emit(os.Stdout, "package main\n\n")
	for _, stmt := range stmts {
		err := convertStmt(os.Stdout, stmt)
		if err != nil {
			return fmt.Errorf("convert %T on line %d: %v", stmt, stmt.Line(), err)
		}
	}
	return nil
}

func emit(stdout io.Writer, parts ...string) {
	_, _ = fmt.Fprint(stdout, strings.Join(parts, ""))
}
