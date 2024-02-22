package main

import (
	"context"

	"github.com/BigPhatNerd/cbaseSLACK/oauth" // Use the actual path to your module
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context) {

	tableName := "Tokens"
	teamID := "T03FFC1TUPR"

	oauth.ScheduleAppTokenRotation(tableName, teamID)

}

func main() {
	lambda.Start(Handler)
}
