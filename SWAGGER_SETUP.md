# Swagger UI Setup for Gimage API

Multiple ways to serve interactive API documentation using Swagger UI.

## Why Swagger UI?

- **Interactive Documentation**: Test API endpoints directly in the browser
- **Auto-Generated**: Automatically created from `openapi.yaml`
- **Developer-Friendly**: "Try it out" feature for all endpoints
- **Zero Code**: No need to write client code to test
- **Shareable**: Send URL to developers for instant API access

---

## Option 1: Local Development (Fastest)

### Using Docker

```bash
# From gimage directory
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/openapi.yaml \
  -v $(pwd):/usr/share/nginx/html \
  swaggerapi/swagger-ui

# Open: http://localhost:8080
```

### Using npx (No Docker)

```bash
# Install globally
npm install -g swagger-ui-watcher

# Serve with hot reload
swagger-ui-watcher openapi.yaml

# Open: http://localhost:8000
```

### Using Python

```bash
# Install
pip install connexion[swagger-ui]

# Create serve_swagger.py (see below)
python serve_swagger.py

# Open: http://localhost:8080/ui
```

**serve_swagger.py**:
```python
import connexion

app = connexion.FlaskApp(__name__, specification_dir='./')
app.add_api('openapi.yaml')

if __name__ == '__main__':
    app.run(port=8080)
```

---

## Option 2: Static HTML (Recommended for Sharing)

Generate a standalone HTML file that can be hosted anywhere.

### Using Redoc

```bash
# Install
npm install -g redoc-cli

# Generate beautiful single-page docs
redoc-cli bundle openapi.yaml -o api-docs.html

# Serve locally
python -m http.server 8080

# Open: http://localhost:8080/api-docs.html
```

### Using Swagger UI (Static)

Create `swagger-ui.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Gimage API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: './openapi.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>
```

Then serve:
```bash
python -m http.server 8080
# Open: http://localhost:8080/swagger-ui.html
```

---

## Option 3: S3 + CloudFront (Production)

Host Swagger UI on AWS with your Lambda API.

### Setup

```bash
# 1. Create docs directory
mkdir -p docs/swagger-ui

# 2. Download Swagger UI
curl -L https://github.com/swagger-api/swagger-ui/archive/v5.10.0.tar.gz | tar xz
cp -r swagger-ui-5.10.0/dist/* docs/swagger-ui/

# 3. Copy your OpenAPI spec
cp openapi.yaml docs/swagger-ui/

# 4. Update docs/swagger-ui/swagger-initializer.js
# Replace:
#   url: "https://petstore.swagger.io/v2/swagger.json",
# With:
#   url: "./openapi.yaml",

# 5. Upload to S3
aws s3 sync docs/swagger-ui s3://gimage-docs/api/ --acl public-read

# 6. Access via S3 URL or CloudFront
# https://gimage-docs.s3.amazonaws.com/api/index.html
```

### CDK Stack Addition

Add to your `infrastructure/cdk/lib/gimage-stack.ts`:

