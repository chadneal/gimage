package lambdahandler

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// successResponse creates a successful API Gateway proxy response
func successResponse(statusCode int, body interface{}) events.APIGatewayProxyResponse {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to marshal response: %v", err))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    corsHeaders(),
		Body:       string(jsonBody),
	}
}

// errorResponse creates an error API Gateway proxy response
func errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	errorResp := ErrorResponse{
		Error:   httpStatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	jsonBody, _ := json.Marshal(errorResp)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    corsHeaders(),
		Body:       string(jsonBody),
	}
}

// corsHeaders returns standard CORS headers for API responses
func corsHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*", // TODO: Configure from env var
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Key",
	}
}

// httpStatusText returns a human-readable status text for HTTP status codes
func httpStatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	case 503:
		return "Service Unavailable"
	default:
		return fmt.Sprintf("HTTP %d", code)
	}
}
