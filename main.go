package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BigPhatNerd/cbaseSLACK/oauth"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)



func main(){
	 tracer.Start(
        tracer.WithService("DatatdogService"),
        tracer.WithEnv("CbaseDemo"),
    )
    defer tracer.Stop()

	logFile, err := os.OpenFile("datadog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

		err = readConfig("config.yaml")
	if err != nil {
		log.Printf("Error reading config: %v", err)
	}

	tableName := "BotToken"
    interval := 11 * time.Hour
	teamID := configure.Slack.TeamID
	oauth.ScheduleTokenRotation( tableName, interval, teamID)

    // Create a traced mux router
    mux := httptrace.NewServeMux()
    // Continue using the router as you normally would.
    mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/oauth/callback", OauthCallbackHandler)
	mux.HandleFunc("/events", EventsHandler)

	port := ":4390"
	log.Printf("Server listening on port %s", port)
	err = http.ListenAndServe(port, mux)
	if err != nil{
		log.Printf("Failed to start server: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // Your handler logic here...
    w.Write([]byte("Hello World!"))
    duration := time.Since(start)
    log.Printf("Request: %s %s, Response: %d, Duration: %v\n", r.Method, r.URL.Path, http.StatusOK, duration)
}