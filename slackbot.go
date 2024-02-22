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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Review struct {
	UserName         string `dynamodbav:"UserName"`
	EmployeeSelected string `dynamodbav:"EmployeeSelected"`
	Feedback         string `dynamodbav:"Feedback"`
}

func OauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
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

func exchangeCodeForBotToken(code string) error {
	values := url.Values{
		"client_id":     {configure.Slack.ClientID},
		"client_secret": {configure.Slack.ClientSecret},
		"code":          {code},
	} // If there is no code, I need to get the refresh token and pass it

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", values)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
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
			"TeamId":          &types.AttributeValueMemberS{Value: response.Team.Id},
			"BotAccessToken":  &types.AttributeValueMemberS{Value: response.BotAccessToken},
			"BotRefreshToken": &types.AttributeValueMemberS{Value: response.BotRefreshToken},
			"BotTokenExpires": &types.AttributeValueMemberN{Value: strconv.Itoa(response.BotTokenExpires)},
			"Scope":           &types.AttributeValueMemberS{Value: response.Scope},
			"BotUserId":       &types.AttributeValueMemberS{Value: response.BotUserId},
			"AppId":           &types.AttributeValueMemberS{Value: response.AppId},
			"TeamName":        &types.AttributeValueMemberS{Value: response.Team.Name},
		},
	})

	if err != nil {
		log.Printf("Failed to store token Bot token in DynamoDB: %v", err)
		return fmt.Errorf("failed to store token in DynamoDB: %w", err)
	}
	log.Printf("Bot token stored successfully for team %s", response.Team.Name)

	return nil
}

func scheduleRefreshBotToken(ctx context.Context, teamID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := RefreshBotToken(ctx, teamID); err != nil {
					log.Printf("Error refreshing bot token: %v", err)
				} else {
					log.Println("Bot token refreshed successfully")
				}
			}
		}
	}()
}

