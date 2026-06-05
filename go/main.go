package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/123654lkj/zero/go/compiler"
	"github.com/123654lkj/zero/go/vm"
)

func main() {
	if len(os.Args) < 2 {
		repl()
		return
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	comp := compiler.NewCompiler()
	c := comp.Compile(string(source))
	v := vm.NewVM()
	v.RunChunk(c)
}

func repl() {
	fmt.Println("Zero v0.1.0 — Type 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)
	comp := compiler.NewCompiler()

	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", r)
				}
			}()
			c := comp.Compile(line)
			v := vm.NewVM()
			v.RunChunk(c)
		}()
	}
}
