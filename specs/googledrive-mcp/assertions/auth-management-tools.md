---
id: auth-management-tools
parent: googledrive-mcp
created: 2026-08-31T20:30:00Z
priority: 1
status: in_progress
locked-by: builder-MacBook-Pro.local-24854-1725134400
depends-on: oauth2-auth
---

# MCP tools expose auth status, profile listing, and profile switching

The LLM can inspect and switch Google accounts mid-session without the user leaving the conversation.

## Success Criteria

- `auth_status` tool returns: active profile name, authenticated email address, granted scopes, and token expiry timestamp
- `auth_list_profiles` tool returns all profiles with their associated Google email addresses; a profile with an expired/invalid token is marked as such rather than omitted
- `auth_switch_profile` tool accepts a profile name, swaps the active token source in-process, and all subsequent Drive/Sheets/Docs calls use the new account — no server restart required
- `auth_switch_profile` returns an error with a clear message if the requested profile does not exist or has no valid token
- When no profiles exist (fresh install), `auth_list_profiles` returns only the default profile
