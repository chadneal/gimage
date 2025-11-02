# AWS Nova Canvas Integration Plan

**Date**: 2025-11-02
**Goal**: Add AWS Bedrock Nova Canvas as a new image generation backend to gimage
**Estimated Time**: 10-15 hours

---

## Executive Summary

AWS Nova Canvas is Amazon's latest image generation model available through AWS Bedrock. This plan outlines how to integrate it into gimage's multi-backend architecture while maintaining consistency with existing Gemini and Vertex AI integrations.

**Why Nova Canvas?**
- High-quality image generation ($0.04/image standard, $0.08 premium)
- Up to 2048x2048 resolution
- Fast performance (10 req/sec)
- Good for users already on AWS infrastructure
- Supports styles, negative prompts, and seeds

---

## Phase-by-Phase Action Plan

### Phase 1: Research (2-3 hours)

#### Tasks
1. **Read AWS Bedrock Documentation**
   - [ ] Visit: https://docs.aws.amazon.com/bedrock/
   - [ ] Find Nova Canvas API reference
   - [ ] Document request/response format
   - [ ] Note authentication requirements (AWS Signature V4)
   - [ ] Check available regions
   - [ ] Document IAM permissions needed

2. **Test API Manually**
   - [ ] Set up AWS account if needed
   - [ ] Request Bedrock model access in console
   - [ ] Create IAM user with `bedrock:InvokeModel` permission
   - [ ] Test API call with `curl` or AWS CLI
   - [ ] Verify request/response format
   - [ ] Test error scenarios (invalid prompt, rate limit)

3. **Document Findings**
   - [ ] Create `API_NOTES.md` with all research findings
   - [ ] List supported image sizes
   - [ ] Document pricing tiers
   - [ ] Note rate limits and quotas
   - [ ] Document error codes

#### Key Questions to Answer
- What is the exact API endpoint format?
- How does AWS Signature V4 auth work? (SDK will handle this)
- What is the request body schema for Nova Canvas?
- What is the response format? (JSON with base64 images?)
- What regions support Nova Canvas?
- What IAM permissions are required?

#### Expected Outputs
- Clear understanding of AWS Bedrock API
- API_NOTES.md with request/response examples
- Test AWS account with working credentials

---

### Phase 2: Design (1 hour)

#### Architecture Decisions

**Decision 1: SDK vs REST**
- ✅ **Use AWS SDK v2 for Go** (recommended)
- Rationale: AWS Signature V4 is complex; SDK handles auth elegantly
- SDK is well-maintained and idiomatic Go
- Trade-off: Adds ~10MB to binary, but worth it for security and maintainability

**Decision 2: Client File Location**
- `internal/generate/bedrock_sdk.go`
- Follow pattern of `gemini_rest.go` and `vertex_sdk.go`

**Decision 3: Authentication Modes**
- Support two modes:
  1. **Access Keys**: `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY`
  2. **AWS Profile**: Use `~/.aws/credentials` via `AWS_PROFILE`
- Priority: env vars > config file > AWS default profile

**Decision 4: Model Naming**
- Official: `amazon.nova-canvas-v1:0`
- Aliases: `nova-canvas`, `nova`, `bedrock-canvas`
- Display name: "AWS Nova Canvas"

#### Configuration Schema

Add to `~/.gimage/config.md`:
```markdown
**aws_access_key_id**: AKIAIOSFODNN7EXAMPLE
**aws_secret_access_key**: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
**aws_region**: us-east-1
**aws_profile**: default
```

Add to `internal/config/config.go`:
```go
type Config struct {
    // ... existing fields ...
    AWSAccessKeyID     string
    AWSSecretAccessKey string
    AWSRegion          string
    AWSProfile         string
}
```

---

### Phase 3: Implementation (4-6 hours)

#### Subtask 3.1: AWS SDK Client (2-3 hours)

**File**: `internal/generate/bedrock_sdk.go`

