package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)



func oauthCallbackHandler(w http.ResponseWriter, r *http.Request){
		code := r.URL.Query().Get("code")
		if code == ""{
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		token, rawResponse, err := exchangeCodeForToken(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logrus.Info("Bot User OAuth Access Token:", token)
	    logrus.Infof("Raw OAuth response: %s\n", rawResponse)
    w.Write([]byte("Success! Bot installed."))

}

func exchangeCodeForToken(code string) (string, string, error){
	values := url.Values{
		"client_id": {config.Slack.ClientID},
		"client_secret": {config.Slack.ClientSecret},
		"code": {code},
	}

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", values)
	if err != nil {
		return "","", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	logrus.Infof("Raw OAuth Response: %v\n", string(body))
	var response struct {
		Ok           bool   `json:"ok"`
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		BotUserId    string `json:"bot_user_id"`
		AppId        string `json:"app_id"`
		Team         struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"team"`
		Enterprise struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"enterprise"`
		AuthedUser struct {
			Id          string `json:"id"`
			Scope       string `json:"scope"`
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
		} `json:"authed_user"`
	}


	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", string(body), err
	}

	return response.AccessToken, string(body), nil
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	signatureHeader := r.Header.Get("X-Slack-Signature")
	timestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	logrus.Infof("X-Slack-Signature: %s, X-Slack-Request-Timestamp: %s\n", signatureHeader, timestampHeader)
	slackSigningSecret := config.Slack.SigningSecret

	// logrus.Info("slacksigningsecret: %v", slackSigningSecret)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
// log.Println("Received request", string(body))
	// Reinitialize the request body to its original state after reading
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	sv, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
	if err != nil {    
		logrus.Infof("Failed to initialize secrets verifier: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := sv.Write(body); err != nil {
		logrus.Infof("Failed to write body to secrets verifier: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		logrus.Infof("Failed to ensure request signature: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		logrus.Infof("Failed to parse events API event: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}



	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var challengeResponse slackevents.ChallengeResponse
		err := json.Unmarshal(body, &challengeResponse)
		if err != nil {
			logrus.Info("Failed to unmarshal url verification challenge: %n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logrus.Infof("Challenge: %s", challengeResponse.Challenge)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"challenge":"%s"}`, challengeResponse.Challenge)))
	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			fmt.Printf("App mention event received: %+v\n", ev)
			// Handle app mention event, e.g., send a message back to the channel
		default:
			logrus.Infof("Unsupported inner event type received: %+v\n", ev)
		}
	default:
		logrus.Infof("Unsupported Events API event received: %+v\n", eventsAPIEvent)
	}
}