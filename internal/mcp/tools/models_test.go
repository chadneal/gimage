package tools

import (
	"testing"

	"github.com/apresai/gimage/internal/mcp"
)

func TestRegisterListModelsTool(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)

	// Register the tool
	RegisterListModelsTool(server)

	// Verify tool is registered
	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	// Verify tool properties
	if tool.Name != "list_models" {
		t.Errorf("Expected name 'list_models', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Tool description is empty")
	}

	if tool.Handler == nil {
		t.Error("Tool handler is nil")
	}
}

func TestListModelsTool_InputSchema(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	schema := tool.InputSchema

	// list_models takes no arguments
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties field is not a map")
	}

	// Should have empty properties (no arguments)
	if len(properties) != 0 {
		t.Errorf("Expected 0 properties (no arguments), got %d", len(properties))
	}
}

func TestListModelsTool_ResponseStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	// Call the handler with empty args
	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check for required top-level fields
	requiredFields := []string{"models", "total", "credentials", "default_model", "pricing_note", "recommendations"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Required field '%s' missing from result", field)
		}
	}

	// Verify models is an array
	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	// Should have at least some models
	if len(models) == 0 {
		t.Error("No models returned (expected at least some models)")
	}

	// Verify total matches length
	total, ok := result["total"].(int)
	if !ok {
		t.Fatal("total field is not an integer")
	}

	if total != len(models) {
		t.Errorf("total (%d) does not match models length (%d)", total, len(models))
	}
}

func TestListModelsTool_ModelStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	if len(models) == 0 {
		t.Fatal("No models to test")
	}

	// Check first model structure
	model := models[0]
	requiredModelFields := []string{
		"name", "display_name", "api", "quality", "description",
		"priority", "available", "requires_auth", "max_resolution",
		"supported_sizes", "pricing", "pricing_summary",
		"supports_styles", "supports_negative_prompt", "supports_seed",
		"supported_styles", "max_prompt_length",
	}

	for _, field := range requiredModelFields {
		if _, exists := model[field]; !exists {
			t.Errorf("Required model field '%s' missing", field)
		}
	}
}

func TestListModelsTool_PricingStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	if len(models) == 0 {
		t.Fatal("No models to test")
	}

	// Check first model's pricing structure
	model := models[0]
	pricing, ok := model["pricing"].(map[string]interface{})
	if !ok {
		t.Fatal("pricing field is not a map")
	}

	// Required pricing fields
	requiredPricingFields := []string{"billing_unit", "currency", "pricing_tier", "free_tier"}
	for _, field := range requiredPricingFields {
		if _, exists := pricing[field]; !exists {
			t.Errorf("Required pricing field '%s' missing", field)
		}
	}
}

func TestListModelsTool_CredentialsStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	credentials, ok := result["credentials"].(map[string]interface{})
	if !ok {
		t.Fatal("credentials field is not a map")
	}

	// Check for credential status fields
	requiredCredFields := []string{"gemini_configured", "vertex_configured"}
	for _, field := range requiredCredFields {
		if _, exists := credentials[field]; !exists {
			t.Errorf("Required credentials field '%s' missing", field)
		}
	}

	// Verify they are booleans
	if _, ok := credentials["gemini_configured"].(bool); !ok {
		t.Error("gemini_configured is not a boolean")
	}

	if _, ok := credentials["vertex_configured"].(bool); !ok {
		t.Error("vertex_configured is not a boolean")
	}
}

func TestListModelsTool_DefaultModelStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	defaultModel, ok := result["default_model"].(map[string]interface{})
	if !ok {
		t.Fatal("default_model field is not a map")
	}

	// Check for default model fields
	requiredDefaultFields := []string{"name", "display_name", "pricing_summary"}
	for _, field := range requiredDefaultFields {
		if _, exists := defaultModel[field]; !exists {
			t.Errorf("Required default_model field '%s' missing", field)
		}
	}
}

func TestListModelsTool_RecommendationsStructure(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	recommendations, ok := result["recommendations"].(map[string]interface{})
	if !ok {
		t.Fatal("recommendations field is not a map")
	}

	// Check for recommendation categories
	requiredRecommendations := []string{"free_users", "paid_users", "max_quality"}
	for _, field := range requiredRecommendations {
		if _, exists := recommendations[field]; !exists {
			t.Errorf("Required recommendation '%s' missing", field)
		}
	}

	// Verify they are strings
	for _, field := range requiredRecommendations {
		if _, ok := recommendations[field].(string); !ok {
			t.Errorf("Recommendation '%s' is not a string", field)
		}
	}
}

func TestListModelsTool_ModelAvailability(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	// Check that each model has an availability status
	for i, model := range models {
		available, exists := model["available"]
		if !exists {
			t.Errorf("Model %d missing 'available' field", i)
			continue
		}

		// Should be a boolean
		if _, ok := available.(bool); !ok {
			t.Errorf("Model %d 'available' field is not a boolean", i)
		}

		// Check requires_auth field
		requiresAuth, exists := model["requires_auth"]
		if !exists {
			t.Errorf("Model %d missing 'requires_auth' field", i)
			continue
		}

		// requires_auth should be a string array
		if _, ok := requiresAuth.([]string); !ok {
			t.Errorf("Model %d 'requires_auth' field is not a string array", i)
		}
	}
}

func TestListModelsTool_ModelPriority(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	// Verify models are sorted by priority
	if len(models) < 2 {
		t.Skip("Need at least 2 models to test priority sorting")
	}

	for i := 1; i < len(models); i++ {
		prevPriority, ok1 := models[i-1]["priority"].(int)
		currPriority, ok2 := models[i]["priority"].(int)

		if !ok1 || !ok2 {
			t.Fatal("priority field is not an integer")
		}

		if prevPriority > currPriority {
			t.Errorf("Models not sorted by priority: %d > %d at position %d", prevPriority, currPriority, i)
		}
	}
}

func TestListModelsTool_PricingNote(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	pricingNote, ok := result["pricing_note"].(string)
	if !ok {
		t.Fatal("pricing_note is not a string")
	}

	if pricingNote == "" {
		t.Error("pricing_note is empty")
	}

	// Should mention USD, free tier, and batch mode
	if len(pricingNote) < 50 {
		t.Error("pricing_note seems too short to be informative")
	}
}

func TestListModelsTool_APITypes(t *testing.T) {
	server := mcp.NewMCPServer("test", "1.0.0", nil, false)
	RegisterListModelsTool(server)

	tool := server.GetTool("list_models")
	if tool == nil {
		t.Fatal("list_models tool not registered")
	}

	result, err := tool.Handler(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	models, ok := result["models"].([]map[string]interface{})
	if !ok {
		t.Fatal("models field is not an array of maps")
	}

	// Track which APIs are present
	apiTypes := make(map[string]bool)
	for _, model := range models {
		api, ok := model["api"].(string)
		if !ok {
			t.Error("Model api field is not a string")
			continue
		}
		apiTypes[api] = true
	}

	// Should have at least one API type
	if len(apiTypes) == 0 {
		t.Error("No API types found in models")
	}

	// Valid API types
	validAPIs := map[string]bool{"gemini": true, "vertex": true, "bedrock": true}
	for api := range apiTypes {
		if !validAPIs[api] {
			t.Errorf("Invalid API type: %s", api)
		}
	}
}