Implementation checklist:
- [ ] Create `BedrockSDKClient` struct
- [ ] Implement `NewBedrockSDKClient(ctx, region)` constructor
- [ ] Implement `GenerateImage(ctx, prompt, options)` method
- [ ] Implement `generateWithRetry()` helper
- [ ] Add circuit breaker integration
- [ ] Add retry logic with exponential backoff
- [ ] Implement `handleSDKError()` for AWS error mapping
- [ ] Implement `Close()` method
- [ ] Add verbose logging with `logVerbose()`

**Key Implementation Details**:
```go
// Request format (check AWS docs for exact schema)
requestBody := map[string]interface{}{
    "textToImageParams": map[string]interface{}{
        "text": prompt,
        "negativeText": options.NegativePrompt,  // if provided
    },
    "taskType": "TEXT_IMAGE",
    "imageGenerationConfig": map[string]interface{}{
        "numberOfImages": 1,
        "quality":        "standard",  // or "premium"
        "height":         1024,
        "width":          1024,
        "cfgScale":       7.0,
        "seed":           options.Seed,  // if provided
    },
}

// Use AWS SDK
input := &bedrockruntime.InvokeModelInput{
    ModelId:     aws.String("amazon.nova-canvas-v1:0"),
    ContentType: aws.String("application/json"),
    Accept:      aws.String("application/json"),
    Body:        requestJSON,
}

result, err := c.client.InvokeModel(ctx, input)
```

#### Subtask 3.2: Model Registry Updates (30 min)

**File**: `internal/generate/models.go`

- [ ] Add model constant: `ModelNovaCanvasV1 = "amazon.nova-canvas-v1:0"`
- [ ] Add model aliases to `ModelAliases` map
- [ ] Add full `ModelInfo` entry to `AvailableModels()`
- [ ] Include pricing, capabilities, rate limits
- [ ] Set priority (suggest 7, after Imagen 4 models)
- [ ] Update `ValidateConfig()` to accept "bedrock" API

#### Subtask 3.3: Configuration Updates (30 min)

**File**: `internal/config/config.go`

- [ ] Add AWS fields to `Config` struct
- [ ] Update `parseMarkdownConfig()` to parse AWS keys
- [ ] Update `SaveConfig()` to write AWS keys
- [ ] Create `GetAWSRegion()` helper function
- [ ] Create `HasBedrockCredentials()` helper function
- [ ] Update environment variable loading

#### Subtask 3.4: CLI Integration (1 hour)

**File**: `internal/cli/generate.go`

- [ ] Add `bedrock` case to API selection logic
- [ ] Add Bedrock auto-detection when only AWS credentials exist
- [ ] Update multi-credential auto-detection
- [ ] Add model info display for Bedrock models
- [ ] Handle Bedrock-specific flags (region)

#### Subtask 3.5: Authentication Command (1 hour)

**File**: `internal/cli/auth.go`

- [ ] Create `authBedrockCmd` command
- [ ] Implement `setupBedrockAuth()` function
- [ ] Create `setupBedrockAccessKeys()` (mode 1)
- [ ] Create `setupBedrockProfile()` (mode 2)
- [ ] Add interactive prompts with defaults
- [ ] Add success messages with usage examples
- [ ] Register command: `authCmd.AddCommand(authBedrockCmd)`

---

### Phase 4: Testing (2-3 hours)

#### Subtask 4.1: Unit Tests (1 hour)

**File**: `internal/generate/bedrock_sdk_test.go`

- [ ] Create mock AWS client
- [ ] Test successful generation
- [ ] Test empty prompt error
- [ ] Test API error responses (ThrottlingException, AccessDeniedException)
- [ ] Test error handling function
- [ ] Test request payload building
- [ ] Test response parsing

**Important**: Unit tests should NOT make real API calls (they cost money!)

#### Subtask 4.2: Integration Tests (30 min)

**File**: `test/integration/bedrock_test.go`

- [ ] Create integration test with `// +build integration` tag
- [ ] Test with real AWS credentials (manual run only)
- [ ] Test successful generation end-to-end
- [ ] Test error scenarios
- [ ] Skip if credentials not available

Run with: `go test -tags=integration ./test/integration/...`

#### Subtask 4.3: Manual Testing (1 hour)

