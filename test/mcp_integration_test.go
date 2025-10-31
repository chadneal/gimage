package test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chadneal/gimage/internal/config"
	"github.com/chadneal/gimage/internal/mcp"
	"github.com/chadneal/gimage/internal/mcp/tools"
)

// TestMCPServerIntegration tests the full MCP server flow from request to response
func TestMCPServerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{}

	// Create a test image
	testImagePath := filepath.Join(tmpDir, "test.png")
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	file, err := os.Create(testImagePath)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	// Create and setup server
	server := mcp.NewMCPServer("test-server", "1.0.0", cfg, false)

	// Register all tools
	tools.RegisterResizeImageTool(server)
	tools.RegisterScaleImageTool(server)
	tools.RegisterCompressImageTool(server)
	tools.RegisterListModelsTool(server)

	// Test 1: Initialize
	t.Run("initialize", func(t *testing.T) {
		initRequest := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"init-1"`),
			Method:  mcp.MethodInitialize,
			Params:  json.RawMessage(`{}`),
		}

		requestBytes, _ := json.Marshal(initRequest)
		stdin := bytes.NewBuffer(requestBytes)
		stdin.WriteString("\n")
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

		// Cancel to stop server
		cancel()
		<-done

		// Parse response
		output := stdout.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) == 0 {
			t.Fatal("No response received")
		}

		var response mcp.JSONRPCResponse
		if err := json.Unmarshal([]byte(lines[0]), &response); err != nil {
			t.Fatalf("Failed to parse response: %v\nOutput: %s", err, output)
		}

		if response.Error != nil {
			t.Fatalf("Initialize failed: %v", response.Error)
		}

		var result map[string]interface{}
		json.Unmarshal(response.Result, &result)

		serverInfo := result["serverInfo"].(map[string]interface{})
		if serverInfo["name"] != "test-server" {
			t.Errorf("Expected server name 'test-server', got %s", serverInfo["name"])
		}
	})

	// Test 2: List Tools
	t.Run("list_tools", func(t *testing.T) {
		listRequest := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"list-1"`),
			Method:  mcp.MethodListTools,
			Params:  json.RawMessage(`{}`),
		}

		requestBytes, _ := json.Marshal(listRequest)
		stdin := bytes.NewBuffer(requestBytes)
		stdin.WriteString("\n")
		stdout := &bytes.Buffer{}

		server.stdin = stdin
		server.stdout = stdout

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- server.Start(ctx)
		}()

		cancel()
		<-done

		output := stdout.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		var response mcp.JSONRPCResponse
		json.Unmarshal([]byte(lines[0]), &response)

		if response.Error != nil {
			t.Fatalf("List tools failed: %v", response.Error)
		}

		var result map[string]interface{}
		json.Unmarshal(response.Result, &result)

		toolsList := result["tools"].([]interface{})
		if len(toolsList) != 4 {
			t.Errorf("Expected 4 tools, got %d", len(toolsList))
		}

		// Verify tool names
		expectedTools := map[string]bool{
			"resize_image":   false,
			"scale_image":    false,
			"compress_image": false,
			"list_models":    false,
		}

		for _, toolInterface := range toolsList {
			tool := toolInterface.(map[string]interface{})
			name := tool["name"].(string)
			if _, exists := expectedTools[name]; exists {
				expectedTools[name] = true
			}
		}

		for name, found := range expectedTools {
			if !found {
				t.Errorf("Expected tool %s not found in list", name)
			}
		}
	})

	// Test 3: Call Tool - Resize
	t.Run("call_resize_tool", func(t *testing.T) {
		params := map[string]interface{}{
			"name": "resize_image",
			"arguments": map[string]interface{}{
				"input":  testImagePath,
				"width":  100,
				"height": 100,
			},
		}
		paramsBytes, _ := json.Marshal(params)

		callRequest := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"call-1"`),
			Method:  mcp.MethodCallTool,
			Params:  json.RawMessage(paramsBytes),
		}

		requestBytes, _ := json.Marshal(callRequest)
		stdin := bytes.NewBuffer(requestBytes)
		stdin.WriteString("\n")
		stdout := &bytes.Buffer{}

		server.stdin = stdin
		server.stdout = stdout

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- server.Start(ctx)
		}()

		cancel()
		<-done

		output := stdout.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		var response mcp.JSONRPCResponse
		json.Unmarshal([]byte(lines[0]), &response)

		if response.Error != nil {
			t.Fatalf("Resize tool call failed: %v", response.Error)
		}

		var result map[string]interface{}
		json.Unmarshal(response.Result, &result)

		content := result["content"].([]interface{})
		if len(content) == 0 {
			t.Fatal("Expected content in response")
		}

		contentItem := content[0].(map[string]interface{})
		text := contentItem["text"].(string)

		// Verify response contains success info
		if !strings.Contains(text, "success") {
			t.Error("Expected success in response text")
		}
	})

	// Test 4: Call Tool - List Models
	t.Run("call_list_models_tool", func(t *testing.T) {
		params := map[string]interface{}{
			"name":      "list_models",
			"arguments": map[string]interface{}{},
		}
		paramsBytes, _ := json.Marshal(params)

		callRequest := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"call-2"`),
			Method:  mcp.MethodCallTool,
			Params:  json.RawMessage(paramsBytes),
		}

		requestBytes, _ := json.Marshal(callRequest)
		stdin := bytes.NewBuffer(requestBytes)
		stdin.WriteString("\n")
		stdout := &bytes.Buffer{}

		server.stdin = stdin
		server.stdout = stdout

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- server.Start(ctx)
		}()

		cancel()
		<-done

		output := stdout.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		var response mcp.JSONRPCResponse
		json.Unmarshal([]byte(lines[0]), &response)

		if response.Error != nil {
			t.Fatalf("List models tool call failed: %v", response.Error)
		}

		var result map[string]interface{}
		json.Unmarshal(response.Result, &result)

		content := result["content"].([]interface{})
		contentItem := content[0].(map[string]interface{})
		text := contentItem["text"].(string)

		// Verify response contains model info
		if !strings.Contains(text, "models") {
			t.Error("Expected models in response text")
		}
	})

	// Test 5: Multiple Requests in Sequence
	t.Run("multiple_sequential_requests", func(t *testing.T) {
		// Prepare multiple requests
		requests := []string{
			`{"jsonrpc":"2.0","id":"seq-1","method":"initialize","params":{}}`,
			`{"jsonrpc":"2.0","id":"seq-2","method":"tools/list","params":{}}`,
		}

		stdin := bytes.NewBufferString(strings.Join(requests, "\n") + "\n")
		stdout := &bytes.Buffer{}

		server.stdin = stdin
		server.stdout = stdout

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- server.Start(ctx)
		}()

		cancel()
		<-done

		output := stdout.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Should have 2 responses
		if len(lines) < 2 {
			t.Errorf("Expected at least 2 responses, got %d", len(lines))
		}

		// Verify both are valid JSON-RPC responses
		for i, line := range lines {
			if line == "" {
				continue
			}
			var response mcp.JSONRPCResponse
			if err := json.Unmarshal([]byte(line), &response); err != nil {
				t.Errorf("Response %d is not valid JSON-RPC: %v", i, err)
			}
		}
	})
}

