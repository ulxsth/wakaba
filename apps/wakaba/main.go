package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	"github.com/yotu/wakaba/internal/handler"
)

func main() {
	lambda.Start(Dispatcher)
}

func Dispatcher(ctx context.Context, payload json.RawMessage) (interface{}, error) {
	// 1. API Gateway Request (V1 - REST API)
	var gwReq events.APIGatewayProxyRequest
	if err := json.Unmarshal(payload, &gwReq); err == nil && gwReq.RequestContext.HTTPMethod != "" {
		pubKey := os.Getenv("DISCORD_PUBLIC_KEY")
		if pubKey == "" {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "DISCORD_PUBLIC_KEY not set"}, nil
		}
		return handler.HandleGateway(ctx, gwReq, pubKey)
	}

	// 2. API Gateway Request (V2 - HTTP API)
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

	// 3. Worker Request (Async invocation)
	var workerReq handler.WorkerRequest
	if err := json.Unmarshal(payload, &workerReq); err == nil && workerReq.InteractionID != "" {
		token := os.Getenv("DISCORD_BOT_TOKEN")
		if token == "" {
			return nil, fmt.Errorf("DISCORD_BOT_TOKEN not set")
		}

		s, err := discordgo.New("Bot " + token)
		if err != nil {
			return nil, err
		}

		log.Printf("DEBUG: Received WorkerRequest: %+v", workerReq)

		if workerReq.Type == "command" {
			log.Printf("DEBUG: Dispatching command: '%s'", workerReq.CommandName)
			switch workerReq.CommandName {
			case "summarize":
				return nil, handler.ProcessSummarize(s, &workerReq)
			case "list":
				return nil, handler.ProcessTodoList(s, &workerReq)
			default:
				return nil, fmt.Errorf("unknown command: %s", workerReq.CommandName)
			}
		} else if workerReq.Type == "component" {
			return nil, handler.ProcessTodoComponent(s, &workerReq)
		}

		return nil, fmt.Errorf("unknown worker request type: %s", workerReq.Type)
	}

	return nil, fmt.Errorf("unknown event type")
}
