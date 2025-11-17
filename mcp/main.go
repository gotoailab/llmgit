package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gotoailab/llmgit/internal/ai"
	"github.com/gotoailab/llmgit/internal/config"
	"github.com/gotoailab/llmgit/internal/git"
	"github.com/gotoailab/llmhub"
)

// MCP Protocol Types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      *ClientInfo            `json:"clientInfo,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerCapabilities    `json:"capabilities"`
	ServerInfo      ServerInfo            `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools map[string]interface{} `json:"tools"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MCPServer struct {
	client *llmhub.Client
	cfg    *config.Config
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please run 'llmgit ai init' first\n")
		os.Exit(1)
	}

	// Create LLM client
	client, err := llmhub.NewClient(llmhub.ClientConfig{
		APIKey:   cfg.APIKey,
		Provider: llmhub.Provider(cfg.Provider),
		Model:    cfg.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating LLM client: %v\n", err)
		os.Exit(1)
	}

	server := &MCPServer{
		client: client,
		cfg:    cfg,
	}

	// Read from stdin and write to stdout
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		response := server.handleRequest(&req)
		if response != nil {
			sendResponse(response)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}
}

func (s *MCPServer) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *MCPServer) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	var params InitializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
					Data:    err.Error(),
				},
			}
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: map[string]interface{}{},
		},
		ServerInfo: ServerInfo{
			Name:    "llmgit-mcp",
			Version: "0.0.2",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *MCPServer) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	tools := []Tool{
		{
			Name:        "generate_commit_message",
			Description: "Generate a commit message based on staged changes using AI. Follows Conventional Commits format.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Language for the commit message (en or zh). Default: en",
						"enum":        []string{"en", "zh"},
					},
				},
			},
		},
		{
			Name:        "review_commit",
			Description: "Review a Git commit using AI to find potential issues and improvements.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"commit": map[string]interface{}{
						"type":        "string",
						"description": "Commit hash or reference (e.g., HEAD, HEAD~1). Default: HEAD",
					},
				},
			},
		},
		{
			Name:        "explain_diff",
			Description: "Explain code changes in plain language using AI.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file": map[string]interface{}{
						"type":        "string",
						"description": "Optional file path to explain. If not provided, explains all working directory changes.",
					},
				},
			},
		},
		{
			Name:        "generate_pr_description",
			Description: "Generate a Pull Request description based on branch differences using AI.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"base_branch": map[string]interface{}{
						"type":        "string",
						"description": "Base branch name (e.g., main, master). Default: main",
					},
					"target_branch": map[string]interface{}{
						"type":        "string",
						"description": "Target branch name. Default: HEAD",
					},
				},
			},
		},
		{
			Name:        "generate_changelog",
			Description: "Generate a CHANGELOG based on commit history using AI.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"range": map[string]interface{}{
						"type":        "string",
						"description": "Commit range (e.g., v1.0.0..HEAD). If not provided, uses commits since last tag.",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Output format: markdown, json, or yaml. Default: markdown",
						"enum":        []string{"markdown", "json", "yaml"},
					},
				},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolsListResult{
			Tools: tools,
		},
	}
}

func (s *MCPServer) handleToolsCall(req *JSONRPCRequest) *JSONRPCResponse {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	ctx := context.Background()
	var result *CallToolResult
	var err error

	switch params.Name {
	case "generate_commit_message":
		result, err = s.handleGenerateCommitMessage(ctx, params.Arguments)
	case "review_commit":
		result, err = s.handleReviewCommit(ctx, params.Arguments)
	case "explain_diff":
		result, err = s.handleExplainDiff(ctx, params.Arguments)
	case "generate_pr_description":
		result, err = s.handleGeneratePRDescription(ctx, params.Arguments)
	case "generate_changelog":
		result, err = s.handleGenerateChangelog(ctx, params.Arguments)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}

	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32000,
				Message: "Tool execution error",
				Data:    err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *MCPServer) handleGenerateCommitMessage(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	// Get language parameter
	lang := "en"
	if langVal, ok := args["language"].(string); ok {
		lang = langVal
	}

	// Get staged changes
	diff, err := git.GetStagedDiff()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff: %w", err)
	}

	if diff == "" {
		return &CallToolResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "No staged changes found. Please stage your changes first using 'git add'.",
				},
			},
		}, nil
	}

	// Get staged files for type hint
	files, err := git.GetStagedFiles()
	if err != nil {
		files = []string{}
	}
	typeHint := ai.AnalyzeChangeType(files)

	// Generate commit message
	message, err := ai.GenerateCommitMessage(ctx, s.client, diff, lang, s.cfg, typeHint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit message: %w", err)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: message,
			},
		},
	}, nil
}

func (s *MCPServer) handleReviewCommit(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	commit := "HEAD"
	if commitVal, ok := args["commit"].(string); ok {
		commit = commitVal
	}

	// Get commit information
	commitInfo, err := git.GetCommitInfo(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %w", err)
	}

	// Review commit
	review, err := ai.ReviewCommit(ctx, s.client, commitInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to review commit: %w", err)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: review,
			},
		},
	}, nil
}

func (s *MCPServer) handleExplainDiff(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	var diff string
	var err error

	if fileVal, ok := args["file"].(string); ok && fileVal != "" {
		// Explain specific file
		diff, err = git.GetFileDiff(fileVal)
	} else {
		// Explain all working directory changes
		diff, err = git.GetDiff()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	if diff == "" {
		return &CallToolResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "No changes found to explain.",
				},
			},
		}, nil
	}

	// Explain diff
	explanation, err := ai.ExplainDiff(ctx, s.client, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to explain diff: %w", err)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: explanation,
			},
		},
	}, nil
}

func (s *MCPServer) handleGeneratePRDescription(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	baseBranch := "main"
	if baseVal, ok := args["base_branch"].(string); ok {
		baseBranch = baseVal
	}

	targetBranch := "HEAD"
	if targetVal, ok := args["target_branch"].(string); ok {
		targetBranch = targetVal
	}

	// Get branch diff
	diff, err := git.GetBranchDiffFull(baseBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch diff: %w", err)
	}

	if diff == "" {
		return &CallToolResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "No changes found between branches.",
				},
			},
		}, nil
	}

	// Get commit list
	commits, err := git.GetBranchCommits(baseBranch, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch commits: %w", err)
	}

	// Generate PR description
	prDesc, err := ai.GeneratePRDescription(ctx, s.client, baseBranch, targetBranch, commits, diff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR description: %w", err)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: prDesc,
			},
		},
	}, nil
}

func (s *MCPServer) handleGenerateChangelog(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	rangeSpec := ""
	if rangeVal, ok := args["range"].(string); ok {
		rangeSpec = rangeVal
	}

	format := "markdown"
	if formatVal, ok := args["format"].(string); ok {
		format = formatVal
	}

	// Get commit log
	commitLog, err := git.GetCommitLog(rangeSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	if commitLog == "" {
		return &CallToolResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: "No commits found in the specified range.",
				},
			},
		}, nil
	}

	// Generate changelog
	changelog, err := ai.GenerateChangelog(ctx, s.client, commitLog, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate changelog: %w", err)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: changelog,
			},
		},
	}, nil
}

func sendResponse(resp *JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func sendError(id interface{}, code int, message string, data interface{}) {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	sendResponse(resp)
}

