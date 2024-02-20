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



 Raw OAuth Response: {"ok":true,
 "app_id":"A06J7D5SJGK",
 "authed_user":{"id":"U03G8JD38GH"},
 "scope":"app_mentions:read,chat:write.customize,chat:write,chat:write.public,incoming-webhook",
 "token_type":"bot",
 "access_token":"xoxe.xoxb-1-MS0yLTM1MjU0MDk5NDY4MDctNjYzMjU0MzkyNTU2OS02NjA1MzI2ODM5ODMxLTY2ODIzOTcwMTUxMzYtNjU3MTQ4NTkzZjU0NGZlMjQ0YmUwMTdkYTU1ZDNjZWU2MDM2ZDk2YzUwZGE4YzEwY2YwNDA3MTU2NDg4ZTZiZg",
 "bot_user_id":"U06JLFZT7GR",
 "refresh_token":"xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYwNTMyNjgzOTgzMS02NjQyNjcyNzUxNjY0LTdkMjY3YjhkZGNhYzMwZGRjNWQ5MTBmMTc1YzMzNzM4MjQzNGIyMzM2ZTk2N2I1NjJhOGYwNTdkODVmM2QxMDg",
 "expires_in":31892,"team":{"id":"T03FFC1TUPR","name":"wilson gmail test"},"enterprise":null,"is_enterprise_install":false,"incoming_webhook":{"channel":"#general","channel_id":"C03FYAC861J","configuration_url":"https:\/\/wilsongmailtest.slack.com\/services\/B06KDRM3QN6","url":"https:\/\/hooks.slack.com\/services\/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"}} 

 xoxe-1-My0xLTM1MjU0MDk5NDY4MDctNjYzMjIxNTc1MzU4NS02NjU4NTQwOTI1NjUxLTUwYWI2OGRjMzkyNTc3YzU0NmNlNjQ2MGQ2YjU3NjkwY2JmYTY2Njg5NzAyZjdhZTkwNDg4NjdmOWFjNmUzM2Y

Bot Tokens and Refresh Tokens:
 https://api.slack.com/apps/A06J7D5SJGK/oauth

 https://ctf-images-01.coinbasecdn.net/c5bd0wqjc7v0/6jPp0W7xH2Pe8kwUS79ZSm/85e33ed928fac1e8e58e2d693f5005e0/CB_blog_image.png