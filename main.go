package main

import (
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
)

type deal struct {
	title       string
	url         string
	photoURL    string
	description string
}

type photo struct {
	ChatID                string `json:"chat_id"`
	Photo                 string `json:"photo"`
	Caption               string `json:"caption"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type config struct {
	MongoConnectionURI string
	MongoDatabase      string
	SiteURL            string
	BotToken           string
	ChatID             string
	AuthToken          string
	Schedule           string
}

type primeZoneReminder struct {
	config *config
}

func (e primeZoneReminder) Run() {
	collection := getMongoCollection(e.config)
	allDeals := parseDeals(e.config)
	newDeals := filter(allDeals, collection)
	sendToTelegram(newDeals, e.config)
	err := saveNewDeals(newDeals, collection)
	if err != nil {
		log.Fatal(err)
	}
}

func jobJSON(c *gin.Context) {
	c.JSON(200, jobrunner.StatusJson())
}

func jobHTML(c *gin.Context) {
	c.HTML(200, "Status.html", jobrunner.StatusPage())
}

func main() {
	routes := gin.Default()

	routes.GET("/jobrunner/json", jobJSON)

	routes.LoadHTMLGlob("views/Status.html")

	routes.GET("/jobrunner/html", jobHTML)

	config := readConfig()

	jobrunner.Start()
	err := jobrunner.Schedule(config.Schedule, primeZoneReminder{config})
	if err != nil {
		log.Fatal(err)
	}

	routes.Run(":8080")
}

func readConfig() *config {
	mongoConnectionURL, found := os.LookupEnv("MONGO_CONNECTION_URI")
	if !found {
		log.Fatal("MONGO_CONNECTION_URI env variable not found")
	}

	mongoDatabase, found := os.LookupEnv("MONGO_DATABASE")
	if !found {
		log.Fatal("MONGO_DATABASE env variable not found")
	}

	siteURL, found := os.LookupEnv("SITE_URL")
	if !found {
		log.Fatal("SITE_URL env variable not found")
	}

	botToken, found := os.LookupEnv("TELEGRAM_BOT_TOKEN")
	if !found {
		log.Fatal("TELEGRAM_BOT_TOKEN env variable not found")
	}

	chatID, found := os.LookupEnv("TELEGRAM_CHAT_ID")
	if !found {
		log.Fatal("TELEGRAM_CHAT_ID env variable not found")
	}

	authToken, found := os.LookupEnv("AUTH_TOKEN")
	if !found {
		log.Fatal("AUTH_TOKEN env variable not found")
	}

	schedule, found := os.LookupEnv("SCHEDULE")
	if !found {
		log.Fatal("SCHEDULE env variable not found")
	}

	return &config{
		MongoConnectionURI: mongoConnectionURL,
		MongoDatabase:      mongoDatabase,
		SiteURL:            siteURL,
		BotToken:           botToken,
		ChatID:             chatID,
		AuthToken:          authToken,
		Schedule:           schedule,
	}
}
