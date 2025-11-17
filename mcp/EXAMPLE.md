# MCP Server Usage Examples

## Basic Testing

### Test initialize method

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}' | ./llmgit-mcp
```

### Test tools/list method

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./llmgit-mcp
```

### Test generate_commit_message tool

First, you need to:
1. Ensure configuration is initialized: `llmgit ai init <provider> <api-key> [model]`
2. Stage some changes in a Git repository: `git add .`

Then test:

```bash
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"generate_commit_message","arguments":{"language":"en"}}}' | ./llmgit-mcp
```

### Test review_commit tool

```bash
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"review_commit","arguments":{"commit":"HEAD"}}}' | ./llmgit-mcp
```

### Test explain_diff tool

```bash
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"explain_diff"}}' | ./llmgit-mcp
```

### Test generate_pr_description tool

```bash
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"generate_pr_description","arguments":{"base_branch":"main","target_branch":"HEAD"}}}' | ./llmgit-mcp
```

### Test generate_changelog tool

```bash
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"generate_changelog","arguments":{"format":"markdown"}}}' | ./llmgit-mcp
```

## Using with Claude Desktop

1. Build the MCP server:
   ```bash
   cd mcp
   make build
   ```

2. Edit the Claude Desktop configuration file (according to your operating system):
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

3. Add the configuration:
   ```json
   {
     "mcpServers": {
       "llmgit": {
         "command": "/absolute/path/to/llmgit-mcp",
         "args": []
       }
     }
   }
   ```

4. Restart Claude Desktop

5. In Claude Desktop, you can directly use these tools, for example:
   - "Generate a commit message for me"
   - "Review the latest commit"
   - "Explain the current code changes"

## Notes

- The MCP Server communicates via stdio and reads JSON-RPC requests from standard input
- All responses are output to standard output
- Error messages are output to standard error
- Ensure llmgit is properly configured before use (run `llmgit ai init`)
