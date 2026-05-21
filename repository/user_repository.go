package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"users-api/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const timeLayout = time.RFC3339

type UserRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewUserRepository(client *dynamodb.Client) *UserRepository {
	table := os.Getenv("DYNAMODB_TABLE")
	if table == "" {
		table = "users"
	}
	return &UserRepository{client: client, tableName: table}
}

func (r *UserRepository) Create(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	now := time.Now().UTC().Format(timeLayout)
	user := models.User{
		ID:        uuid.NewString(),
		Name:      input.Name,
		CPF:       input.CPF,
		Email:     input.Email,
		Birthdate: input.Birthdate,
		Phone:     input.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("marshal user: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(cpf)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if isErrorType(err, &ccf) {
			return nil, fmt.Errorf("cpf already registered")
		}
		return nil, fmt.Errorf("put item: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	out, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	var users []models.User
	if err = attributevalue.UnmarshalListOfMaps(out.Items, &users); err != nil {
		return nil, fmt.Errorf("unmarshal users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return nil, nil
	}

	var user models.User
	if err = attributevalue.UnmarshalMap(out.Item, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, input models.UpdateUserInput) (*models.User, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Email != nil {
		existing.Email = *input.Email
	}
	if input.Birthdate != nil {
		existing.Birthdate = *input.Birthdate
	}
	if input.Phone != nil {
		existing.Phone = *input.Phone
	}
	existing.UpdatedAt = time.Now().UTC().Format(timeLayout)

	expr := "SET #name = :name, email = :email, birthdate = :birthdate, phone = :phone, updated_at = :updated_at"
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String(expr),
		ExpressionAttributeNames: map[string]string{
			"#name": "name", // 'name' is a reserved word in DynamoDB
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name":       &types.AttributeValueMemberS{Value: existing.Name},
			":email":      &types.AttributeValueMemberS{Value: existing.Email},
			":birthdate":  &types.AttributeValueMemberS{Value: existing.Birthdate},
			":phone":      &types.AttributeValueMemberS{Value: existing.Phone},
			":updated_at": &types.AttributeValueMemberS{Value: existing.UpdatedAt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}

	return existing, nil
}

func isErrorType(err error, target interface{}) bool {
	switch t := target.(type) {
	case **types.ConditionalCheckFailedException:
		if e, ok := err.(*types.ConditionalCheckFailedException); ok {
			*t = e
			return true
		}
	}
	return false
}