Manual test checklist:
- [ ] `gimage auth bedrock` - interactive setup
- [ ] `gimage generate --list-models` - verify Nova Canvas appears
- [ ] `gimage generate --api bedrock "test"` - basic generation
- [ ] `gimage generate --model amazon.nova-canvas-v1:0 "test"` - explicit model
- [ ] `gimage generate "test" --size 1024x1792` - different size
- [ ] `gimage generate "test" --negative "blur"` - negative prompt
- [ ] `gimage generate "test" --seed 42` - reproducible generation
- [ ] `gimage generate "test" --verbose` - check verbose logs
- [ ] Test with only Bedrock credentials (auto-detection)
- [ ] Test with AWS_PROFILE instead of access keys
- [ ] Test error: missing credentials
- [ ] Test error: invalid model ID
- [ ] Test error: invalid region

#### Subtask 4.4: MCP Server Testing (30 min)

- [ ] Start MCP server: `gimage serve`
- [ ] Verify `list_models` includes Nova Canvas
- [ ] Test `generate_image` with Bedrock model
- [ ] Test parameter validation
- [ ] Test error responses

---

### Phase 5: Documentation (1-2 hours)

#### Subtask 5.1: Update Core Documentation (30 min)

**Files to Update**:
- [ ] `docs/CLAUDE.md` - Add AWS Bedrock Backend section
- [ ] `docs/MCP_TOOLS.md` - Add Nova Canvas to supported models
- [ ] `README.md` - Add AWS Bedrock to providers table
- [ ] `CHANGELOG.md` - Document new feature (version 0.3.0)

#### Subtask 5.2: Create Migration Guide (30 min)

**File**: `docs/BEDROCK_MIGRATION.md`

- [ ] Prerequisites (AWS account, Bedrock access)
- [ ] Step-by-step setup instructions
- [ ] IAM permissions documentation
- [ ] Troubleshooting common errors
- [ ] Example commands
- [ ] Cost estimation examples

#### Subtask 5.3: Update Usage Examples (30 min)

Add Bedrock examples to:
- [ ] `docs/MCP_EXAMPLES.md` - MCP usage examples
- [ ] `docs/API.md` - CLI usage examples
- [ ] README.md - Quick start examples

---

### Phase 6: Integration & Release (1 hour)

#### Subtask 6.1: Dependencies (15 min)

```bash
# Add AWS SDK dependencies
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime@latest
go get github.com/aws/aws-sdk-go-v2/credentials@latest
go mod tidy
```

- [ ] Run `go get` commands
- [ ] Run `go mod tidy`
- [ ] Verify `go.mod` is clean
- [ ] Check for version conflicts

#### Subtask 6.2: Build & Validate (15 min)

```bash
# Run full test suite
make test

# Run linter
make lint

# Build binary
make build

# Check binary size (should be ~15-20MB)
ls -lh bin/gimage
```

- [ ] All tests pass
- [ ] No lint errors
- [ ] Binary builds successfully
- [ ] Binary size is reasonable

#### Subtask 6.3: Version Bump (15 min)

- [ ] Update version in `cmd/gimage/main.go`: `0.3.0`
- [ ] Update `CHANGELOG.md` with date (run `date +%Y-%m-%d`)
- [ ] Add comprehensive changelog entry
- [ ] Update README.md version references

#### Subtask 6.4: Final Testing (15 min)

End-to-end smoke test:
```bash
# Clean build
make clean && make build

# Test auth
./bin/gimage auth bedrock

# Test generation
./bin/gimage generate --api bedrock "a mountain landscape" -o test.png

# Verify image
file test.png  # Should be PNG

# Test MCP
./bin/gimage serve  # Verify no crashes
```

---

## File Changes Summary

