# AWS Nova Canvas Integration - Quick Start

**Last Updated**: 2025-11-02

This is your starting point for adding AWS Bedrock Nova Canvas to gimage. Read this first, then follow the detailed plans.

---

## 📚 Documentation Structure

I've created a complete documentation suite for this integration:

### 1. **This Document** (NOVA_CANVAS_QUICKSTART.md)
- High-level overview
- Quick decision guide
- Where to start

### 2. **MODEL_ONBOARDING.md** (General Guide)
- Complete process for adding ANY new model/provider
- Detailed research methodology
- Code patterns and best practices
- Testing strategies
- Use this as your reference manual

### 3. **AWS_NOVA_CANVAS_PLAN.md** (Specific Plan)
- Phase-by-phase action plan for Nova Canvas specifically
- All tasks with checkboxes
- Time estimates
- File changes summary
- Testing strategy
- Use this as your project plan

---

## 🎯 Quick Decision: Which Approach?

### Option A: Follow the Specific Plan (Recommended)
**Best for**: You want to add Nova Canvas specifically, step-by-step

**Start here**: `docs/AWS_NOVA_CANVAS_PLAN.md`

**Timeline**: 10-16 hours total
- Day 1: Research + Design (4 hours)
- Day 2: Implementation (6 hours)
- Day 3: Testing + Docs (6 hours)

**What you get**:
- Complete task list with checkboxes
- Estimated time for each phase
- Specific code examples for Nova Canvas
- File-by-file change list

### Option B: Learn the General Process First
**Best for**: You want to understand the architecture before diving in, or plan to add multiple models

**Start here**: `docs/MODEL_ONBOARDING.md`

**Timeline**: +2 hours reading, then follow specific plan

**What you get**:
- Deep understanding of gimage architecture
- Patterns applicable to any provider (Azure, Anthropic, etc.)
- Best practices and common pitfalls
- Reusable knowledge for future integrations

---

## 🚀 Recommended Workflow

### For Adding AWS Nova Canvas (Most Common)

```
1. Read this document (5 min) ✓ You are here
   ↓
2. Skim MODEL_ONBOARDING.md overview (15 min)
   → Understand the architecture
   ↓
3. Follow AWS_NOVA_CANVAS_PLAN.md phase-by-phase (10-15 hours)
   → Phase 1: Research (2-3h)
   → Phase 2: Design (1h)
   → Phase 3: Implementation (4-6h)
   → Phase 4: Testing (2-3h)
   → Phase 5: Documentation (1-2h)
   → Phase 6: Integration (1h)
   ↓
4. Create feature branch and PR
   ↓
5. Ship it! 🎉
```

### For Understanding Before Starting

```
1. Read MODEL_ONBOARDING.md fully (1-2 hours)
   → Sections: Overview, Architecture, Patterns
   ↓
2. Review existing implementations
   → internal/generate/gemini_rest.go
   → internal/generate/vertex_sdk.go
   → internal/generate/models.go
   ↓
3. Follow AWS_NOVA_CANVAS_PLAN.md
```

---

## 🎓 Key Concepts to Understand

### gimage Architecture (5-Minute Version)

**Multi-Backend System**: gimage supports multiple AI providers through a common interface:

```
User runs: gimage generate "sunset"
           ↓
CLI (generate.go) determines which API to use
           ↓
Creates appropriate client (Gemini/Vertex/Bedrock)
           ↓
Client implements: GenerateImage(ctx, prompt, options)
           ↓
Returns: *models.GeneratedImage
           ↓
CLI saves image to disk
```

**Key Files**:
- `internal/generate/` - All provider clients live here
- `internal/generate/models.go` - Model registry (metadata, pricing, capabilities)
- `internal/config/config.go` - Configuration management
- `internal/cli/generate.go` - CLI command logic
- `internal/cli/auth.go` - Authentication setup commands

