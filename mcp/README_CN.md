# llmgit MCP Server

llmgit MCP Server 是一个基于 Model Context Protocol (MCP) 的服务器，将 llmgit 的 AI 功能暴露为 MCP 工具，供支持 MCP 的 AI 客户端使用。

## 功能特性

MCP Server 提供以下工具：

1. **generate_commit_message** - 基于暂存区变更生成 commit message
2. **review_commit** - 使用 AI 审查 Git commit
3. **explain_diff** - 用通俗易懂的语言解释代码变更
4. **generate_pr_description** - 基于分支差异生成 Pull Request 描述
5. **generate_changelog** - 基于 commit 历史生成 CHANGELOG

## 前置要求

1. 已安装并配置 llmgit：
   ```bash
   llmgit ai init <provider> <api-key> [model]
   ```

2. 配置文件位于 `~/.llmgit/config.json`

## 构建

```bash
cd mcp
go build -o llmgit-mcp main.go
```

或使用 Makefile：

```bash
cd mcp
make build
```

## 使用方法

### 作为 MCP Server 运行

MCP Server 通过标准输入输出 (stdio) 与客户端通信，使用 JSON-RPC 协议。

```bash
./llmgit-mcp
```

### 在 Claude Desktop 中配置

在 Claude Desktop 的配置文件中添加：

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

配置文件位置：
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

### 在支持 MCP 的其他客户端中使用

根据客户端的要求配置 MCP server。大多数客户端支持通过 stdio 或 HTTP 连接。

## 工具说明

### generate_commit_message

生成基于暂存区变更的 commit message。

**参数：**
- `language` (可选): 语言，`en` 或 `zh`，默认为 `en`

**示例：**
```json
{
  "name": "generate_commit_message",
  "arguments": {
    "language": "zh"
  }
}
```

### review_commit

审查指定的 Git commit。

**参数：**
- `commit` (可选): Commit hash 或引用（如 HEAD, HEAD~1），默认为 `HEAD`

**示例：**
```json
{
  "name": "review_commit",
  "arguments": {
    "commit": "HEAD~1"
  }
}
```

### explain_diff

解释代码变更的含义。

**参数：**
- `file` (可选): 要解释的文件路径。如果不提供，解释所有工作区变更

**示例：**
```json
{
  "name": "explain_diff",
  "arguments": {
    "file": "src/main.go"
  }
}
```

### generate_pr_description

生成 Pull Request 描述。

**参数：**
- `base_branch` (可选): 基础分支名，默认为 `main`
- `target_branch` (可选): 目标分支名，默认为 `HEAD`

**示例：**
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

生成 CHANGELOG。

**参数：**
- `range` (可选): Commit 范围（如 `v1.0.0..HEAD`）。如果不提供，使用自上次 tag 以来的 commits
- `format` (可选): 输出格式，`markdown`、`json` 或 `yaml`，默认为 `markdown`

**示例：**
```json
{
  "name": "generate_changelog",
  "arguments": {
    "range": "v1.0.0..HEAD",
    "format": "markdown"
  }
}
```

## MCP 协议

本服务器实现了以下 MCP 协议方法：

- `initialize` - 初始化服务器连接
- `tools/list` - 列出所有可用工具
- `tools/call` - 调用指定工具

## 错误处理

如果工具执行失败，服务器会返回包含错误信息的 JSON-RPC 错误响应。常见错误：

- 配置未初始化：请先运行 `llmgit ai init`
- 无暂存区变更：使用 `generate_commit_message` 前需要先 `git add`
- Git 操作失败：检查是否在 Git 仓库中

## 开发

MCP Server 使用 Go 1.21+ 开发，依赖 llmgit 的内部包：

- `internal/ai` - AI 功能实现
- `internal/config` - 配置管理
- `internal/git` - Git 操作封装

## License

MIT

