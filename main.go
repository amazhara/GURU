package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Global cache to store users
	users = make(map[uint64]*User)
	// Statistic for each user by ID
	// If the user id exists -
	// stats must exist too
	userStats = make(map[uint64]*UserStat)
	// All deposits users made
	deposits = make(map[uint64]Deposit)
	// All transactions users made
	transactions = make(map[uint64]Transaction)
	// Updated users
	usersUpdate = make(map[uint64]*User)
	// Collection connection
	collection *mongo.Collection
)

func main() {

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				updateCollection()
			}
		}
	}()

	// Create collection
	collection = client.Database("test").Collection("users")

	http.Handle("/user/create", postValidate(
		http.HandlerFunc(userCreateHandler)))

	http.Handle("/user/get", postValidate(
		http.HandlerFunc(getUserHandler)))

	http.Handle("/user/deposit", postValidate(
		http.HandlerFunc(addDepositHandler)))

	http.Handle("/transaction", postValidate(
		http.HandlerFunc(transactionHandler)))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
