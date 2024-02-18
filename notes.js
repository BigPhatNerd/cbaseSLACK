logger.WithFields(logrus.Fields{
    "statusCode": statusCode,
    "body":       body,
}).Info("Received HTTP response")


 userID := 42
    log.Printf("User ID: %d", userID)

    Table Name:
    Token
    SlackUsers



    

    xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjM0MzE1OTAxNzI5LThiOGZjYWM0OGI2ZGYxNTljYmI4YTdlOTU1YWRjNjBlYjcwODc0OTZlM2ZmN2NjNjVkZGMyOTA4ZTJmNjIyMTQ

    MAKE SURE it doesnt' go back to this:
    xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjIxMzAwOTcyNzIyLTRjYjBlYmMyNzAyZGU0NjczZmJhM2M0OTU3NTg3OWYxMjU1MDIwNTQ0ZTljMTA2YzQzNTExYjQ5ODg5OWJiYjM

    This is what it should be:
    "token":"xoxe.xoxp-1-Mi0yLTM1MjU0MDk5NDY4MDctMzU1MjYyNTExMDU2MS02NjMyMjE1NzUzNTg1LTY2MDcxMTY3MTMwNjMtNGM3MjAwMjJkMmYyYTc4MDYxMjViYmEzMmNmZTA1MzI3MmU4ODhiODc5NTAxYjkyNGNlMzk0OGVhY2ZiYzlkZA"

    "refresh_token":"xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjI0MTM0NzA5NDEyLTUxZWEwNDkwNmFmYTkzY2FkNzQ1MzQxNGEwZGQ5YTIxMjQ5YzBmYTE0NzVlZDBhZjQ2NTlkZjI4ZThiMTM4OTI",

{
    "Items": [
        {
            "Token": {
                "S": "xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjIxMzAwOTcyNzIyLTRjYjBlYmMyNzAyZGU0NjczZmJhM2M0OTU3NTg3OWYxMjU1MDIwNTQ0ZTljMTA2YzQzNTExYjQ5ODg5OWJiYjM"
            },
            "ExpiresAt": {
                "N": "1707742369"
            },
            "CurrentToken": {
                "S": ""
            },
            "Key": {
                "S": "SlackRefreshToken"
            },
            "RefreshToken": {
                "S": "xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjI0MTM0NzA5NDEyLTUxZWEwNDkwNmFmYTkzY2FkNzQ1MzQxNGEwZGQ5YTIxMjQ5YzBmYTE0NzVlZDBhZjQ2NTlkZjI4ZThiMTM4OTI"
            }
        }
        
    ],
    "Count": 1,
    "ScannedCount": 1,
    "ConsumedCapacity": null
}


Stop datadog agent:
launchctl stop com.datadoghq.agent

Start datadog agent:
launchctl start com.datadoghq.agent

Configuration Source: file:/opt/datadog-agent/etc/conf.d/uptime.d/conf.yaml.default

Datadog Agent Manager
http://127.0.0.1:5002/

Check for errors in datadog agent:
open /opt/datadog-agent/logs/launchd.log