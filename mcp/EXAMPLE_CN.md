# MCP Server 使用示例

## 基本测试

### 测试 initialize 方法

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{}}}' | ./llmgit-mcp
```

### 测试 tools/list 方法

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./llmgit-mcp
```

### 测试 generate_commit_message 工具

首先需要：
1. 确保已初始化配置：`llmgit ai init <provider> <api-key> [model]`
2. 在 Git 仓库中暂存一些变更：`git add .`

然后测试：

```bash
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"generate_commit_message","arguments":{"language":"en"}}}' | ./llmgit-mcp
```

### 测试 review_commit 工具

```bash
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"review_commit","arguments":{"commit":"HEAD"}}}' | ./llmgit-mcp
```

### 测试 explain_diff 工具

```bash
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"explain_diff"}}' | ./llmgit-mcp
```

### 测试 generate_pr_description 工具

```bash
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"generate_pr_description","arguments":{"base_branch":"main","target_branch":"HEAD"}}}' | ./llmgit-mcp
```

### 测试 generate_changelog 工具

```bash
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"generate_changelog","arguments":{"format":"markdown"}}}' | ./llmgit-mcp
```

## 在 Claude Desktop 中使用

1. 构建 MCP server：
   ```bash
   cd mcp
   make build
   ```

2. 编辑 Claude Desktop 配置文件（根据你的操作系统）：
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

3. 添加配置：
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

4. 重启 Claude Desktop

5. 在 Claude Desktop 中，你可以直接使用这些工具，例如：
   - "帮我生成一个 commit message"
   - "审查一下最新的 commit"
   - "解释一下当前的代码变更"

## 注意事项

- MCP Server 通过 stdio 通信，需要从标准输入读取 JSON-RPC 请求
- 所有响应都输出到标准输出
- 错误信息输出到标准错误
- 确保在使用前已正确配置 llmgit（运行 `llmgit ai init`）

