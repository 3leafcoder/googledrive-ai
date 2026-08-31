---
id: headless-auth-fallback
parent: googledrive-mcp
created: 2026-08-31T20:00:00Z
priority: 1
status: in_progress
locked-by: builder-macbook-pro.local-29001-1788185726
depends-on: oauth2-auth
---

# Auth flow completes via manual code paste when callback server is unreachable

When the localhost callback cannot receive the redirect (headless, SSH, Docker, WSL, port blocked), the user can still authenticate by pasting the authorization code at a terminal prompt.

## Success Criteria

- `browserAuthFlow` races two code sources: the HTTP callback handler and a stdin line reader — first one to deliver a code wins
- After redirect, the callback HTML page displays the authorization code on screen with instructions to copy it
- CLI prints a `"Paste authorization code here: "` prompt to stderr after the auth URL
- Pasting a code at the prompt completes the OAuth2 exchange and saves the token identically to the callback path
- The existing browser callback flow is unchanged — if the redirect arrives first, stdin is never read
- Stdin reader does not block shutdown: Ctrl+C still exits cleanly
