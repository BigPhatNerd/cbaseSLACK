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

// AppTokenRotationResponse represents the response from Slack token rotation endpoint.
type AppTokenRotationResponse struct {
	OK            bool   `json:"ok"`
	AppAuthToken     string `json:"token"`
	AppRefreshToken  string `json:"refresh_token"`
	TeamID        string `json:"team_id"`
	AppUserID        string `json:"user_id"`
	AppTokenIssuedAt      int64  `json:"iat"`
	AppTokenExpiresAt     int64  `json:"exp"`
}

// rotateToken uses the refresh token to obtain a new configuration token.
func rotateToken(refreshToken string) (*AppTokenRotationResponse, error) {
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

	var response AppTokenRotationResponse
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
        log.Printf("Error rotating token: %v", err)
		return fmt.Errorf("failed to rotate token: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
log.Printf("Error loading SDK config: %v", err)
		return fmt.Errorf("unable to load SDK config, %w", err)
	}
	svc := dynamodb.NewFromConfig(cfg)

    iatStr := strconv.FormatInt(response.AppTokenIssuedAt, 10)
    expStr := strconv.FormatInt(response.AppTokenExpiresAt, 10)
log.Printf("Response: %v", response)
update := map[string]types.AttributeValue{
        ":t": &types.AttributeValueMemberS{Value: response.AppAuthToken},
        ":r": &types.AttributeValueMemberS{Value: response.AppRefreshToken},
        ":e": &types.AttributeValueMemberN{Value: expStr},
        ":iat": &types.AttributeValueMemberN{Value: iatStr},
        ":uid": &types.AttributeValueMemberS{Value: response.AppUserID},
         
        
    }

 _, err = svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "TeamID": &types.AttributeValueMemberS{Value: response.TeamID}, 
        },
        UpdateExpression: aws.String("SET AppAuthToken = :t, AppRefreshToken = :r, AppTokenExpiresAt = :e, AppTokenIssuedAt = :iat, AppUserID = :uid"),
        ExpressionAttributeValues: update,
    })
    if err != nil {
        log.Printf("Error updating item: %v", err)
        return fmt.Errorf("failed to update item in DynamoDB: %w", err)
    }

    log.Printf("Token rotated and stored successfully for user %s", response.TeamID)
    return nil
}

func ScheduleAppTokenRotation(tableName string, interval time.Duration, teamID string) {
    
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    
        rotateTokenFunc := func() {
            refreshToken, err := getAppRefreshTokenFromStorage(teamID) 
            
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

func getAppRefreshTokenFromStorage(teamID string) (string, error) {
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
        "TeamId": &types.AttributeValueMemberS{Value: teamID},
    }

    result, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
        TableName: aws.String("Tokens"),
        Key: key,
    })
    if err != nil {
        return "", fmt.Errorf("failed to get item from DynamoDB: %w", err)
    }

    if result.Item == nil {
        return "", fmt.Errorf("no item found with the key TeamId")
    }

    tokenAttr, exists := result.Item["AppRefreshToken"]
    if !exists {
        return "", fmt.Errorf("item does not contain a Token attribute")
    }

    refreshToken, ok := tokenAttr.(*types.AttributeValueMemberS)
    if !ok {
        return "", fmt.Errorf("token attribute is not a string")
    }

    return refreshToken.Value, nil
}

func FetchAppAuthToken(teamID string)(string, error){
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Printf("Error loading SDK config: %v", err)
        return "", fmt.Errorf("unable to load SDK config: %w", err)
    }

    svc := dynamodb.NewFromConfig(cfg)

    key := map[string]types.AttributeValue{
        "TeamId": &types.AttributeValueMemberS{Value: teamID},
    }

    result, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
        TableName: aws.String("Tokens"),
        Key: key,
    })
    if err != nil{
        log.Printf("Error getting item from DynamoDB: %v", err)
        return "", fmt.Errorf("failed to get item from DynamoDB: %w", err)
    }

    if result.Item == nil {
        log.Printf("No item found with the key TeamId %s", teamID)
        return "", fmt.Errorf("no item found with the key TeamId %s", teamID)
    }

    tokenAttr, exists := result.Item["AppAuthToken"]
    if !exists{
        log.Printf("Item does not contain a Token attribute")
        return "", fmt.Errorf("item does not contain a Token attribute")
    }
    authToken, ok := tokenAttr.(*types.AttributeValueMemberS)
    if !ok{
        log.Printf("Token attribute is not a string")
        return "", fmt.Errorf("token attribute is not a string")
    }

    return authToken.Value, nil
}


