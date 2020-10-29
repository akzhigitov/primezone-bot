package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func getMongoCollection(config *config) *mongo.Collection {
	client, err := mongo.NewClient(options.Client().ApplyURI(config.MongoConnectionURI))
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database(config.MongoDatabase)
	return database.Collection("updaters")
}

func saveNewDeals(deals []deal, collection *mongo.Collection) error {
	if len(deals) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var documents []interface{}
	for _, deal := range deals {
		documents = append(documents, bson.D{{"url", deal.url}})
	}
	opts := options.InsertMany().SetOrdered(false)
	_, err := collection.InsertMany(ctx, documents, opts)

	return err
}

func filter(deals []deal, collection *mongo.Collection) []deal {
	var result []deal
	for _, deal := range deals {
		if !isExists(collection, deal) {
			result = append(result, deal)
		}
	}
	if len(result) == 0{
		log.Infoln("No new deals")
	}
	return result
}

func isExists(collection *mongo.Collection, deal deal) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"url": bson.M{"$eq": deal.url}}
	single := collection.FindOne(ctx, filter)
	return single.Err() != mongo.ErrNoDocuments
}
