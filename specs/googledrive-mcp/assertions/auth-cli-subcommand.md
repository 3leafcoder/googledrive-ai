---
id: auth-cli-subcommand
parent: googledrive-mcp
created: 2026-08-31T20:30:00Z
priority: 1
status: not_started
depends-on: headless-auth-fallback
---

# CLI `auth` subcommand handles interactive OAuth and profile management

Browser consent cannot happen through MCP, so a CLI subcommand handles initial auth and re-auth for each profile. Each profile is fully isolated with its own `credentials.json` and `token.json` to support different Google Workspace orgs that require separate approved OAuth clients.

## Success Criteria

- Each profile directory contains both `credentials.json` and `token.json`: `~/.config/googledrive-ai/profiles/<name>/credentials.json` and `token.json`
- `googledrive-mcp auth --profile <name>` reads credentials from that profile's `credentials.json` and saves the token to that profile's `token.json`
- `googledrive-mcp auth --profile <name>` exits with a clear error if the profile's `credentials.json` does not exist, with instructions to place it at the expected path
- `googledrive-mcp auth` runs the OAuth2 browser flow for the default profile and exits after saving the token
- `googledrive-mcp auth --force` deletes the existing token before re-running the OAuth flow
- After successful auth, prints the authenticated email to stderr (fetched from Google userinfo endpoint)
- `googledrive-mcp auth list` prints all profiles with their associated email addresses and which OAuth client each uses
- Running `googledrive-mcp` with no subcommand starts the MCP server as before — backwards compatible
- Default profile (no `--profile` flag) uses the existing `~/.config/googledrive-ai/credentials.json` and `token.json` locations — no migration needed
- `googledrive-mcp --profile <name>` starts the MCP server using that profile's credentials and token