```typescript
// S3 bucket for API documentation
const docsBucket = new s3.Bucket(this, 'GimageDocsBucket', {
  bucketName: `gimage-docs-${this.account}`,
  websiteIndexDocument: 'index.html',
  publicReadAccess: true,
  blockPublicAccess: new s3.BlockPublicAccess({
    blockPublicAcls: false,
    blockPublicPolicy: false,
    ignorePublicAcls: false,
    restrictPublicBuckets: false,
  }),
  cors: [
    {
      allowedMethods: [s3.HttpMethods.GET],
      allowedOrigins: ['*'],
      allowedHeaders: ['*'],
    },
  ],
});

// CloudFront distribution for docs
const docsDistribution = new cloudfront.Distribution(this, 'DocsDistribution', {
  defaultBehavior: {
    origin: new origins.S3Origin(docsBucket),
    viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
  },
  defaultRootObject: 'index.html',
});

// Output the docs URL
new cdk.CfnOutput(this, 'DocsUrl', {
  value: `https://${docsDistribution.distributionDomainName}`,
  description: 'API Documentation URL',
});
```

---

## Option 4: API Gateway Integration (Built-In!)

✅ **Already implemented!** Swagger UI is built into the Lambda API.

After deploying, access interactive docs at:
```
https://YOUR_API_URL/prod/docs
```

The OpenAPI spec is available at:
```
https://YOUR_API_URL/prod/openapi.yaml
```

**Note**: The OpenAPI spec is currently served as a placeholder. To embed the full spec:

1. Copy openapi.yaml to internal/lambdahandler/:
   ```bash
   cp openapi.yaml internal/lambdahandler/
   ```

2. Update `internal/lambdahandler/docs.go`:
   ```go
   //go:embed openapi.yaml
   var openapiSpec string
   ```

3. Rebuild: `make build-lambda`

Alternatively, serve the spec from S3 (recommended for production):

### Option 4A: API Gateway Integration (Full Implementation)

Serve Swagger UI directly from your API Gateway endpoint with the spec loaded from S3 or embedded.

### Add `/docs` Endpoint

**Update `internal/lambdahandler/handler.go`**:

```go
func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// ... existing code ...

	switch routeKey {
	case "GET /docs":
		return h.handleDocs(ctx)
	case "GET /openapi.yaml":
		return h.handleOpenAPISpec(ctx)
	// ... existing routes ...
	}
}
```

**Add handler methods**:

```go
func (h *Handler) handleDocs(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Gimage API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: './openapi.yaml',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
      });
    };
  </script>
</body>
</html>`

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
			"Access-Control-Allow-Origin": "*",
		},
		Body: html,
	}, nil
}

func (h *Handler) handleOpenAPISpec(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	// Embed openapi.yaml at build time
	spec, err := os.ReadFile("openapi.yaml")
	if err != nil {
		return errorResponse(500, "OpenAPI spec not found"), nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/yaml",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(spec),
	}, nil
}
```

**Embed OpenAPI spec in binary**:

```go
// Add to top of handlers.go
import _ "embed"

//go:embed ../../openapi.yaml
var openapiSpec string

// Update handleOpenAPISpec
func (h *Handler) handleOpenAPISpec(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/yaml",
			"Access-Control-Allow-Origin": "*",
		},
		Body: openapiSpec,
	}, nil
}
```

**Update CDK to add routes**:

```typescript
const docsResource = api.root.addResource('docs');
docsResource.addMethod('GET', lambdaIntegration);

const openapiResource = api.root.addResource('openapi.yaml');
openapiResource.addMethod('GET', lambdaIntegration);
```

**Access**:
```
https://your-api-url/prod/docs
```

---

## Option 5: GitHub Pages (Free Hosting)

Host documentation on GitHub Pages for free.

### Setup

```bash
# 1. Create docs branch
git checkout -b gh-pages

# 2. Create docs directory structure
mkdir -p docs
cd docs

# 3. Download Swagger UI
curl -L https://github.com/swagger-api/swagger-ui/archive/v5.10.0.tar.gz | tar xz
cp -r swagger-ui-5.10.0/dist/* .
rm -rf swagger-ui-5.10.0

# 4. Copy OpenAPI spec
cp ../openapi.yaml .

# 5. Update swagger-initializer.js
# Change url to: "./openapi.yaml"

# 6. Commit and push
git add .
git commit -m "Add Swagger UI documentation"
git push origin gh-pages

# 7. Enable GitHub Pages in repository settings
# Settings → Pages → Source: gh-pages branch, /docs folder
```

**Access**:
```
https://YOUR_USERNAME.github.io/gimage/
```

---

## Comparison

