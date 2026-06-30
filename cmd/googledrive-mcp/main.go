package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pariwork/googledrive-ai/internal/auth"
	internalmcp "github.com/pariwork/googledrive-ai/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ctx := context.Background()

	config, err := auth.LoadCredentials()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ts, err := auth.GetTokenSource(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "authentication failed: %v\n", err)
		os.Exit(1)
	}

	s := internalmcp.NewServer(ts)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
