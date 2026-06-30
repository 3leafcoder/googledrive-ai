# Installation

## Build from Source

Clone the repository and build the binary:

```bash
git clone https://github.com/pariwork/googledrive-ai.git
cd googledrive-ai
go build ./cmd/googledrive-mcp
```

This produces a `googledrive-mcp` binary in the project root.

## Verify the Build

Check that the binary was created:

```bash
ls -la googledrive-mcp
```

The binary communicates over stdio, so running it directly will appear to hang (it's waiting for MCP protocol messages on stdin). You'll connect it to an MCP client in the [MCP Client Setup](mcp-clients.md) step.

## Install to PATH (Optional)

To make the binary available system-wide:

```bash
go install ./cmd/googledrive-mcp
```

This installs to `$GOPATH/bin` (usually `~/go/bin`). Make sure this directory is in your `PATH`.
