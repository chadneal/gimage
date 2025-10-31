package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *MCPServer) handleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	if s.verbose {
		s.logInfo("Handling method: %s", req.Method)
	}

	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case MethodListTools:
		return s.handleListTools(req)
	case MethodCallTool:
		return s.handleCallTool(ctx, req)
	default:
		return s.errorResponse(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *MCPServer) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	if s.verbose {
		s.logInfo("Initializing MCP connection")
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": ProtocolVersion,
			"serverInfo": map[string]interface{}{
				"name":    s.name,
				"version": s.version,
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		},
	}
}

func (s *MCPServer) handleListTools(req *JSONRPCRequest) *JSONRPCResponse {
	tools := make([]map[string]interface{}, 0, len(s.tools))

	for _, tool := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	if s.verbose {
		s.logInfo("Listing %d tools", len(tools))
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *MCPServer) handleCallTool(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	// Extract tool name and arguments
	name, ok := req.Params["name"].(string)
	if !ok {
		return s.errorResponse(req.ID, -32602, "Invalid params: missing tool name")
	}

	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	if s.verbose {
		s.logInfo("Calling tool: %s with arguments: %v", name, arguments)
	}

	// Find tool
	tool, exists := s.tools[name]
	if !exists {
		return s.errorResponse(req.ID, ErrorCodeMethodNotFound, fmt.Sprintf("Tool not found: %s", name))
	}

	// Execute tool
	result, err := tool.Handler(arguments)
	if err != nil {
		if s.verbose {
			s.logError("Tool execution failed: %v", err)
		}
		return s.errorResponse(req.ID, -32603, fmt.Sprintf("Tool execution failed: %v", err))
	}

	if s.verbose {
		s.logInfo("Tool executed successfully: %s", name)
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": formatToolResult(result),
				},
			},
		},
	}
}

func (s *MCPServer) errorResponse(id interface{}, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}

func formatToolResult(result map[string]interface{}) string {
	// Format result as readable JSON for LLM
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result)
	}
	return string(data)
}
