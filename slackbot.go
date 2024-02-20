package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)




func OauthCallbackHandler(w http.ResponseWriter, r *http.Request){
		code := r.URL.Query().Get("code")
		if code == ""{
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		 err := exchangeCodeForBotToken(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success! Bot installed."))
	

}

func exchangeCodeForBotToken(code string) (error){
	values := url.Values{
		"client_id": {configure.Slack.ClientID},
		"client_secret": {configure.Slack.ClientSecret},
		"code": {code},
	}// If there is no code, I need to get the refresh token and pass it

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", values)
	if err != nil {
		return  err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return  err
	}
	log.Printf("Raw OAuth Response: %v\n", string(body))
	var response struct {
        Ok              bool   `json:"ok"`
        BotAccessToken  string `json:"access_token"`
        BotRefreshToken string `json:"refresh_token"`
        BotTokenExpires int    `json:"expires_in"`
        Scope           string `json:"scope"`
        BotUserId       string `json:"bot_user_id"`
        AppId           string `json:"app_id"`
        Team            struct {
            Name string `json:"name"`
            Id   string `json:"id"`
        } `json:"team"`
    }


	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	// print the response to the console
	log.Printf("OAUTHRESONSE: %v\n", response)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	 _, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
        TableName: aws.String("Tokens"),
        Item: map[string]types.AttributeValue{
            "TeamId":           &types.AttributeValueMemberS{Value: response.Team.Id},
            "BotAccessToken":   &types.AttributeValueMemberS{Value: response.BotAccessToken},
            "BotRefreshToken":  &types.AttributeValueMemberS{Value: response.BotRefreshToken},
            "BotTokenExpires":  &types.AttributeValueMemberN{Value: strconv.Itoa(response.BotTokenExpires)},
            "Scope":            &types.AttributeValueMemberS{Value: response.Scope},
            "BotUserId":        &types.AttributeValueMemberS{Value: response.BotUserId},
            "AppId":            &types.AttributeValueMemberS{Value: response.AppId},
            "TeamName":         &types.AttributeValueMemberS{Value: response.Team.Name},
        },
    })

	if err != nil {
		log.Printf("Failed to store token Bot token in DynamoDB: %v", err)
		return fmt.Errorf("failed to store token in DynamoDB: %w", err)
	}
log.Printf("Bot token stored successfully for team %s", response.Team.Name)

	return nil
}

func FetchBotAuthToken(teamID string) (string, error) {
    // Load the AWS SDK config
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Printf("Error loading SDK config: %v", err)
        return "", fmt.Errorf("unable to load SDK config: %w", err)
    }

    // Create a new DynamoDB client
    svc := dynamodb.NewFromConfig(cfg)

    // Define the key for the item to retrieve
    key := map[string]types.AttributeValue{
        "TeamId": &types.AttributeValueMemberS{Value: teamID},
    }

    // Retrieve the item from the DynamoDB table
    result, err := svc.GetItem(context.TODO(), &dynamodb.GetItemInput{
        TableName: aws.String("Tokens"),
        Key:       key,
    })
    if err != nil {
        log.Printf("Error getting item from DynamoDB: %v", err)
        return "", fmt.Errorf("failed to get item from DynamoDB: %w", err)
    }

    if result.Item == nil {
        log.Printf("No item found with the key TeamID %s", teamID)
        return "", fmt.Errorf("no item found with the key TeamID %s", teamID)
    }

    // Extract the BotAccessToken from the result
    tokenAttr, exists := result.Item["BotAccessToken"]
    if !exists {
        log.Printf("Item does not contain a BotAccessToken attribute")
        return "", fmt.Errorf("item does not contain a BotAccessToken attribute")
    }
    botAccessToken, ok := tokenAttr.(*types.AttributeValueMemberS)
    if !ok {
        log.Printf("BotAccessToken attribute is not a string")
        return "", fmt.Errorf("botAccessToken attribute is not a string")
    }

    return botAccessToken.Value, nil
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	signatureHeader := r.Header.Get("X-Slack-Signature")
	timestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	log.Printf("X-Slack-Signature: %s, X-Slack-Request-Timestamp: %s\n", signatureHeader, timestampHeader)
	slackSigningSecret := configure.Slack.SigningSecret

	
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		 http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	sv, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
	if err != nil {
        log.Printf("Failed to initialize secrets verifier: %v", err)
        http.Error(w, "Verification failed", http.StatusBadRequest)
        return
    }

	if _, err := sv.Write(body); err != nil {
        log.Printf("Failed to write body to secrets verifier: %v", err)
        http.Error(w, "Verification failed", http.StatusInternalServerError)
        return
    }
	if err := sv.Ensure(); err != nil {
        log.Printf("Failed to ensure request signature: %v", err)
        http.Error(w, "Verification failed", http.StatusUnauthorized)
        return
    }

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
        log.Printf("Failed to parse events API event: %s", err)
        http.Error(w, "Failed to parse event", http.StatusInternalServerError)
        return
    }



	switch eventsAPIEvent.Type {
    case slackevents.URLVerification:
        var challengeResponse slackevents.ChallengeResponse
        if err := json.Unmarshal(body, &challengeResponse); err != nil {
            log.Printf("Failed to unmarshal url verification challenge: %v", err)
            http.Error(w, "Failed to unmarshal challenge", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(fmt.Sprintf(`{"challenge":"%s"}`, challengeResponse.Challenge)))
    case slackevents.CallbackEvent:
		teamID := eventsAPIEvent.TeamID
		userToken, err := FetchBotAuthToken(teamID)
       innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
			
		case *slackevents.AppMentionEvent:
			log.Printf("App mention event received: %+v\n", ev)// Import the package containing the publishHomeTab function




          if err != nil {
	        log.Printf("Error fetching user token: %v", err)
	        http.Error(w, "Failed to fetch user token", http.StatusInternalServerError)
	        return
           }

          case *slackevents.AppHomeOpenedEvent:
	        userID := ev.User
	        log.Printf("App home opened event received: %+v\n", ev)


            PublishHomePage(userToken, userID) 

		    w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK) // Explicitly set status code to 200 OK
            w.Write([]byte(`{"response": "Event received"}`)) 
		    log.Printf("Events: %s %s, Status: %d\n", r.Method, r.URL.Path, http.StatusOK)
		default:
			log.Printf("Unsupported inner event type received: %+v\n", ev)
		}
    default:
        log.Printf("Unsupported Events API event received: %+v", eventsAPIEvent)
        w.WriteHeader(http.StatusOK) // It's common to return OK for unhandled events to acknowledge receipt
        w.Write([]byte(`{"response": "Unsupported event received"}`))
    }
}