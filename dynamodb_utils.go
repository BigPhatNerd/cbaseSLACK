package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

func StoreUserToken(userID, username, token string, tokenExpiration time.Time) error {



	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %w", err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item: map[string]types.AttributeValue{
			"UserID":          &types.AttributeValueMemberS{Value: userID},
			"Username":        &types.AttributeValueMemberS{Value: username},
			"Token":           &types.AttributeValueMemberS{Value: token},
			"TokenExpiration": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", tokenExpiration.Unix())},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to put item, %w", err)
	}

	fmt.Println("User token stored/updated successfully")
	return nil
}