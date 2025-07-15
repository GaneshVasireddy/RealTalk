package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/GaneshVasireddy/RealTalk/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Connect establishes a connection to the MongoDB server using the provided configuration.
// It returns a pointer to the mongo.Client instance.
// If the connection fails, it prints an error message and returns nil.
func Connect(dbConfig *config.Mongo) *mongo.Client {

	// Connect to the MongoDB server
	client, err := mongo.Connect(options.Client().ApplyURI(dbConfig.ConnectionString))
	if err != nil {
		fmt.Println("Error connecting to MongoDB:", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Check the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println("Error pinging MongoDB:", err)
		return nil
	}

	fmt.Println("Connected to MongoDB!")
	return client
}
