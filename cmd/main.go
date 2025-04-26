package main

import (
	"context"
	"fmt"
	"os"
	"wonk/cmd/server"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx, os.Getenv, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
