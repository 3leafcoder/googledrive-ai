package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/3leafcoder/googledrive-ai/internal/auth"
	internalmcp "github.com/3leafcoder/googledrive-ai/internal/mcp"
	"github.com/3leafcoder/googledrive-ai/internal/update"
	verpkg "github.com/3leafcoder/googledrive-ai/internal/version"

	"github.com/mark3labs/mcp-go/server"
)

// version is the build version, injected at compile time via
// -ldflags "-X main.version=...". A plain `go build` leaves it at "dev".
var version = "dev"

func main() {
	// Publish the ldflags-injected version to the shared package so other
	// packages (self-update) can read it without importing main.
	verpkg.SetCurrent(version)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth":
			if err := runAuth(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		case "update":
			os.Exit(updateCmd(os.Args[2:]))
		case "version", "--version", "-v":
			fmt.Println(version)
			return
		}
	}

	// Parse top-level flags for MCP server mode.
	fs := flag.NewFlagSet("googledrive-mcp", flag.ExitOnError)
	profile := fs.String("profile", "", "profile name to use for MCP server")
	fs.Parse(os.Args[1:])

	resolved := auth.ResolveProfile(*profile)

	ctx := context.Background()

	config, err := auth.LoadCredentialsForProfile(resolved)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ts, err := auth.GetTokenSourceForProfile(ctx, config, resolved)
	if err != nil {
		fmt.Fprintf(os.Stderr, "authentication failed: %v\n", err)
		os.Exit(1)
	}

	pm := auth.NewProfileManager(config, resolved, ts)
	s := internalmcp.NewServer(pm)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// updateCmd self-updates the binary from the latest public GitHub release. It
// prints "Already on latest version" when nothing newer exists and the
// before → after version on a successful upgrade. The HTTP client carries a
// generous timeout that covers the asset download.
//
// With --check it only previews availability: it queries the Releases API and
// compares versions without downloading an asset or writing to disk, and exits 0
// whether or not an update is available.
func updateCmd(args []string) int {
	flags := flag.NewFlagSet("update", flag.ContinueOnError)
	check := flags.Bool("check", false, "report whether an update is available without installing")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	run := update.Run
	if *check {
		run = update.RunCheck
	}
	if err := run(client, version, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runAuth() error {
	// Handle "auth list" subcommand.
	if len(os.Args) > 2 && os.Args[2] == "list" {
		return runAuthList()
	}
	// Handle "auth default" subcommand.
	if len(os.Args) > 2 && os.Args[2] == "default" {
		return runAuthDefault()
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

func runAuthDefault() error {
	if len(os.Args) > 3 {
		name := os.Args[3]
		if err := auth.SaveDefaultProfile(name); err != nil {
			return fmt.Errorf("saving default profile: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Default profile set to: %s\n", name)
		return nil
	}
	name, err := auth.LoadDefaultProfile()
	if err != nil || name == "" {
		fmt.Fprintln(os.Stderr, "No default profile set (will use first available)")
		return nil
	}
	fmt.Fprintln(os.Stderr, name)
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
		marker := "  "
		if p.IsDefault {
			marker = "* "
		}
		email := p.Email
		if email == "" {
			email = "(email unknown)"
		}
		client := p.ClientID
		if client == "" {
			client = "(no credentials)"
		}
		fmt.Fprintf(os.Stderr, "%s%-20s %-30s %s\n", marker, p.Name, email, client)
	}
	return nil
}
