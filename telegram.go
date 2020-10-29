package main

import (
	"fmt"
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

func sendToTelegram(deals []deal, config *config) {

	client := resty.New()
	for _, deal := range deals {

		p := photo{
			ChatID:                config.ChatID,
			ParseMode:             "MarkdownV2",
			DisableWebPagePreview: false,
			Caption:               fmt.Sprintf("[%s](%s)\n%s", deal.title, deal.url, deal.description),
			Photo:                 deal.photoURL,
		}

		post, err := client.R().SetHeader("Content-Type", "application/json").SetBody(p).
			Post(fmt.Sprintf("https://api.telegram.org/bot%v/sendPhoto", config.BotToken))

		if err != nil {
			log.Fatal(err)
		}

		log.Infoln(post)
	}
}