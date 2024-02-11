package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

// TokenRotationResponse represents the response from Slack token rotation endpoint.
type TokenRotationResponse struct {
	OK            bool   `json:"ok"`
	Token         string `json:"token"`
	RefreshToken  string `json:"refresh_token"`
	TeamID        string `json:"team_id"`
	UserID        string `json:"user_id"`
	IssuedAt      int64  `json:"iat"`
	ExpiresAt     int64  `json:"exp"`
}

// RotateToken uses the refresh token to obtain a new configuration token.
func RotateToken(refreshToken string) (*TokenRotationResponse, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("GET", "https://slack.com/api/tooling.tokens.rotate", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add refresh_token as query parameter
	req.URL.RawQuery = data.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var response TokenRotationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("slack API reported an error")
	}

	return &response, nil
}

func rotateAndStoreToken(refreshToken, tableName string) error{

	response, err := RotateToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to rotate token: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %w", err)
	}
	svc := dynamodb.NewFromConfig(cfg)

update := map[string]types.AttributeValue{
	":t": &types.AttributeValueMemberS{Value: response.Token},
        ":r": &types.AttributeValueMemberS{Value: response.RefreshToken},
        ":e": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", response.ExpiresAt)},
}

 _, err = svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "UserID": &types.AttributeValueMemberS{Value: response.UserID}, // Assuming you use UserID as the key
        },
        UpdateExpression: aws.String("SET Token = :t, RefreshToken = :r, ExpiresAt = :e"),
        ExpressionAttributeValues: update,
    })
    if err != nil {
        return fmt.Errorf("failed to update item in DynamoDB: %w", err)
    }

    log.Printf("Token rotated and stored successfully for user %s", response.UserID)
    return nil
}

func scheduleTokenRotation(refreshToken, tableName string, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-ticker.C:
                err := rotateAndStoreToken(refreshToken, tableName)
                if err != nil {
                    log.Printf("Error rotating token: %v", err)
                }
            }
        }
    }()
}

