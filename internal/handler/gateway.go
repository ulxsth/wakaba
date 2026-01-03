package handler

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/bwmarrin/discordgo"
)

// API Gateway からの Webhook Request を受け取り処理するハンドラ
func HandleGateway(ctx context.Context, request events.APIGatewayProxyRequest, publicKey string) (events.APIGatewayProxyResponse, error) {
	// body をデコード
	var bodyBytes []byte
	var err error
	if request.IsBase64Encoded {
		bodyBytes, err = base64.StdEncoding.DecodeString(request.Body)
		if err != nil {
			log.Printf("Error decoding base64 body: %v", err)
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: "invalid body encoding"}, nil
		}
	} else {
		bodyBytes = []byte(request.Body)
	}

	// 署名を検証
	if !verifySignature(request.Headers, bodyBytes, publicKey) {
		log.Println("Signature verification failed")
		return events.APIGatewayProxyResponse{StatusCode: 401, Body: "invalid request signature"}, nil
	}

	// body から Interaction を解析
	var interaction discordgo.Interaction
	if err := json.Unmarshal(bodyBytes, &interaction); err != nil {
		log.Printf("Error unmarshalling interaction: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "bad request"}, nil
	}

	// ping の場合は pong 返して終了
	if interaction.Type == discordgo.InteractionPing {
		return jsonResponse(discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		})
	}

	// 本リクエストの場合は処理を呼び出す
	if interaction.Type == discordgo.InteractionApplicationCommand {
		payload := SummaryRequest{
			InteractionID:    interaction.ID,
			InteractionToken: interaction.Token,
			ChannelID:        interaction.ChannelID,
			ApplicationID:    interaction.AppID,
		}

		// Parse options
		data := interaction.ApplicationCommandData()
		for _, opt := range data.Options {
			switch opt.Name {
			case "date":
				payload.DateArg = opt.StringValue()
			case "with_title":
				payload.WithTitle = opt.BoolValue()
			}
		}

		// discord は 3秒以内にリクエストを返さないとタイムアウト判定になるので
		// 非同期で本リクエストを呼び出す
		if err := invokeSelfAsync(ctx, payload); err != nil {
			log.Printf("Failed to invoke self: %v", err)
			return jsonResponse(discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("処理の開始に失敗しました: %v", err),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		// 本処理中なので、とりあえず待ってねレスポンスを返して終了
		return jsonResponse(discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	}

	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "unknown interaction type"}, nil
}

// リクエストヘッダを解析し、Ed25519 検証器で検証する
func verifySignature(headers map[string]string, body []byte, keyHex string) bool {
	var signature, timestamp string

	// Case-insensitive header lookup
	for k, v := range headers {
		key := strings.ToLower(k)
		switch key {
		case "x-signature-ed25519":
			signature = v
		case "x-signature-timestamp":
			timestamp = v
		default:
			return false
		}
	}

	if signature == "" || timestamp == "" {
		return false
	}

	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return false
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return false
	}

	msg := append([]byte(timestamp), body...)
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return ed25519.Verify(keyBytes, msg, sigBytes)
}

func jsonResponse(body interface{}) (events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func invokeSelfAsync(ctx context.Context, payload SummaryRequest) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	client := lambda.NewFromConfig(cfg)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	functionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if functionName == "" {
		// Fallback or error. Locally we can't really invoke self easily without mock.
		log.Println("AWS_LAMBDA_FUNCTION_NAME not set")
		return nil
	}

	_, err = client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: types.InvocationTypeEvent, // Async
		Payload:        payloadBytes,
	})

	return err
}