// TestMCPServerErrorHandling tests error scenarios
func TestMCPServerErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	tools.RegisterResizeImageTool(server)

	// Test 1: Invalid tool name
	t.Run("invalid_tool_name", func(t *testing.T) {
		params := map[string]interface{}{
			"name":      "nonexistent_tool",
			"arguments": map[string]interface{}{},
		}
		paramsBytes, _ := json.Marshal(params)

		request := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"err-1"`),
			Method:  mcp.MethodCallTool,
			Params:  json.RawMessage(paramsBytes),
		}

		response := server.handleRequest(context.Background(), &request)

		if response.Error == nil {
			t.Error("Expected error for nonexistent tool")
		}

		if response.Error.Code != mcp.ErrorCodeMethodNotFound {
			t.Errorf("Expected error code %d, got %d", mcp.ErrorCodeMethodNotFound, response.Error.Code)
		}
	})

	// Test 2: Invalid parameters
	t.Run("invalid_parameters", func(t *testing.T) {
		params := map[string]interface{}{
			"name": "resize_image",
			"arguments": map[string]interface{}{
				"input": "test.png",
				// Missing width and height
			},
		}
		paramsBytes, _ := json.Marshal(params)

		request := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"err-2"`),
			Method:  mcp.MethodCallTool,
			Params:  json.RawMessage(paramsBytes),
		}

		response := server.handleRequest(context.Background(), &request)

		if response.Error == nil {
			t.Error("Expected error for invalid parameters")
		}
	})

	// Test 3: Invalid method
	t.Run("invalid_method", func(t *testing.T) {
		request := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"err-3"`),
			Method:  "invalid.method",
			Params:  json.RawMessage(`{}`),
		}

		response := server.handleRequest(context.Background(), &request)

		if response.Error == nil {
			t.Error("Expected error for invalid method")
		}

		if response.Error.Code != mcp.ErrorCodeMethodNotFound {
			t.Errorf("Expected error code %d, got %d", mcp.ErrorCodeMethodNotFound, response.Error.Code)
		}
	})
}

// TestMCPServerConcurrency tests concurrent request handling
func TestMCPServerConcurrency(t *testing.T) {
	cfg := &config.Config{}
	server := mcp.NewMCPServer("test", "1.0.0", cfg, false)
	tools.RegisterListModelsTool(server)

	// Send multiple requests concurrently via multiple lines
	numRequests := 10
	var requests []string
	for i := 0; i < numRequests; i++ {
		req := mcp.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`"` + string(rune('0'+i)) + `"`),
			Method:  mcp.MethodListTools,
			Params:  json.RawMessage(`{}`),
		}
		reqBytes, _ := json.Marshal(req)
		requests = append(requests, string(reqBytes))
	}

	stdin := bytes.NewBufferString(strings.Join(requests, "\n") + "\n")
	stdout := &bytes.Buffer{}

	server.stdin = stdin
	server.stdout = stdout

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Start(ctx)
	}()

	cancel()
	<-done

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should receive responses for all requests
	if len(lines) < numRequests {
		t.Errorf("Expected at least %d responses, got %d", numRequests, len(lines))
	}

	// Verify all responses are valid
	for _, line := range lines {
		if line == "" {
			continue
		}
		var response mcp.JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Errorf("Invalid response: %v", err)
		}

		if response.Error != nil {
			t.Errorf("Request failed: %v", response.Error)
		}
	}
}