| File | Status | Description |
|------|--------|-------------|
| `internal/generate/bedrock_sdk.go` | ➕ Create | AWS Bedrock SDK client implementation |
| `internal/generate/bedrock_sdk_test.go` | ➕ Create | Unit tests for Bedrock client |
| `internal/generate/models.go` | ✏️ Modify | Add Nova Canvas model metadata |
| `internal/config/config.go` | ✏️ Modify | Add AWS config fields and helpers |
| `internal/cli/auth.go` | ✏️ Modify | Add `gimage auth bedrock` command |
| `internal/cli/generate.go` | ✏️ Modify | Add Bedrock API case in generation |
| `test/integration/bedrock_test.go` | ➕ Create | Integration tests (manual) |
| `docs/MODEL_ONBOARDING.md` | ✅ Done | General model onboarding guide |
| `docs/AWS_NOVA_CANVAS_PLAN.md` | ✅ Done | This document |
| `docs/BEDROCK_MIGRATION.md` | ➕ Create | User-facing setup guide |
| `docs/CLAUDE.md` | ✏️ Modify | Add Bedrock backend section |
| `docs/MCP_TOOLS.md` | ✏️ Modify | Add Nova Canvas to models list |
| `docs/MCP_EXAMPLES.md` | ✏️ Modify | Add Bedrock usage examples |
| `README.md` | ✏️ Modify | Add AWS Bedrock to providers |
| `CHANGELOG.md` | ✏️ Modify | Document v0.3.0 changes |
| `go.mod` | ✏️ Modify | Add AWS SDK v2 dependencies |
| `cmd/gimage/main.go` | ✏️ Modify | Bump version to 0.3.0 |

**Summary**: 6 new files, 11 modified files

---

## Dependencies to Add

```go
// go.mod additions
require (
    github.com/aws/aws-sdk-go-v2/config v1.18.0
    github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.2.0
    github.com/aws/aws-sdk-go-v2/credentials v1.13.0
)
```

**Binary Size Impact**: +8-10MB (AWS SDK)

---

## Testing Strategy

### Unit Tests (Fast, No Credentials)
```bash
go test ./internal/generate/bedrock_sdk_test.go -v
```
- Mock AWS API responses
- Test error handling
- Test request payload building
- Test response parsing
- **Cost**: $0 (mocked)

### Integration Tests (Slow, Requires Credentials)
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
go test -tags=integration ./test/integration/bedrock_test.go -v
```
- Real API calls
- End-to-end flow
- **Cost**: ~$0.04 per test run

### Manual Tests (Complete Validation)
```bash
# Auth setup
gimage auth bedrock

# Basic generation
gimage generate --api bedrock "test image"

# All features
gimage generate --model amazon.nova-canvas-v1:0 \
  --size 1024x1792 \
  --negative "blur, artifacts" \
  --seed 42 \
  --verbose \
  -o test_nova.png
