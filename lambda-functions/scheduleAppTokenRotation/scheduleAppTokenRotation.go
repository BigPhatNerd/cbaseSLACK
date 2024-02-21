package main

import (
	"context"

	"github.com/BigPhatNerd/cbaseSLACK/oauth" // Use the actual path to your module
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context) {
	// Example: Fetch necessary data like tableName and teamID
	tableName := "Tokens"   // This should ideally come from environment variables or secure storage
	teamID := "T03FFC1TUPR" // Same as above

	// Schedule app token rotation
	oauth.ScheduleAppTokenRotation(tableName, teamID)

}

func main() {
	lambda.Start(Handler)
}
