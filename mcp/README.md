# llmgit MCP Server

llmgit MCP Server is a Model Context Protocol (MCP) server that exposes llmgit's AI capabilities as MCP tools for use with MCP-compatible AI clients.

## Features

The MCP Server provides the following tools:

1. **generate_commit_message** - Generate commit messages based on staged changes
2. **review_commit** - Review Git commits using AI
3. **explain_diff** - Explain code changes in plain language
4. **generate_pr_description** - Generate Pull Request descriptions based on branch differences
5. **generate_changelog** - Generate CHANGELOG based on commit history

## Prerequisites

1. llmgit must be installed and configured:
   ```bash
   llmgit ai init <provider> <api-key> [model]
   ```

2. Configuration file is located at `~/.llmgit/config.json`

## Building

```bash
cd mcp
go build -o llmgit-mcp main.go
```

Or use the Makefile:

```bash
cd mcp
make build
```

## Usage

### Running as MCP Server

The MCP Server communicates with clients via standard input/output (stdio) using the JSON-RPC protocol.

```bash
./llmgit-mcp
```

### Configuring in Claude Desktop

Add the following to your Claude Desktop configuration file:

```json
{
  "mcpServers": {
    "llmgit": {
      "command": "/path/to/llmgit-mcp",
      "args": []
    }
  }
}
```

Configuration file locations:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

### Using with Other MCP-Compatible Clients

Configure the MCP server according to your client's requirements. Most clients support connections via stdio or HTTP.

## Tool Documentation

### generate_commit_message

Generates a commit message based on staged changes.

**Parameters:**
- `language` (optional): Language for the commit message, `en` or `zh`. Default: `en`

**Example:**
```json
{
  "name": "generate_commit_message",
  "arguments": {
    "language": "zh"
  }
}
```

### review_commit

Reviews a specified Git commit.

**Parameters:**
- `commit` (optional): Commit hash or reference (e.g., HEAD, HEAD~1). Default: `HEAD`

**Example:**
```json
{
  "name": "review_commit",
  "arguments": {
    "commit": "HEAD~1"
  }
}
```

### explain_diff

Explains the meaning of code changes.

**Parameters:**
- `file` (optional): File path to explain. If not provided, explains all working directory changes

**Example:**
```json
{
  "name": "explain_diff",
  "arguments": {
    "file": "src/main.go"
  }
}
```

### generate_pr_description

Generates a Pull Request description.

**Parameters:**
- `base_branch` (optional): Base branch name. Default: `main`
- `target_branch` (optional): Target branch name. Default: `HEAD`

**Example:**
```json
{
  "name": "generate_pr_description",
  "arguments": {
    "base_branch": "main",
    "target_branch": "feature-branch"
  }
}
```

### generate_changelog

Generates a CHANGELOG.

**Parameters:**
- `range` (optional): Commit range (e.g., `v1.0.0..HEAD`). If not provided, uses commits since the last tag
- `format` (optional): Output format: `markdown`, `json`, or `yaml`. Default: `markdown`

**Example:**
```json
{
  "name": "generate_changelog",
  "arguments": {
    "range": "v1.0.0..HEAD",
    "format": "markdown"
  }
}
```

## MCP Protocol

This server implements the following MCP protocol methods:

- `initialize` - Initialize server connection
- `tools/list` - List all available tools
- `tools/call` - Call a specified tool

## Error Handling

If tool execution fails, the server returns a JSON-RPC error response with error information. Common errors:

- Configuration not initialized: Please run `llmgit ai init` first
- No staged changes: Use `git add` before using `generate_commit_message`
- Git operation failed: Check if you're in a Git repository

## Development

The MCP Server is developed using Go 1.21+ and depends on llmgit's internal packages:

- `internal/ai` - AI functionality implementation
- `internal/config` - Configuration management
- `internal/git` - Git operations wrapper

## License

MIT