**Pattern**: Each provider gets:
1. Client file: `<provider>_rest.go` or `<provider>_sdk.go`
2. Test file: `<provider>_rest_test.go`
3. Entry in model registry: `models.go`
4. Auth command: `gimage auth <provider>`
5. Config fields: `~/.gimage/config.md`

### Why This Design Works

✅ **Consistent Interface**: All clients implement same methods
✅ **Auto-Detection**: Users don't need to specify `--api` usually
✅ **Extensible**: Adding providers doesn't break existing code
✅ **Testable**: Each client can be unit tested independently
✅ **MCP Integration**: New models automatically exposed via MCP

---

## 🔑 Key Decisions Made for You

Based on thorough analysis of gimage's architecture and AWS Bedrock, these decisions are already made:

### 1. **Use AWS SDK v2** (not REST)
**Why**: AWS Signature V4 auth is complex; SDK handles it perfectly
**Trade-off**: +10MB binary size, but worth it for reliability

### 2. **File: `internal/generate/bedrock_sdk.go`**
**Why**: Follows naming convention (`*_sdk.go` for SDK-based, `*_rest.go` for REST)

### 3. **Model Name: `amazon.nova-canvas-v1:0`**
**Why**: Matches AWS's official model ID
**Aliases**: `nova-canvas`, `nova`, `bedrock-canvas` (user-friendly)

### 4. **Auth: Access Keys + AWS Profile**
**Why**: Supports both common AWS auth methods
**Priority**: env vars > config file > AWS default profile

### 5. **Priority: 7** (in model selection)
**Why**: After free Gemini models (1-2), after paid Vertex models (3-5)
**Rationale**: Paid, not free tier, good quality but not premium

### 6. **Config Format: Markdown** (existing pattern)
**Why**: gimage uses markdown config, not YAML/JSON
**Location**: `~/.gimage/config.md`

---

## 📋 What You Need Before Starting

### Prerequisites

- [x] Go 1.22+ installed
- [x] gimage codebase cloned
- [ ] AWS account with Bedrock access
- [ ] Basic AWS knowledge (IAM, regions, credentials)
- [ ] 10-16 hours available

### AWS Setup (Do This First)

1. **Create AWS Account** (if needed)
   - Go to https://aws.amazon.com/

2. **Enable Bedrock**
   - AWS Console → Amazon Bedrock
   - Choose region (e.g., us-east-1)
   - Go to "Model access"
   - Request access to "Amazon Nova Canvas"
   - Wait for approval (usually instant)

3. **Create IAM User**
   - AWS Console → IAM → Users → Create User
   - Attach policy: `AmazonBedrockFullAccess`
   - Create access key
   - Save Access Key ID and Secret Access Key securely

4. **Verify Access**
   ```bash
   # Install AWS CLI (if needed)
   aws --version

   # Configure credentials
   aws configure

   # Test Bedrock access
   aws bedrock list-foundation-models --region us-east-1
   ```

---

## 🎬 Next Steps (Choose Your Path)

### Path 1: Dive In Immediately ⚡

If you're comfortable with the overview and just want to start coding:

```bash
cd ~/dev/gimage
open docs/AWS_NOVA_CANVAS_PLAN.md
```

Follow Phase 1 → Phase 2 → ... → Phase 6

### Path 2: Learn First, Then Build 📚

If you want to deeply understand the system first:

```bash
cd ~/dev/gimage
open docs/MODEL_ONBOARDING.md  # Read thoroughly
open docs/AWS_NOVA_CANVAS_PLAN.md  # Then follow this
```

Study the architecture, then execute the plan.

### Path 3: Review Existing Code 🔍

If you learn best by reading code:

```bash
cd ~/dev/gimage

# Study existing implementations
code internal/generate/gemini_rest.go
code internal/generate/vertex_sdk.go
code internal/generate/models.go
code internal/config/config.go
code internal/cli/generate.go

# Then follow the plan
open docs/AWS_NOVA_CANVAS_PLAN.md
```