func RefreshBotToken(ctx context.Context, teamID string) error {
	botRefreshToken, err := FetchBotRefreshToken(ctx, teamID)
	if err != nil {
		log.Printf("Error fetching bot refresh token: %v", err)

	}
	values := url.Values{
		"client_id":     {configure.Slack.ClientID},
		"client_secret": {configure.Slack.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {botRefreshToken},
	}

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", values)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return err
	}

	var response struct {
		OK              bool   `josn:"ok"`
		BotAccessToken  string `json:"access_token"`
		BotRefreshToken string `json:"refresh_token"`
		BotTokenExpires int    `json:"expires_in"`
		BotUserId       string `json:"bot_user_id"`
		AppId           string `json:"app_id"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return err
	}

	if !response.OK {
		log.Printf("failed to refresh bot token: %v", string(body))
		return fmt.Errorf("failed to refresh bot token: %s", string(body))
	}

	return updateBotTokensInDynamoDB(ctx, teamID, response.BotAccessToken, response.BotRefreshToken, response.BotTokenExpires, response.BotUserId, response.AppId)
}

func updateBotTokensInDynamoDB(ctx context.Context, teamID, botAccessToken, botRefreshToken string, botTokenExpires int, botUserId, appId string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Failed to load SDK config: %v", err)
		return err
	}
	expiryTimestamp := time.Now().Add(time.Second * time.Duration(botTokenExpires)).Unix()
	svc := dynamodb.NewFromConfig(cfg)

	_, err = svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("Tokens"),
		Key: map[string]types.AttributeValue{
			"TeamId": &types.AttributeValueMemberS{Value: teamID},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bat":  &types.AttributeValueMemberS{Value: botAccessToken},
			":brt":  &types.AttributeValueMemberS{Value: botRefreshToken},
			":bte":  &types.AttributeValueMemberN{Value: strconv.FormatInt(expiryTimestamp, 10)},
			":bui":  &types.AttributeValueMemberS{Value: botUserId},
			":aid":  &types.AttributeValueMemberS{Value: appId},
			":expr": &types.AttributeValueMemberN{Value: strconv.FormatInt(expiryTimestamp, 10)},
		},
		UpdateExpression: aws.String("SET BotAccessToken = :bat, BotRefreshToken = :brt, BotTokenExpires = :bte, BotUserId = :bui, AppId = :aid, ExpiryTimestamp = :expr"),
		ReturnValues:     types.ReturnValueAllNew,
	})

	if err != nil {
		log.Printf("Failed to update tokens in DynamoDB: %v", err)
		return err
	}

	log.Printf("Tokens updated successfully in DynamoDB")
	return nil
}

func FetchBotRefreshToken(ctx context.Context, teamID string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading AWS SDK config: %v", err)
		return "", fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	key := map[string]types.AttributeValue{
		"TeamId": &types.AttributeValueMemberS{Value: teamID},
	}

	// Retrieve the item from the DynamoDB table
	result, err := svc.GetItem(ctx, &dynamodb.GetItemInput{
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

	// Extract the BotRefreshToken from the result
	tokenAttr, exists := result.Item["BotRefreshToken"]
	if !exists {
		log.Printf("Item does not contain a BotRefreshToken attribute")
		return "", fmt.Errorf("item does not contain a BotRefreshToken attribute")
	}
	refreshToken, ok := tokenAttr.(*types.AttributeValueMemberS)
	if !ok {
		log.Printf("BotRefreshToken attribute is not a string")
		return "", fmt.Errorf("BotRefreshToken attribute is not a string")
	}

	return refreshToken.Value, nil
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
	log.Printf("Request body: %s\n", body)
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
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {

		case *slackevents.AppMentionEvent:
			log.Printf("App mention event received: %+v\n", ev) // Import the package containing the publishHomeTab function

			if err != nil {
				log.Printf("Error fetching user token: %v", err)
				http.Error(w, "Failed to fetch user token", http.StatusInternalServerError)
				return
			}

		case *slackevents.AppHomeOpenedEvent:
			userID := ev.User
			log.Printf("App home opened event received: %+v\n", ev)

			PublishHomePage(userID, nil)

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

func InteractionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	payload := r.FormValue("payload")
	var callback slack.InteractionCallback
	if err := json.Unmarshal([]byte(payload), &callback); err != nil {
		log.Printf("Could not parse interaction payload: %v", err)
		http.Error(w, "Could not parse interaction payload", http.StatusBadRequest)
		return
	}

	switch callback.Type {
	case slack.InteractionTypeViewSubmission:

		userID := callback.User.ID
		userName := callback.User.Name

		values := callback.View.State.Values
		employeeSelected := values["employee_select"]["employee_select_action"].SelectedOption.Value
		feedback := values["feedback"]["feedback_input"].Value
		log.Printf("Employee selected: %s, Feedback: %s\n, UserID: %s\n, userName: %s\n", employeeSelected, feedback, userID, userName)
		err := storeSurveyData(userID, userName, employeeSelected, feedback)
		if err != nil {
			log.Printf("Error storing survey data: %v", err)
			http.Error(w, "Error storing data", http.StatusInternalServerError)
			return
		}

		go func() {
			if err := showSuccessModal(callback.TriggerID); err != nil {
				log.Printf("Error showing success modal: %v", err)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return

	case slack.InteractionTypeBlockActions:
		for _, action := range callback.ActionCallback.BlockActions {
			switch action.ActionID {
			case "create_action":
				options := []*slack.OptionBlockObject{
					slack.NewOptionBlockObject("ted_smith", slack.NewTextBlockObject("plain_text", "Ted Smith", false, false), nil),
					slack.NewOptionBlockObject("janet_kelso", slack.NewTextBlockObject("plain_text", "Janet Kelso", false, false), nil),
					slack.NewOptionBlockObject("mike_brown", slack.NewTextBlockObject("plain_text", "Mike Brown", false, false), nil),
					slack.NewOptionBlockObject("wilson_horrell", slack.NewTextBlockObject("plain_text", "Wilson Horrell", false, false), nil),
				}

				element := slack.NewOptionsSelectBlockElement("static_select", slack.NewTextBlockObject("plain_text", "Select an option...", false, false), "employee_select_action", options...)

				inputBlock := slack.NewInputBlock("employee_select", slack.NewTextBlockObject("plain_text", "Select an Employee", false, false), nil, element)

				// Define the modal
				modalRequest := slack.ModalViewRequest{
					Type:   "modal",
					Title:  slack.NewTextBlockObject("plain_text", "Feedback Form", false, false),
					Close:  slack.NewTextBlockObject("plain_text", "Cancel", false, false),
					Submit: slack.NewTextBlockObject("plain_text", "Submit", false, false),
					Blocks: slack.Blocks{
						BlockSet: []slack.Block{
							inputBlock,
							slack.NewInputBlock(
								"feedback", // Block ID
								slack.NewTextBlockObject("plain_text", "Feedback", false, false), // Label
								nil, // Hint (optional, can be nil)
								&slack.PlainTextInputBlockElement{ // Element
									Type:        "plain_text_input",
									ActionID:    "feedback_input",
									Placeholder: slack.NewTextBlockObject("plain_text", "Enter your feedback here...", false, false),
									Multiline:   true,
								},
							),
						},
					},
				}
				modalRequestJSON, err := json.Marshal(modalRequest)
				if err != nil {
					log.Printf("Error marshalling modal request: %v", err)
					http.Error(w, "Failed to marshal modal request", http.StatusInternalServerError)
				}
				log.Printf("Modal request: %s\n", modalRequestJSON)
				log.Printf("Callback trigger ID: %s\n", callback.TriggerID)
				log.Printf("Type of triggerID: %T\n", callback.TriggerID)
				_, err = api.OpenView(callback.TriggerID, modalRequest)
				if err != nil {
					log.Printf("Error opening modal: %v", err)
					http.Error(w, "Failed to open modal", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response": "Modal opened"}`))
				return

			case "view_action":
				// Fetch the last 10 reviews from DynamoDB
				reviews, err := fetchLast10Reviews()
				if err != nil {
					log.Printf("Error fetching reviews: %v", err)
					http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
					return
				}

				// Publish the home page with the reviews
				go PublishHomePage(callback.User.ID, reviews)

				// Respond immediately to the button click without waiting for the reviews to be displayed
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
				return

			case "remove_reviews_action":
				go PublishHomePage(callback.User.ID, []Review{})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{}"))
				return
			}
		}

	}
}

