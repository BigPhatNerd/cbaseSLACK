package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

// TokenRotationResponse represents the response from Slack token rotation endpoint.
type TokenRotationResponse struct {
	OK            bool   `json:"ok"`
	AuthToken     string `json:"auth_token"`
	RefreshToken  string `json:"refresh_token"`
	TeamID        string `json:"team_id"`
	UserID        string `json:"user_id"`
	IssuedAt      int64  `json:"iat"`
	ExpiresAt     int64  `json:"exp"`
}

// rotateToken uses the refresh token to obtain a new configuration token.
func rotateToken(refreshToken string) (*TokenRotationResponse, error) {
	data := url.Values{}
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("GET", "https://slack.com/api/tooling.tokens.rotate", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

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
log.Printf("Response: %v", string(body))
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

	response, err := rotateToken(refreshToken)
	if err != nil {
		return fmt.Errorf("failed to rotate token: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %w", err)
	}
	svc := dynamodb.NewFromConfig(cfg)

    iatStr := strconv.FormatInt(response.IssuedAt, 10)
    expStr := strconv.FormatInt(response.ExpiresAt, 10)

update := map[string]types.AttributeValue{
        ":t": &types.AttributeValueMemberS{Value: response.AuthToken},
        ":r": &types.AttributeValueMemberS{Value: response.RefreshToken},
        ":e": &types.AttributeValueMemberN{Value: expStr},
        ":iat": &types.AttributeValueMemberN{Value: iatStr},
        ":uid": &types.AttributeValueMemberS{Value: response.UserID},
        ":ok":  &types.AttributeValueMemberBOOL{Value: response.OK}, 
        
    }

 _, err = svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "TeamID": &types.AttributeValueMemberS{Value: response.TeamID}, 
        },
        UpdateExpression: aws.String("SET AuthToken = :t, RefreshToken = :r, ExpiresAt = :e, IssuedAt = :iat, UserID = :uid, OK = :ok"),
        ExpressionAttributeValues: update,
    })
    if err != nil {
        return fmt.Errorf("failed to update item in DynamoDB: %w", err)
    }

    log.Printf("Token rotated and stored successfully for user %s", response.TeamID)
    return nil
}

func ScheduleTokenRotation(tableName string, interval time.Duration, teamID string) {
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    
        rotateTokenFunc := func() {
            refreshToken, err := getRefreshTokenFromStorage(teamID) 
            // var refreshToken string
            // var err error
            // refreshToken = "xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjI4Njc1ODUxNjg2LTgyYzNlZTlmNTRiYTM0ZTY0MjZlNTJiYTgyZWM3YzExOWVhZmNlMjIzMmJhMGRjZjgxMzU5MWNhNmZhNDY1YWU"
            if err != nil {
                log.Printf("Error retrieving refresh token: %v", err)
                return
            }
    		
            err = rotateAndStoreToken(refreshToken, tableName)
            if err != nil {
                log.Printf("Error rotating token: %v", err)
            }
        }

    rotateTokenFunc()

    go func() {
        for {
            select {
            case <-ticker.C:
                rotateTokenFunc()
            }
        }
    }()
}

func getRefreshTokenFromStorage(teamID string) (string, error) {
  if envToken := os.Getenv("HEROKU_REFRESH_TOKEN"); envToken != "" {
        log.Println("Using refresh token from environment variable")
        return envToken, nil
    }
    
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        return "", fmt.Errorf("unable to load SDK config: %w", err)
    }

    svc := dynamodb.NewFromConfig(cfg)

    
    key := map[string]types.AttributeValue{
        "TeamID": &types.AttributeValueMemberS{Value: teamID},
    }

    result, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
        TableName: aws.String("BotToken"),
        Key: key,
    })
    if err != nil {
        return "", fmt.Errorf("failed to get item from DynamoDB: %w", err)
    }

    if result.Item == nil {
        return "", fmt.Errorf("no item found with the key TeamID")
    }

    tokenAttr, exists := result.Item["RefreshToken"]
    if !exists {
        return "", fmt.Errorf("item does not contain a Token attribute")
    }

    refreshToken, ok := tokenAttr.(*types.AttributeValueMemberS)
    if !ok {
        return "", fmt.Errorf("token attribute is not a string")
    }
log.Printf("refreshToken: %v\n\n\n\n\n", refreshToken.Value)
    return refreshToken.Value, nil
}


