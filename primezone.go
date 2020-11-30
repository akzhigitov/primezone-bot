package main

import (
	"github.com/PuerkitoBio/goquery"
	resty "github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func parseDeals(config *config) []deal {

	client := resty.New()
	authResult, authError := client.R().Get(config.SiteURL + "auth?token=" + config.AuthToken)

	if authError != nil {
		log.Fatal(authError)
	}

	var deals []deal

	for pageID := 1; pageID < 2; pageID++ {

		page, pageError := client.R().SetCookies(authResult.Cookies()).Get(config.SiteURL + "?page=" + strconv.Itoa(pageID) + "&sort=new")
		if pageError != nil {
			log.Fatal(pageError)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page.String()))
		if err != nil {
			log.Fatal(err)
		}

		doc.Find(".coupon-thumb").Each(func(i int, s *goquery.Selection) {

			title := s.Find(".coupon-title").Text()
			href, hrefExists := s.Attr("href")
			if !hrefExists {
				log.Fatalf("href not exists")
			}
			imgSrc, srcExists := s.Find("img").Attr("src")
			if !srcExists {
				log.Fatalf("img src not exists")
			}
			description := s.Find(".coupon-desciption").Text()

			deals = append(deals, deal{
				title:       title,
				description: description,
				url:         config.SiteURL + href,
				photoURL:    config.SiteURL + imgSrc,
			})
		})
	}

	return deals
}
