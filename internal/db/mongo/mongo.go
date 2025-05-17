package mongo

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
}

var Mongo *MongoDB

func InitMongoDB() {
	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		fmt.Println("MONGO_URI env var not set")
		os.Exit(1)
	}
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoUri))
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		os.Exit(1)
	}
	Mongo = &MongoDB{
		Client: mongoClient,
	}
	fmt.Println("Connected to MongoDB")
}

func InsertJson(dbName string, collectionName string, data interface{}) (*mongo.InsertOneResult, error) {
	/*
	  bsonData, err := JsonToBson(data)
	  if err != nil {
	    return nil, fmt.Errorf("Error converting JSON to BSON: %v", err)
	  }
	*/
	collection := Mongo.Client.Database(dbName).Collection(collectionName)
	if collection == nil {
		return nil, fmt.Errorf("Collection not found: %s %s", dbName, collectionName)
	}
	res, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		return nil, fmt.Errorf("Error inserting document: %v", err)
	}
	return res, nil
}

func JsonToBson(data interface{}) (interface{}, error) {
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Error converting JSON to BSON: %v", err)
	}
	return bsonData, nil
}