```
- Full CLI workflow
- User experience validation
- **Cost**: ~$0.12-0.20 (3-5 test images)

---

## Risk Assessment

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| AWS SDK version conflicts | Low | Medium | Pin SDK versions, test thoroughly |
| Binary size increase | High | Low | Acceptable trade-off for functionality |
| Auth complexity (IAM) | Medium | High | Comprehensive error messages, docs |
| Rate limiting different than expected | Medium | Medium | Implement conservative circuit breaker |
| Regional availability varies | Medium | Medium | Document region support clearly |

### User Experience Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| AWS setup too complex | Medium | High | Create detailed migration guide |
| Confusion about IAM permissions | High | High | Document exact permissions needed |
| Cost surprises (no free tier) | Medium | Medium | Show cost warnings in CLI |
| Existing users affected | Low | Critical | Ensure backward compatibility |

### Mitigation Strategies
1. **Clear Documentation**: Step-by-step setup guide with screenshots
2. **Error Messages**: Include IAM permission errors with fix suggestions
3. **Cost Warnings**: Display estimated cost before generation
4. **Backward Compatibility**: Don't change existing Gemini/Vertex behavior
5. **Testing**: Extensive manual testing before release

---

## Success Criteria

### Functional Requirements
- [x] Nova Canvas model available in `--list-models`
- [x] `gimage generate --api bedrock` works end-to-end
- [x] `gimage auth bedrock` sets up credentials correctly
- [x] Auto-detection works when only AWS credentials exist
- [x] All image sizes supported (512x512 to 2048x2048)
- [x] Negative prompts work
- [x] Seed-based reproducibility works
- [x] MCP server exposes Bedrock models

### Non-Functional Requirements
- [x] Unit test coverage >80%
- [x] All lint checks pass
- [x] Error messages are clear and actionable
- [x] Documentation complete and accurate
- [x] Binary size increase <15MB
- [x] No breaking changes to existing functionality

### User Experience Requirements
- [x] Setup takes <5 minutes for AWS users
- [x] Error messages include AWS-specific troubleshooting
- [x] Verbose mode shows detailed AWS SDK logs
- [x] Cost warnings shown for paid models
- [x] Migration guide covers all setup scenarios

---

## Timeline Estimate

| Phase | Time | Cumulative |
|-------|------|------------|
| Phase 1: Research | 2-3h | 3h |
| Phase 2: Design | 1h | 4h |
| Phase 3: Implementation | 4-6h | 10h |
| Phase 4: Testing | 2-3h | 13h |
| Phase 5: Documentation | 1-2h | 15h |
| Phase 6: Integration | 1h | 16h |

**Total**: 10-16 hours (depends on AWS API familiarity)

**Recommended Schedule**:
- Day 1: Research + Design (4 hours)
- Day 2: Implementation (6 hours)
- Day 3: Testing + Documentation + Integration (6 hours)

---

## Next Steps

### Immediate Actions (Today)
1. ✅ Review this plan document
2. ✅ Review MODEL_ONBOARDING.md
3. [ ] Set up AWS account with Bedrock access
4. [ ] Request Nova Canvas model access in AWS Console
5. [ ] Create IAM user with necessary permissions
6. [ ] Start Phase 1: Research

### Week 1
- Complete Phases 1-3 (Research, Design, Implementation)
- Have working prototype generating images

### Week 2
- Complete Phases 4-6 (Testing, Documentation, Integration)
- Ready for feature branch PR

### Before Release
- [ ] Full test suite passes
- [ ] Documentation complete
- [ ] Manual testing complete
- [ ] Code review completed
- [ ] CHANGELOG.md updated
- [ ] Version bumped

---

## Questions to Resolve

### Before Starting Implementation
1. What is the exact AWS Bedrock API endpoint for Nova Canvas?
2. What is the exact request body schema?
3. What IAM permissions are required beyond `bedrock:InvokeModel`?
4. Which AWS regions support Nova Canvas?
5. What are the exact rate limits? (docs vs reality)

### During Implementation
1. Does AWS SDK handle retries automatically, or do we need our own?
2. How do we map AWS error types to user-friendly messages?
3. Should we support AWS session tokens for temporary credentials?
4. Do we need to handle AWS credential rotation?

### Before Release
1. Should we default to Bedrock if user has AWS credentials?
2. What priority should Nova Canvas have in model selection?
3. Should we add CloudWatch metrics integration?
4. Do we need to document AWS billing/cost tracking?

---

## Reference Links

### AWS Documentation
- [AWS Bedrock Developer Guide](https://docs.aws.amazon.com/bedrock/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
- [Bedrock Runtime API Reference](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_InvokeModel.html)
- [IAM Permissions for Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/security-iam.html)

### Internal Documentation
- [MODEL_ONBOARDING.md](./MODEL_ONBOARDING.md) - General guide
- [CLAUDE.md](../CLAUDE.md) - Project conventions
- Existing implementations: `internal/generate/gemini_rest.go`, `internal/generate/vertex_sdk.go`

### Testing Resources
- [AWS SDK Mocking](https://aws.github.io/aws-sdk-go-v2/docs/unit-testing/)
- [Testify Documentation](https://github.com/stretchr/testify)

---

## Conclusion

This plan provides a complete roadmap for adding AWS Bedrock Nova Canvas to gimage. By following the phases sequentially and using the MODEL_ONBOARDING.md guide as reference, you can complete this integration in 10-16 hours with a production-ready, well-tested, fully-documented feature.

**Key Success Factors**:
1. Thorough research before coding
2. Follow existing patterns (Gemini/Vertex implementations)
3. Comprehensive testing (unit, integration, manual)
4. Clear documentation with examples
5. Backward compatibility maintained

**Ready to start?** Begin with Phase 1: Research. Read AWS Bedrock docs, set up an account, and document your findings. Then proceed through the phases systematically.

Good luck! 🚀
