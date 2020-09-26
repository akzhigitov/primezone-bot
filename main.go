package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty"
)

type Deal struct {
	title    string
	url      string
	photoUrl string
}

type Photo struct {
	ChatId                string `json:"chat_id"`
	Photo                 string `json:"photo"`
	Caption               string `json:"caption"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type record struct {
	Name string `bson:"name"`
	Code string `bson:"code"`
}

type Config struct {
	MongoConnectionUri string
	MongoDatabase      string
	SiteUrl            string
	BotToken           string
	ChatId             string
}

const UpdatedItemName = "primezone"

func main() {
	config := readConfig()
	collection := getMongoCollection(config)
	allDeals := parseDeals(config.SiteUrl)
	newDeals := filter(allDeals, collection)
	sendToTelegram(newDeals, config)
	saveLastDealUrl(newDeals, collection)
}

func readConfig() *Config {
	mongoConnectionUri, found := os.LookupEnv("MONGO_CONNECTION_URI")
	if !found {
		log.Fatal("MONGO_CONNECTION_URI env variable not found")
	}

	mongoDatabase, found := os.LookupEnv("MONGO_DATABASE")
	if !found {
		log.Fatal("MONGO_DATABASE env variable not found")
	}

	siteUrl, found := os.LookupEnv("SITE_URL")
	if !found {
		log.Fatal("SITE_URL env variable not found")
	}

	botToken, found := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !found {
		log.Fatal("TELEGRAM_BOT_TOKEN env variable not found")
	}

	chatId, found := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !found {
		log.Fatal("TELEGRAM_CHAT_ID env variable not found")
	}

	return &Config{
		MongoConnectionUri: mongoConnectionUri,
		MongoDatabase:      mongoDatabase,
		SiteUrl:            siteUrl,
		BotToken:           botToken,
		ChatId:             chatId,
	}
}

func getMongoCollection(config *Config) *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoConnectionUri))
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database(config.MongoDatabase)
	return database.Collection("updaters")
}

func parseDeals(siteUrl string) []Deal {
	res, err := http.Get(siteUrl + "/?sort=new")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var deals []Deal

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

		deals = append(deals, Deal{
			title:    title,
			url:      siteUrl + href,
			photoUrl: siteUrl + imgSrc,
		})
	})

	return deals
}

func sendToTelegram(deals []Deal, config *Config) {

	client := resty.New()
	for _, deal := range deals {

		p := Photo{
			ChatId:                config.ChatId,
			ParseMode:             "MarkdownV2",
			DisableWebPagePreview: false,
			Caption:               fmt.Sprintf("[%s](%s)", deal.title, deal.url),
			Photo:                 deal.photoUrl,
		}

		post, err := client.R().SetHeader("Content-Type", "application/json").SetBody(p).
			Post(fmt.Sprintf("https://api.telegram.org/bot%v/sendPhoto", config.BotToken))

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%+v\n", post)
	}
}

func saveLastDealUrl(deals []Deal, collection *mongo.Collection) {
	if len(deals) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	latestDeal := deals[0]

	opts := options.FindOneAndReplace().SetUpsert(true)
	filter := bson.M{"name": bson.M{"$eq": UpdatedItemName}}
	replacement := bson.M{
		"name": UpdatedItemName,
		"code": latestDeal.url,
	}

	collection.FindOneAndReplace(ctx, filter, replacement, opts)
}

func filter(deals []Deal, collection *mongo.Collection) []Deal {
	lastDealUrl := getLastDealUrl(collection)
	var result []Deal
	for _, deal := range deals {
		if deal.url == lastDealUrl {
			break
		}
		result = append(result, deal)
	}

	return result
}

func getLastDealUrl(collection *mongo.Collection) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"name": bson.M{"$eq": UpdatedItemName}}
	result := &record{}
	single := collection.FindOne(ctx, filter)
	err := single.Decode(result)

	if err != nil {
		log.Fatal(err)
	}
	return result.Code
}
