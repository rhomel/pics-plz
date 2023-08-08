package main

import (
	"fmt"
	"os"

	"github.com/rhomel/pics-plz/pkg/server"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <source-path-to-serve>\n", os.Args[0])
		os.Exit(1)
	}
	sourcePath := os.Args[1]
	s, err := server.New(sourcePath)
	if err != nil {
		fmt.Printf("failed to initialize the server: %v", err)
		os.Exit(2)
	}
	s.Serve()
}
