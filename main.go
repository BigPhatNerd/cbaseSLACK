package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/BigPhatNerd/cbaseSLACK/oauth"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func main(){
	//log to a file
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
    logger.Fatal("Failed to open log file", err)
}
logger.SetOutput(logFile)
logger.SetFormatter(&logrus.JSONFormatter{})
logger.SetLevel(logrus.InfoLevel)

mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello, World"))
})

loggedMux := LogRequests(mux)


	err = readConfig("config.yaml")
	if err != nil {
		logger.Infof("Error reading config: %v", err)
	}
token := oauth.TokenExchange()
fmt.Println(token)
	http.HandleFunc("/oauth/callback", oauthCallbackHandler)
	http.HandleFunc("/events", eventsHandler)

	logger.Infof("Server started ...")
	err = http.ListenAndServe(":4390", loggedMux)
	if err != nil{
		logger.Infof("Failed to start server: %v", err)
	}
}