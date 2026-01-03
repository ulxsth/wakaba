package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yotu/wakaba/internal/handler"
)

func main() {
	lambda.Start(Dispatcher)
}

func Dispatcher(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	// Attempt to parse as API Gateway Request (V1 - REST API)
	var gwReq events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &gwReq); err == nil && gwReq.RequestContext.HTTPMethod != "" {
		pubKey := os.Getenv("DISCORD_PUBLIC_KEY")
		if pubKey == "" {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "DISCORD_PUBLIC_KEY not set"}, nil
		}
		return handler.HandleGateway(ctx, gwReq, pubKey)
	}

	// Attempt to parse as API Gateway Request (V2 - HTTP API)
	var gwV2Req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(payload, &gwV2Req); err == nil && gwV2Req.RequestContext.HTTP.Method != "" {
		pubKey := os.Getenv("DISCORD_PUBLIC_KEY")
		if pubKey == "" {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "DISCORD_PUBLIC_KEY not set"}, nil
		}

		// Map V2 to V1 structure for handler compatibility
		proxyReq := events.APIGatewayProxyRequest{
			Body:            gwV2Req.Body,
			IsBase64Encoded: gwV2Req.IsBase64Encoded,
			Headers:         gwV2Req.Headers,
			RequestContext: events.APIGatewayProxyRequestContext{
				HTTPMethod: gwV2Req.RequestContext.HTTP.Method,
			},
		}
		return handler.HandleGateway(ctx, proxyReq, pubKey)
	}

	// Attempt to parse as Worker Request
	var workerReq handler.SummaryRequest
	if err := json.Unmarshal(payload, &workerReq); err == nil && workerReq.InteractionID != "" {
		token := os.Getenv("DISCORD_BOT_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("DISCORD_BOT_TOKEN not set")
		}
		err := handler.ProcessCommand(&workerReq, token)
		return nil, err
	}

	// Unknown event
	// Log the payload for debugging? Be careful with secrets.
	// fmt.Printf("Unknown payload: %s\n", string(payload))
	return nil, fmt.Errorf("unknown event type")
}
