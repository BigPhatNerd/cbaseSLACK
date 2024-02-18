package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)



func OauthCallbackHandler(w http.ResponseWriter, r *http.Request){
		code := r.URL.Query().Get("code")
		if code == ""{
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		 err := exchangeCodeForToken(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
    w.Write([]byte("Success! Bot installed."))
	

}

func exchangeCodeForToken(code string) (error){
	values := url.Values{
		"client_id": {configure.Slack.ClientID},
		"client_secret": {configure.Slack.ClientSecret},
		"code": {code},
	}

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", values)
	if err != nil {
		return  err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return  err
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
		return err
	}

	return nil
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	signatureHeader := r.Header.Get("X-Slack-Signature")
	timestampHeader := r.Header.Get("X-Slack-Request-Timestamp")
	logrus.Infof("X-Slack-Signature: %s, X-Slack-Request-Timestamp: %s\n", signatureHeader, timestampHeader)
	slackSigningSecret := configure.Slack.SigningSecret

	
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		 http.Error(w, "Failed to read request body", http.StatusBadRequest)
        return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	sv, err := slack.NewSecretsVerifier(r.Header, slackSigningSecret)
	if err != nil {
        logrus.Infof("Failed to initialize secrets verifier: %v", err)
        http.Error(w, "Verification failed", http.StatusBadRequest)
        return
    }

	if _, err := sv.Write(body); err != nil {
        logrus.Infof("Failed to write body to secrets verifier: %v", err)
        http.Error(w, "Verification failed", http.StatusInternalServerError)
        return
    }
	if err := sv.Ensure(); err != nil {
        logrus.Infof("Failed to ensure request signature: %v", err)
        http.Error(w, "Verification failed", http.StatusUnauthorized)
        return
    }

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
        logrus.Infof("Failed to parse events API event: %s", err)
        http.Error(w, "Failed to parse event", http.StatusInternalServerError)
        return
    }



	switch eventsAPIEvent.Type {
    case slackevents.URLVerification:
        var challengeResponse slackevents.ChallengeResponse
        if err := json.Unmarshal(body, &challengeResponse); err != nil {
            logrus.Infof("Failed to unmarshal url verification challenge: %v", err)
            http.Error(w, "Failed to unmarshal challenge", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(fmt.Sprintf(`{"challenge":"%s"}`, challengeResponse.Challenge)))
    case slackevents.CallbackEvent:
       innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			fmt.Printf("App mention event received: %+v\n", ev)

			// Handle app mention event, e.g., send a message back to the channel

			w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK) // Explicitly set status code to 200 OK
        w.Write([]byte(`{"response": "Event received"}`)) 
		log.Printf("Events: %s %s, Status: %d\n", r.Method, r.URL.Path, http.StatusOK)
		default:
			logrus.Infof("Unsupported inner event type received: %+v\n", ev)
		}
    default:
        logrus.Infof("Unsupported Events API event received: %+v", eventsAPIEvent)
        w.WriteHeader(http.StatusOK) // It's common to return OK for unhandled events to acknowledge receipt
        w.Write([]byte(`{"response": "Unsupported event received"}`))
    }
}