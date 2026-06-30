# MCP Client Setup

Connect the Google Drive MCP server to your preferred MCP client.

## Claude Code

Add the server using the CLI:

```bash
claude mcp add googledrive-mcp /path/to/googledrive-mcp
```

Or if you installed it to your PATH:

```bash
claude mcp add googledrive-mcp googledrive-mcp
```

The server will be available in your next Claude Code session.

## Claude Desktop

Edit the Claude Desktop configuration file:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

Add the server to the `mcpServers` section:

```json
{
  "mcpServers": {
    "googledrive-mcp": {
      "command": "/path/to/googledrive-mcp"
    }
  }
}
```

Restart Claude Desktop to pick up the change.

## Other MCP Clients

The server uses stdio transport, which is supported by all MCP clients. The general pattern is:

1. Point the client to the `googledrive-mcp` binary path
2. Set transport to `stdio`
3. No additional arguments or environment variables are required

The server handles authentication on first run via the browser flow described in [Configuration](configuration.md).
