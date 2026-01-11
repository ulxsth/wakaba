package repository

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TodoItem struct {
	ID      string `json:"id" dynamodbav:"id"`
	Content string `json:"content" dynamodbav:"content"`
	Status  string `json:"status" dynamodbav:"status"` // "open", "done"
}

type TodoList struct {
	ChannelID string     `json:"channel_id" dynamodbav:"channel_id"`
	Items     []TodoItem `json:"items" dynamodbav:"items"`
	MessageID string     `json:"message_id" dynamodbav:"message_id"` // Pinned message ID
}

type TodoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewTodoRepository(ctx context.Context) (*TodoRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &TodoRepository{
		client: dynamodb.NewFromConfig(cfg),
		// Note: The table name should ideally be passed via environment variable or config.
		// Since we changed Terraform to use "${var.project_name}-todo", and we plan to change project_name.
		// For now, let's look for env var or default to "wakaba-production-todo" ?
		// Better: use os.Getenv("DYNAMODB_TABLE_NAME")
		tableName: getTableName(),
	}, nil
}

func getTableName() string {
	if t := os.Getenv("DYNAMODB_TABLE_NAME"); t != "" {
		return t
	}
	return "wakaba-production-todo" // updated default
}

func (r *TodoRepository) GetTodoList(ctx context.Context, channelID string) (*TodoList, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"channel_id": &types.AttributeValueMemberS{Value: channelID},
		},
	})
	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return &TodoList{ChannelID: channelID, Items: []TodoItem{}}, nil
	}

	var list TodoList
	if err := attributevalue.UnmarshalMap(out.Item, &list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (r *TodoRepository) SaveTodoList(ctx context.Context, list *TodoList) error {
	item, err := attributevalue.MarshalMap(list)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}

// Helper to avoid circular dependency in imports if needed, but for now sticking to strict types
