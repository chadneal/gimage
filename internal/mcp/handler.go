package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chadneal/gimage/internal/observability"
)

// HandleRequest processes an MCP JSON-RPC request and returns a response
// This method is exported for testing purposes
func (s *MCPServer) HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")

	logger.Debug().
		Str("method", req.Method).
		Msg("Handling request")

	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(ctx, req)
	case MethodListTools:
		return s.handleListTools(ctx, req)
	case MethodCallTool:
		return s.handleCallTool(ctx, req)
	case MethodListPrompts:
		return s.handleListPrompts(ctx, req)
	case MethodListResources:
		return s.handleListResources(ctx, req)
	default:
		logger.Warn().
			Str("method", req.Method).
			Msg("Method not found")
		return s.errorResponse(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *MCPServer) handleInitialize(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")

	logger.Info().
		Str("protocol_version", ProtocolVersion).
		Str("server_name", s.name).
		Str("server_version", s.version).
		Msg("Initializing MCP connection")

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
				"tools": map[string]interface{}{
					"listChanged": true, // Notify when tool list changes
				},
			},
		},
	}
}

func (s *MCPServer) handleListTools(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")

	tools := make([]map[string]interface{}, 0, len(s.tools))

	for _, tool := range s.tools {
		toolInfo := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}

		// Include annotations if present (MCP spec 2025-06-18)
		if tool.Annotations != nil {
			toolInfo["annotations"] = tool.Annotations
		}

		tools = append(tools, toolInfo)
	}

	logger.Debug().
		Int("tools_count", len(tools)).
		Msg("Listing tools")

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *MCPServer) handleCallTool(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")
	metrics := observability.GetMetrics()

	// Extract tool name and arguments
	name, ok := req.Params["name"].(string)
	if !ok {
		logger.Warn().Msg("Invalid params: missing tool name")
		return s.errorResponse(req.ID, -32602, "Invalid params: missing tool name")
	}

	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	logger.Info().
		Str("tool", name).
		Interface("arguments", arguments).
		Msg("Calling tool")

	// Find tool
	tool, exists := s.tools[name]
	if !exists {
		logger.Warn().
			Str("tool", name).
			Msg("Tool not found")
		return s.errorResponse(req.ID, ErrorCodeMethodNotFound, fmt.Sprintf("Tool not found: %s", name))
	}

	// Execute tool and track metrics
	startTime := time.Now()
	result, err := tool.Handler(arguments)
	duration := time.Since(startTime)

	// Record metrics
	success := err == nil
	metrics.RecordToolInvocation(ctx, name, duration, success)

	if err != nil {
		logger.Error().
			Err(err).
			Str("tool", name).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Tool execution failed")
		return s.errorResponse(req.ID, -32603, fmt.Sprintf("Tool execution failed: %v", err))
	}

	logger.Info().
		Str("tool", name).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("Tool executed successfully")

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

func (s *MCPServer) handleListPrompts(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")

	// gimage doesn't expose prompts, return empty list
	logger.Debug().Msg("Listing prompts (gimage has none)")

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"prompts": []interface{}{},
		},
	}
}

func (s *MCPServer) handleListResources(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	logger := observability.LoggerWithComponent(ctx, "mcp-handler")

	// gimage doesn't expose resources, return empty list
	logger.Debug().Msg("Listing resources (gimage has none)")

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"resources": []interface{}{},
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
