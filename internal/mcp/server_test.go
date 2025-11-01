package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/apresai/gimage/internal/config"
)

// Helper to parse JSON params
func parseParams(jsonStr string) map[string]interface{} {
	var params map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &params)
	return params
}

func TestNewMCPServer(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test-server", "1.0.0", cfg, false)

	if server == nil {
		t.Fatal("NewMCPServer returned nil")
	}

	// Server fields are private, so we just verify it was created
	// We can test behavior through the public methods
	tool := server.GetTool("nonexistent")
	if tool != nil {
		t.Error("Expected nil for nonexistent tool")
	}
}

func TestRegisterTool(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"success": true}, nil
		},
	}

	server.RegisterTool(tool)

	if len(server.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(server.tools))
	}

	if server.tools["test_tool"].Name != "test_tool" {
		t.Error("Tool not registered with correct name")
	}
}

func TestHandleInitialize(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test-server", "1.0.0", cfg, false)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  MethodInitialize,
		Params:  map[string]interface{}{},
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error != nil {
		t.Errorf("Initialize request failed: %v", response.Error)
	}

	result := response.Result

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing or invalid serverInfo in response")
	}

	if serverInfo["name"] != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", serverInfo["name"])
	}
}

func TestHandleListTools(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	// Register test tools
	server.RegisterTool(Tool{
		Name:        "tool1",
		Description: "First tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler:     func(args map[string]interface{}) (map[string]interface{}, error) { return nil, nil },
	})
	server.RegisterTool(Tool{
		Name:        "tool2",
		Description: "Second tool",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler:     func(args map[string]interface{}) (map[string]interface{}, error) { return nil, nil },
	})

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "2",
		Method:  MethodListTools,
		Params:  map[string]interface{}{},
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error != nil {
		t.Errorf("List tools request failed: %v", response.Error)
	}

	result := response.Result

	toolsRaw, ok := result["tools"]
	if !ok {
		t.Fatal("Missing tools in response")
	}

	// Convert to slice
	var tools []interface{}
	switch v := toolsRaw.(type) {
	case []interface{}:
		tools = v
	case []map[string]interface{}:
		for _, tool := range v {
			tools = append(tools, tool)
		}
	default:
		t.Fatalf("Unexpected tools type: %T", toolsRaw)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestHandleCallTool(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	// Register a test tool
	server.RegisterTool(Tool{
		Name:        "echo",
		Description: "Echoes back input",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "string",
				},
			},
		},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"echo": args["message"],
			}, nil
		},
	})

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "3",
		Method:  MethodCallTool,
		Params:  parseParams(`{"name": "echo", "arguments": {"message": "hello"}}`),
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error != nil {
		t.Errorf("Call tool request failed: %v", response.Error)
	}

	result := response.Result

	contentRaw, ok := result["content"]
	if !ok {
		t.Fatal("Missing content in response")
	}

	// Convert to slice
	var content []interface{}
	switch v := contentRaw.(type) {
	case []interface{}:
		content = v
	case []map[string]interface{}:
		for _, item := range v {
			content = append(content, item)
		}
	default:
		t.Fatalf("Unexpected content type: %T", contentRaw)
	}

	if len(content) == 0 {
		t.Fatal("Empty content in response")
	}

	contentItem := content[0].(map[string]interface{})
	if contentItem["type"] != "text" {
		t.Errorf("Expected type 'text', got %s", contentItem["type"])
	}

	text := contentItem["text"].(string)
	if !strings.Contains(text, "hello") {
		t.Errorf("Expected response to contain 'hello', got: %s", text)
	}
}

func TestHandleCallToolNotFound(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "4",
		Method:  MethodCallTool,
		Params:  parseParams(`{"name": "nonexistent", "arguments": {}}`),
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error for nonexistent tool, got nil")
	}

	if response.Error.Code != ErrorCodeMethodNotFound {
		t.Errorf("Expected error code %d, got %d", ErrorCodeMethodNotFound, response.Error.Code)
	}
}

func TestHandleInvalidMethod(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "5",
		Method:  "invalid.method",
		Params:  parseParams(`{}`),
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error for invalid method, got nil")
	}

	if response.Error.Code != ErrorCodeMethodNotFound {
		t.Errorf("Expected error code %d, got %d", ErrorCodeMethodNotFound, response.Error.Code)
	}
}

func TestServerStartWithValidRequest(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	// Create a test input with an initialize request
	input := `{"jsonrpc":"2.0","id":"1","method":"initialize","params":{}}`
	stdin := bytes.NewBufferString(input + "\n")
	stdout := &bytes.Buffer{}

	server.stdin = stdin
	server.stdout = stdout

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine since it blocks
	done := make(chan error, 1)
	go func() {
		done <- server.Start(ctx)
	}()

	// Give server a moment to process the request
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Wait for server to finish
	<-done

	// Check that we got a response
	output := stdout.String()
	if !strings.Contains(output, "serverInfo") {
		t.Logf("Output: %s", output)
		t.Error("Expected initialize response in output")
	}
}

func TestServerStartWithInvalidJSON(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	// Create input with invalid JSON
	input := `{invalid json}`
	stdin := bytes.NewBufferString(input + "\n")
	stdout := &bytes.Buffer{}

	server.stdin = stdin
	server.stdout = stdout

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	done := make(chan error, 1)
	go func() {
		done <- server.Start(ctx)
	}()

	// Cancel context to stop server
	cancel()

	// Wait for server to finish
	err := <-done
	if err != nil && err != context.Canceled {
		t.Errorf("Server returned unexpected error: %v", err)
	}

	// Server should handle invalid JSON gracefully (log error but continue)
	// No panic or fatal error should occur
}

func TestToolHandlerError(t *testing.T) {
	cfg := &config.Config{}
	server := NewMCPServer("test", "1.0.0", cfg, false)

	// Register a tool that returns an error
	server.RegisterTool(Tool{
		Name:        "error_tool",
		Description: "A tool that errors",
		InputSchema: map[string]interface{}{"type": "object"},
		Handler: func(args map[string]interface{}) (map[string]interface{}, error) {
			return nil, &ToolError{
				Code:    ErrorCodeInternalError,
				Message: "Tool error occurred",
			}
		},
	})

	request := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "6",
		Method:  MethodCallTool,
		Params:  parseParams(`{"name": "error_tool", "arguments": {}}`),
	}

	response := server.HandleRequest(context.Background(), request)

	if response.Error == nil {
		t.Error("Expected error from tool handler, got nil")
	}

	if response.Error.Code != ErrorCodeInternalError {
		t.Errorf("Expected error code %d, got %d", ErrorCodeInternalError, response.Error.Code)
	}

	if !strings.Contains(response.Error.Message, "Tool error occurred") {
		t.Errorf("Expected error message to contain 'Tool error occurred', got: %s", response.Error.Message)
	}
}