---

## 📊 Progress Tracking

Use this checklist to track your progress:

- [ ] **Pre-work**: AWS account setup, Bedrock access granted
- [ ] **Phase 1**: Research complete, API_NOTES.md created
- [ ] **Phase 2**: Design decisions documented
- [ ] **Phase 3**: Implementation complete (bedrock_sdk.go works)
- [ ] **Phase 4**: Tests pass (unit + manual)
- [ ] **Phase 5**: Documentation updated
- [ ] **Phase 6**: Feature branch ready for PR

**Estimated completion**: 10-16 hours from now

---

## 🆘 If You Get Stuck

### Common Issues

**"I don't understand the architecture"**
→ Read MODEL_ONBOARDING.md Section 2: Architecture Patterns
→ Study gemini_rest.go (simplest implementation)

**"AWS SDK is confusing"**
→ AWS_NOVA_CANVAS_PLAN.md has complete code examples
→ Check AWS SDK docs: https://aws.github.io/aws-sdk-go-v2/

**"Tests are failing"**
→ Make sure unit tests use mocks (don't call real API)
→ Check test/integration/bedrock_test.go for examples

**"Not sure what to do next"**
→ Follow AWS_NOVA_CANVAS_PLAN.md phase by phase
→ Check off each task as you complete it

**"Need more context on a specific topic"**
→ MODEL_ONBOARDING.md has deep dives on each phase
→ Review existing code in internal/generate/

### Getting Help

1. **Check existing implementations** - gemini_rest.go is simplest
2. **Search AWS Bedrock docs** - https://docs.aws.amazon.com/bedrock/
3. **Review gimage CLAUDE.md** - Project conventions
4. **Ask specific questions** - Open a GitHub discussion

---

## 🎉 Success Looks Like This

When you're done, users will be able to:

```bash
# Set up AWS credentials
gimage auth bedrock

# List available models (includes Nova Canvas)
gimage generate --list-models

# Generate an image with Nova Canvas
gimage generate --api bedrock "a mountain landscape at sunset"

# Use model aliases
gimage generate --model nova-canvas "abstract art"

# Full control
gimage generate \
  --model amazon.nova-canvas-v1:0 \
  --size 1024x1792 \
  --negative "blur, low quality" \
  --seed 42 \
  -o my-image.png
```

And the MCP server will automatically expose Nova Canvas to Claude and other MCP clients!

---

## 📈 Beyond Nova Canvas

Once you complete this integration, you'll have a proven pattern for adding:

- **Azure AI Image Generator** - Follow same process
- **Stability AI** - REST client, similar to Gemini
- **Midjourney API** - When/if they release it
- **Anthropic** - When they add image generation
- **Open source models** - Hugging Face, Replicate, etc.

The MODEL_ONBOARDING.md guide works for all of them.

---

## 🎯 TL;DR - Start Here

**For most people, do this:**

1. ✅ Read this document (you just did!)
2. [ ] Skim MODEL_ONBOARDING.md sections 1-2 (15 min)
3. [ ] Set up AWS account + Bedrock access (30 min)
4. [ ] Open AWS_NOVA_CANVAS_PLAN.md (your roadmap)
5. [ ] Start Phase 1: Research (2-3 hours)
6. [ ] Follow phases 2-6 sequentially

**Total time**: 10-16 hours → Production-ready AWS Bedrock integration

**What you get**:
- Nova Canvas model working end-to-end
- Full test coverage
- Complete documentation
- MCP integration
- Ready for PR and release

---

## 🚀 Ready? Let's Go!

```bash
# Open your roadmap
code docs/AWS_NOVA_CANVAS_PLAN.md

# Start Phase 1
# Good luck! 🎉
```

**Remember**: Follow the plan phase by phase. Don't skip ahead. Each phase builds on the previous one.

You've got this! 💪
