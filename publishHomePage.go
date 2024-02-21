package main

import (
	"fmt"
	"log"

	"github.com/slack-go/slack"
)

func PublishHomePage(userID string, reviews []Review) {

	headerText := slack.NewTextBlockObject("plain_text", "Welcome to Cbase Demo!", false, false)
	headerSection := slack.NewHeaderBlock(headerText)

	sectionText := slack.NewTextBlockObject("plain_text", " Let's evaluate, shall we.....", false, false)
	sectionBlock := slack.NewSectionBlock(sectionText, nil, nil)

	imageURL := "https://ctf-images-01.coinbasecdn.net/c5bd0wqjc7v0/6jPp0W7xH2Pe8kwUS79ZSm/85e33ed928fac1e8e58e2d693f5005e0/CB_blog_image.png"
	imageAltText := "Coinbase Logo"
	imageBlock := slack.NewImageBlock(imageURL, imageAltText, "", nil)

	divider := slack.NewDividerBlock()

	blocks := []slack.Block{headerSection, sectionBlock, imageBlock, divider}

	for _, review := range reviews {
		reviewBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Reviewer:* %s\n*Employee Reviewed:* %s\n*Feedback:* %s", review.UserName, review.EmployeeSelected, review.Feedback), false, false),
			nil, nil,
		)
		blocks = append(blocks, reviewBlock, slack.NewDividerBlock())
	}

	if len(reviews) > 0 {
		removeButton := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "Remove reviews from the interface", false, false),
			nil,
			slack.NewAccessory(slack.NewButtonBlockElement("remove_reviews_action", "remove_value", slack.NewTextBlockObject("plain_text", "Remove Reviews", true, false))),
		)
		blocks = append(blocks, removeButton, slack.NewDividerBlock())
	}

	if len(reviews) == 0 {
		createButton := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "Create a review of your co-worker", false, false),
			nil,
			slack.NewAccessory(slack.NewButtonBlockElement("create_action", "create_value", slack.NewTextBlockObject("plain_text", "Create", true, false))),
		)
		viewButton := slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", "View submitted reviews", false, false),
			nil,
			slack.NewAccessory(slack.NewButtonBlockElement("view_action", "view_value", slack.NewTextBlockObject("plain_text", "View Reviews", true, false))),
		)

		blocks = append(blocks, createButton, divider, viewButton)
	}

	view := slack.HomeTabViewRequest{
		Type:   "home",
		Blocks: slack.Blocks{BlockSet: blocks},
	}

	res, err := api.PublishView(userID, view, "")
	if err != nil {
		log.Printf("Error publishing home tab: %v", err)
	}

	log.Printf("PublishView() response: %v", res)
}
