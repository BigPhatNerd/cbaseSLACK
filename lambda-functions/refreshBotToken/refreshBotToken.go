package main

import (
	"context"
	"log"

	cbaseSLACK "github.com/BigPhatNerd/cbaseSLACK"
	"github.com/aws/aws-lambda-go/lambda"
	// Adjust the import path according to your module name and structure
)

func Handler(ctx context.Context) {
	teamID := "T03FFC1TUPR"

	if err := cbaseSLACK.RefreshBotToken(ctx, teamID); err != nil {
		log.Printf("Error refreshing bot token: %v", err)
	} else {
		log.Println("Bot token refreshed successfully.")
	}
}

func main() {
	lambda.Start(Handler)
}
