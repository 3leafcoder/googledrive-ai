# Google Cloud Setup

This guide walks you through creating a Google Cloud project and obtaining OAuth2 credentials for the MCP server.

## 1. Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown at the top of the page
3. Click **New Project**
4. Enter a project name (e.g., "Google Drive MCP")
5. Click **Create**

## 2. Enable APIs

Enable the three APIs the server needs:

1. Go to **APIs & Services > Library**
2. Search for and enable each of these:
   - **Google Drive API**
   - **Google Sheets API**
   - **Google Docs API**

Alternatively, use the direct links:

- [Enable Drive API](https://console.cloud.google.com/apis/library/drive.googleapis.com)
- [Enable Sheets API](https://console.cloud.google.com/apis/library/sheets.googleapis.com)
- [Enable Docs API](https://console.cloud.google.com/apis/library/docs.googleapis.com)

## 3. Configure OAuth Consent Screen

1. Go to **APIs & Services > OAuth consent screen**
2. Select **External** user type (or Internal if using Google Workspace)
3. Fill in the required fields:
   - **App name:** Google Drive MCP
   - **User support email:** your email
   - **Developer contact:** your email
4. Click **Save and Continue**
5. On the **Scopes** page, click **Add or Remove Scopes** and add:
   - `https://www.googleapis.com/auth/drive`
   - `https://www.googleapis.com/auth/spreadsheets`
   - `https://www.googleapis.com/auth/documents`
6. Click **Save and Continue**
7. On the **Test users** page, add your Google account email
8. Click **Save and Continue**

## 4. Create OAuth2 Credentials

1. Go to **APIs & Services > Credentials**
2. Click **Create Credentials > OAuth client ID**
3. Select **Desktop app** as the application type
4. Enter a name (e.g., "googledrive-mcp")
5. Click **Create**
6. Click **Download JSON** to download the credentials file

## 5. Save the Credentials File

Move the downloaded file to the configuration directory:

```bash
mkdir -p ~/.config/googledrive-ai
mv ~/Downloads/client_secret_*.json ~/.config/googledrive-ai/credentials.json
```

The server expects the file at `~/.config/googledrive-ai/credentials.json`.
