# gimage Architecture Design

## Model Registry System

The gimage tool abstracts multiple cloud providers (Google Gemini, Vertex AI, AWS Bedrock) and their various client implementations (SDK vs REST). To manage this complexity, we use a **Model Registry** pattern.

## Architecture Overview

```
User Input (model name/alias)
            ↓
    Model Registry
            ↓
    Resolve Provider
            ↓
    Check Credentials
            ↓
    Create Client (REST/SDK)
            ↓
    Generate Image
```

## Key Components

### 1. Model Registry (`registry.go`)

The central registry that manages all model providers and their configurations:

```go
type ModelProvider struct {
    ModelID             string              // Exact model ID
    DisplayName         string              // Human-friendly name
    Aliases             []string            // Alternative names
    API                 string              // API provider (gemini/vertex/bedrock)
    ClientType          ClientType          // REST or SDK
    APIModelName        map[string]string   // API-specific model names
    RequiredCredentials []string            // Required env vars/config
    ClientFactory       ClientFactory      // Creates the appropriate client
}
```

### 2. Client Selection Logic

The registry automatically selects the appropriate client based on:

1. **Model Requirements**: Some models only work with specific clients
2. **Bug Workarounds**: e.g., Gemini SDK has image generation bugs, so we use REST
3. **Credential Availability**: Uses API key for Express mode, SDK for service accounts
4. **Performance**: REST clients for simple auth, SDK for complex IAM

### 3. Example Flow

When a user requests "gemini flash 2.5":
1. Registry resolves alias "gemini flash 2.5" → "gemini-2.5-flash-image"
2. Finds the ModelProvider configuration
3. Checks if GEMINI_API_KEY is available
4. Creates GeminiRESTClient (not SDK due to bugs)
5. Translates model name to API format if needed
6. Executes generation with proper client

## Client Implementations

### REST Clients
- **GeminiRESTClient**: Direct HTTP calls to Gemini API
- **VertexRESTClient**: Vertex AI Express mode with API key
- **BedrockRESTClient**: AWS Bedrock with bearer token

### SDK Clients
- **GeminiClient**: google.golang.org/genai SDK (has bugs, not used)
- **VertexSDKClient**: Vertex AI with service account auth
- **BedrockSDKClient**: AWS SDK with IAM credentials

## Benefits

1. **Single Entry Point**: `GenerateWithRegistry()` handles all complexity
2. **Automatic Client Selection**: Picks the right client based on credentials and bugs
3. **Credential Validation**: Checks all required credentials before attempting
4. **Model Aliasing**: Users can use friendly names like "gemini" or "imagen"
5. **Easy Extension**: Add new models by registering a ModelProvider

## Adding a New Model

To add a new model, register it in `registerAllProviders()`:

```go
r.Register(&ModelProvider{
    ModelID:     "new-model-id",
    DisplayName: "New Model",
    Aliases:     []string{"new", "model"},
    API:         "provider",
    ClientType:  ClientTypeREST,
    RequiredCredentials: []string{"PROVIDER_API_KEY"},
    ClientFactory: func(cfg *config.Config) (ImageGenerator, error) {
        return NewProviderClient(cfg.ProviderAPIKey)
    },
})
```

## Current Workarounds

1. **Gemini SDK Bug**: The google.golang.org/genai SDK always routes image generation to Vertex AI endpoints. We use REST client instead.

2. **Vertex Express Mode**: When Vertex API key is available, use REST client. Otherwise fall back to SDK with service account.

3. **Bedrock Auth**: Try bearer token first (simpler), fall back to IAM/SDK.