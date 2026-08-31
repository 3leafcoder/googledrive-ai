---
id: auth-cli-subcommand
parent: googledrive-mcp
created: 2026-08-31T20:30:00Z
priority: 1
status: done
depends-on: headless-auth-fallback
---

# CLI `auth` subcommand handles interactive OAuth and profile management

Browser consent cannot happen through MCP, so a CLI subcommand handles initial auth and re-auth for each profile.

## Success Criteria

- `googledrive-mcp auth` runs the OAuth2 browser flow for the default profile and exits after saving the token
- `googledrive-mcp auth --profile <name>` runs the OAuth2 flow and saves the token under `~/.config/googledrive-ai/profiles/<name>/token.json`
- `googledrive-mcp auth --force` deletes the existing token before re-running the OAuth flow
- After successful auth, prints the authenticated email to stderr (fetched from Google userinfo endpoint)
- `googledrive-mcp auth list` prints all profiles with their associated email addresses
- Running `googledrive-mcp` with no subcommand starts the MCP server as before — backwards compatible
- Default profile (no `--profile` flag) uses the existing `~/.config/googledrive-ai/token.json` location — no migration needed
