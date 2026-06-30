# Configuration

## Credentials File

The server reads OAuth2 client credentials from:

```
~/.config/googledrive-ai/credentials.json
```

This file is the JSON you downloaded from Google Cloud Console during [Google Cloud Setup](google-cloud.md). If the file is missing, the server prints setup instructions and exits.

## First-Run Authentication

On first run, the server performs a browser-based OAuth2 consent flow:

1. The server starts a temporary local HTTP server on a random port
2. Your default browser opens to Google's consent screen
3. Sign in and grant the requested permissions (Drive, Sheets, Docs)
4. The browser redirects back to the local server with an authorization code
5. The server exchanges the code for access and refresh tokens
6. Tokens are saved to `~/.config/googledrive-ai/token.json`

If the browser doesn't open automatically, the server prints a URL to stderr that you can copy and paste.

## Token Caching

After initial authorization, the server uses cached tokens from:

```
~/.config/googledrive-ai/token.json
```

- Tokens are refreshed automatically when they expire
- Refreshed tokens are saved back to disk
- You should not need to re-authorize unless you revoke access or delete the token file

## Re-Authenticating

If you need to re-authorize (e.g., changed scopes or revoked access):

```bash
rm ~/.config/googledrive-ai/token.json
```

The next time the server starts, it will prompt for authorization again.

## File Permissions

The server creates config files with restricted permissions:

- `token.json` is written with `0600` (owner read/write only)
- The config directory is created with `0700` (owner access only)
