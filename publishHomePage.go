package main

import (
	"log"

	"github.com/slack-go/slack"
)

func PublishHomePage(userToken string, userID string){
api := slack.New(userToken);

headerText := slack.NewTextBlockObject("plain_text", "Welcome to Cbase Demo!", false, false)
headerSection := slack.NewHeaderBlock(headerText)

sectionText := slack.NewTextBlockObject("plain_text"," Let's evaluate, shall we.....", false, false)
sectionBlock := slack.NewSectionBlock(sectionText, nil, nil)

imageURL := "https://ctf-images-01.coinbasecdn.net/c5bd0wqjc7v0/6jPp0W7xH2Pe8kwUS79ZSm/85e33ed928fac1e8e58e2d693f5005e0/CB_blog_image.png"
imageAltText := "Coinbase Logo"
imageBlock := slack.NewImageBlock(imageURL, imageAltText, "", nil)

blocks := []slack.Block{headerSection, sectionBlock, imageBlock}

view := slack.HomeTabViewRequest{
	Type: "home",
	Blocks: slack.Blocks{BlockSet: blocks},
}

	res, err := api.PublishView(userID, view, "")
	if err != nil {
		log.Printf("Error publishing home tab: %v", err)
	}

	log.Printf("PublishView() response: %v", res)
}

