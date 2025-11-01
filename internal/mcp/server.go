package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/chadneal/gimage/internal/config"
)

// MCPServer implements the Model Context Protocol server
type MCPServer struct {
	name    string
	version string
	config  *config.Config
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	tools   map[string]Tool
	verbose bool
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(name, version string, cfg *config.Config, verbose bool) *MCPServer {
	return &MCPServer{
		name:    name,
		version: version,
		config:  cfg,
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
		tools:   make(map[string]Tool),
		verbose: verbose,
	}
}

// RegisterTool adds a tool to the server
func (s *MCPServer) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
	if s.verbose {
		s.logInfo("Registered tool: %s", tool.Name)
	}
}

// GetTool returns a tool by name, or nil if not found
func (s *MCPServer) GetTool(name string) *Tool {
	if tool, exists := s.tools[name]; exists {
		return &tool
	}
	return nil
}

// Start begins listening for MCP protocol messages
func (s *MCPServer) Start(ctx context.Context) error {
	if s.verbose {
		s.logInfo("MCP server starting")
		s.logInfo("Protocol version: %s", ProtocolVersion)
		s.logInfo("Registered tools: %d", len(s.tools))
	}

	scanner := bufio.NewScanner(s.stdin)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()

		if s.verbose {
			s.logInfo("Received request: %s", string(line))
		}

		var request JSONRPCRequest
		if err := json.Unmarshal(line, &request); err != nil {
			s.logError("Failed to parse request: %v", err)
			continue
		}

		// CRITICAL: Detect notifications vs requests
		// Notifications have NO id field and must NOT receive responses
		if request.ID == nil {
			if s.verbose {
				s.logInfo("Received notification: %s (no response will be sent)", request.Method)
			}
			// Handle notification but DO NOT send response
			s.handleNotification(ctx, &request)
			continue
		}

		// This is a request (has ID), send a response
		response := s.handleRequest(ctx, &request)

		responseBytes, err := json.Marshal(response)
		if err != nil {
			s.logError("Failed to marshal response: %v", err)
			continue
		}

		if s.verbose {
			s.logInfo("Sending response: %s", string(responseBytes))
		}

		fmt.Fprintln(s.stdout, string(responseBytes))
	}

	return scanner.Err()
}

// handleNotification processes MCP notifications (messages with no ID that expect no response)
func (s *MCPServer) handleNotification(ctx context.Context, req *JSONRPCRequest) {
	// According to MCP spec, notifications are fire-and-forget
	// We log them but take no action
	if s.verbose {
		s.logInfo("Notification received: %s", req.Method)
	}
	// No response is sent for notifications
}

func (s *MCPServer) logInfo(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, "[gimage-mcp] "+format+"\n", args...)
}

func (s *MCPServer) logError(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, "[gimage-mcp] ERROR: "+format+"\n", args...)
}
