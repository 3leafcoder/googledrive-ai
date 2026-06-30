package main

import (
	"fmt"
	"os"

	internalmcp "github.com/pariwork/googledrive-ai/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := internalmcp.NewServer()
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