| Method | Pros | Cons | Best For |
|--------|------|------|----------|
| **Docker** | Instant, no install | Requires Docker | Quick local testing |
| **Static HTML** | Works anywhere | Manual updates | Sharing with team |
| **S3 + CloudFront** | Fast, scalable | AWS costs | Production docs |
| **API Gateway** | Same domain as API | Increases Lambda size | Production API |
| **GitHub Pages** | Free, automatic | Public only | Open source projects |

---

## Recommended Setup

### Development
```bash
# Quick local testing
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/openapi.yaml \
  -v $(pwd):/usr/share/nginx/html \
  swaggerapi/swagger-ui
```

### Production (Recommended)

**Option A: Separate S3 + CloudFront** (Best)
- Dedicated documentation site
- Fast global delivery
- Independent of API
- Low cost (~$0.50/month)

**Option B: API Gateway `/docs` endpoint**
- Unified API + docs
- Same authentication
- No separate infrastructure
- Slightly slower (Lambda cold starts)

---

## Testing the API with Swagger UI

Once Swagger UI is running:

1. **Open in browser**: Navigate to Swagger UI URL
2. **Authorize** (if API keys enabled): Click "Authorize" button
3. **Try endpoints**:
   - Click on any endpoint (e.g., POST /generate)
   - Click "Try it out"
   - Fill in parameters
   - Click "Execute"
4. **View response**: See real API response with status code, headers, body

### Example: Generate Image

1. Expand `POST /generate`
2. Click "Try it out"
3. Edit request body:
   ```json
   {
     "prompt": "a sunset over mountains",
     "size": "1024x1024",
     "response_format": "s3_url"
   }
   ```
4. Click "Execute"
5. See generated image URL in response

---

## Customization

### Custom Theme

Create `custom.css`:

```css
.swagger-ui .topbar {
  background-color: #1e293b;
}

.swagger-ui .info .title {
  color: #3b82f6;
}
```

Reference in HTML:

```html
<link rel="stylesheet" href="custom.css" />
```

### Custom Logo

Update HTML:

```html
<div id="swagger-ui"></div>
<script>
  window.ui = SwaggerUIBundle({
    url: './openapi.yaml',
    dom_id: '#swagger-ui',
    // Add custom logo
    customCss: '.topbar-wrapper img { content: url("logo.png"); }',
  });
</script>
```

---

## Automation

### Auto-Deploy on Push

**GitHub Actions** (`.github/workflows/deploy-docs.yml`):

```yaml
name: Deploy Swagger UI

on:
  push:
    branches: [main]
    paths:
      - 'openapi.yaml'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download Swagger UI
        run: |
          curl -L https://github.com/swagger-api/swagger-ui/archive/v5.10.0.tar.gz | tar xz
          cp -r swagger-ui-5.10.0/dist/* docs/
          cp openapi.yaml docs/

      - name: Upload to S3
        run: |
          aws s3 sync docs/ s3://gimage-docs/api/ --acl public-read
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}

      - name: Invalidate CloudFront
        run: |
          aws cloudfront create-invalidation \
            --distribution-id ${{ secrets.CLOUDFRONT_DIST_ID }} \
            --paths "/*"
```

---

## Next Steps

1. **Choose deployment method** (recommended: S3 + CloudFront for production)
2. **Set up Swagger UI** using one of the options above
3. **Test all endpoints** interactively
4. **Share URL** with developers
5. **Add to README** as primary documentation link

---

## Benefits for Developers

- ✅ **No Postman needed** - Test directly in browser
- ✅ **No client code needed** - Instant API exploration
- ✅ **Always up-to-date** - Auto-generated from OpenAPI spec
- ✅ **Beautiful UI** - Professional, standardized interface
- ✅ **Sharable** - Send URL to anyone

---

## Support

- Swagger UI Docs: https://swagger.io/tools/swagger-ui/
- OpenAPI Spec: https://spec.openapis.org/oas/v3.0.3
- Redoc (Alternative): https://redocly.com/docs/redoc/
