# Google Drive MCP

An MCP server that gives AI assistants (Claude, etc.) access to Google Drive, Sheets, and Docs.

Built in Go. Communicates over stdio. Formula-aware Sheets writes that don't clobber adjacent formulas.

## Tools

| Category | Tools |
|----------|-------|
| **Drive** | `drive_list`, `drive_search`, `drive_get`, `drive_create_folder`, `drive_upload`, `drive_download`, `drive_delete`, `drive_move` |
| **Sheets** | `sheets_list_sheets`, `sheets_get_metadata`, `sheets_read_range`, `sheets_create`, `sheets_add_sheet`, `sheets_write_range`, `sheets_append_rows`, `sheets_batch_update` |
| **Docs** | `docs_create`, `docs_read`, `docs_append`, `docs_replace` |
| **Auth** | `auth_status`, `auth_list_profiles`, `auth_switch_profile` |

Full tool reference: [Drive](docs/tools/drive.md) | [Sheets](docs/tools/sheets.md) | [Docs](docs/tools/docs.md)

## Quick Start

### 1. Google Cloud project

1. Create a project at [console.cloud.google.com](https://console.cloud.google.com/)
2. Enable **Google Drive API**, **Google Sheets API**, and **Google Docs API**
3. Configure an OAuth consent screen (External, add your email as test user)
4. Create **OAuth client ID** credentials (Desktop app type)
5. Download the JSON file

Detailed walkthrough: [docs/setup/google-cloud.md](docs/setup/google-cloud.md)

### 2. Install

**One-line install** (macOS/Linux) downloads the latest release for your platform
to `~/.local/bin` — a user-owned directory, so the binary can self-update without
sudo:

```bash
curl -fsSL https://raw.githubusercontent.com/3leafcoder/googledrive-ai/main/install.sh | sh
```

Override the destination with `GOOGLEDRIVE_MCP_INSTALL_DIR`, or pin a version with
`GOOGLEDRIVE_MCP_VERSION`. If `~/.local/bin` is not on your `$PATH`, the script
prints the exact line to add.

**Windows** (PowerShell) — download from the [latest release](https://github.com/3leafcoder/googledrive-ai/releases/latest):

```powershell
Invoke-WebRequest -Uri https://github.com/3leafcoder/googledrive-ai/releases/latest/download/googledrive-mcp-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\Microsoft\WindowsApps\googledrive-mcp.exe"
```

**Or build from source** (requires Go 1.25+):

```bash
git clone https://github.com/3leafcoder/googledrive-ai && cd googledrive-ai
make build   # produces ./googledrive-mcp
```

### Updating

A released binary updates itself in place from the latest GitHub release:

```bash
googledrive-mcp update          # download and install the latest release
googledrive-mcp update --check  # report whether an update is available, change nothing
```

A `dev` build (built from source without a release version) refuses to update, and
if the binary lives in a directory you don't own it fails with an actionable
message rather than escalating to sudo.

### 3. Configure credentials

```bash
mkdir -p ~/.config/googledrive-ai
mv ~/Downloads/client_secret_*.json ~/.config/googledrive-ai/credentials.json
```

### 4. Authenticate

```bash
./googledrive-mcp auth
```

Opens your browser for Google OAuth consent. Tokens are cached at `~/.config/googledrive-ai/token.json` and auto-refreshed.

For headless environments (SSH, Docker, WSL), the server falls back to a paste-the-code flow.

### 5. Connect to your MCP client

**Claude Code:**

```bash
claude mcp add googledrive-mcp /path/to/googledrive-mcp
```

**Claude Desktop** — edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "googledrive-mcp": {
      "command": "/path/to/googledrive-mcp"
    }
  }
}
```

More clients: [docs/setup/mcp-clients.md](docs/setup/mcp-clients.md)

## Multiple Accounts

Named profiles keep credentials and tokens isolated per Google account.

```bash
# Authenticate a named profile
./googledrive-mcp auth --profile work

# Start the server with that profile
./googledrive-mcp --profile work

# List all profiles
./googledrive-mcp auth list
```

Profiles are stored in `~/.config/googledrive-ai/profiles/<name>/`.

At runtime, use `auth_switch_profile` to switch accounts without restarting the server.

## Architecture

```
MCP Client (Claude Code, Claude Desktop, etc.)
    ↕ stdio
googledrive-mcp
    ↕ HTTPS
Google APIs (Drive v3, Sheets v4, Docs v1)
```

## Documentation

- [Getting Started](docs/getting-started.md)
- Setup: [Google Cloud](docs/setup/google-cloud.md) | [Installation](docs/setup/installation.md) | [Configuration](docs/setup/configuration.md) | [MCP Clients](docs/setup/mcp-clients.md)
- Tools: [Drive](docs/tools/drive.md) | [Sheets](docs/tools/sheets.md) | [Docs](docs/tools/docs.md)