func storeSurveyData(userID, userName, employeeSelected, feedback string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	svc := dynamodb.NewFromConfig(cfg)
	submissionID := uuid.New().String()

	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("SurveyData"),
		Item: map[string]types.AttributeValue{
			"SubmissionID":         &types.AttributeValueMemberS{Value: submissionID},
			"ConstantPartitionKey": &types.AttributeValueMemberS{Value: "ALL"},
			"UserID":               &types.AttributeValueMemberS{Value: userID},
			"UserName":             &types.AttributeValueMemberS{Value: userName},
			"EmployeeSelected":     &types.AttributeValueMemberS{Value: employeeSelected},
			"Feedback":             &types.AttributeValueMemberS{Value: feedback},
			"Timestamp":            &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %v", err)
	}

	return nil
}

func showSuccessModal(triggerID string) error {
	responseModal := createSuccessModal()

	_, err := api.OpenView(triggerID, responseModal)
	if err != nil {
		return fmt.Errorf("failed to open success modal: %v", err)
	}
	return err // Return the error to be handled by the caller
}

func createSuccessModal() slack.ModalViewRequest {

	titleText := slack.NewTextBlockObject("plain_text", "Success", false, false)
	sectionText := slack.NewTextBlockObject("mrkdwn", "Your feedback has been successfully submitted. Thank you!", false, false)
	section := slack.NewSectionBlock(sectionText, nil, nil)
	blocks := slack.Blocks{
		BlockSet: []slack.Block{section},
	}
	return slack.ModalViewRequest{
		Type:   "modal",
		Title:  titleText,
		Blocks: blocks,
		Close:  slack.NewTextBlockObject("plain_text", "Close", false, false),
	}
}

func fetchLast10Reviews() ([]Review, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Unable to load SDK config: %v", err)
		return nil, err
	}
	svc := dynamodb.NewFromConfig(cfg)

	// Assuming you have a GSI on Timestamp to fetch the latest reviews
	// This example does not include pagination, error handling, or GSI details
	out, err := svc.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("SurveyData"),
		IndexName:              aws.String("TimestampIndex"),
		ScanIndexForward:       aws.Bool(false), // false for descending order
		Limit:                  aws.Int32(10),
		KeyConditionExpression: aws.String("ConstantPartitionKey = :cpk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cpk": &types.AttributeValueMemberS{Value: "ALL"},
		},
	})
	if err != nil {
		return nil, err
	}

	var reviews []Review
	for _, item := range out.Items {
		review := Review{}
		err := attributevalue.UnmarshalMap(item, &review)
		if err != nil {
			log.Printf("Failed to unmarshal DynamoDB item to Review: %v", err)
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}
