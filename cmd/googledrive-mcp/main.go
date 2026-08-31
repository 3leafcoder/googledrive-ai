package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/pariwork/googledrive-ai/internal/auth"
	internalmcp "github.com/pariwork/googledrive-ai/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		if err := runAuth(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// Parse top-level flags for MCP server mode.
	fs := flag.NewFlagSet("googledrive-mcp", flag.ExitOnError)
	profile := fs.String("profile", "", "profile name to use for MCP server")
	fs.Parse(os.Args[1:])

	ctx := context.Background()

	config, err := auth.LoadCredentialsForProfile(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ts, err := auth.GetTokenSourceForProfile(ctx, config, *profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "authentication failed: %v\n", err)
		os.Exit(1)
	}

	pm := auth.NewProfileManager(config, *profile, ts)
	s := internalmcp.NewServer(pm)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func runAuth() error {
	// Handle "auth list" subcommand.
	if len(os.Args) > 2 && os.Args[2] == "list" {
		return runAuthList()
	}

	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	profile := fs.String("profile", "", "profile name")
	force := fs.Bool("force", false, "delete existing token before re-authenticating")
	fs.Parse(os.Args[2:])

	ctx := context.Background()

	config, err := auth.LoadCredentialsForProfile(*profile)
	if err != nil {
		return err
	}

	ts, err := auth.RunAuthFlow(ctx, config, *profile, *force)
	if err != nil {
		return err
	}

	email, err := auth.FetchUserEmail(ctx, ts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authenticated (could not fetch email: %v)\n", err)
		return nil
	}
	fmt.Fprintf(os.Stderr, "Authenticated as %s\n", email)
	return nil
}

func runAuthList() error {
	profiles, err := auth.ListProfiles()
	if err != nil {
		return fmt.Errorf("listing profiles: %w", err)
	}
	if len(profiles) == 0 {
		fmt.Fprintln(os.Stderr, "No profiles configured. Run: googledrive-mcp auth")
		return nil
	}
	for _, p := range profiles {
		email := p.Email
		if email == "" {
			email = "(email unknown)"
		}
		client := p.ClientID
		if client == "" {
			client = "(no credentials)"
		}
		fmt.Fprintf(os.Stderr, "%-20s %-30s %s\n", p.Name, email, client)
	}
	return nil
}
