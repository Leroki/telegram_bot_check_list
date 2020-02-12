package db

import (
	"context"
	"log"
	"os"
	"sync"

	// "github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DataBase aa
type DataBase struct {
	client     *mongo.Client
	checkLists *mongo.Collection
	dbMutex    *sync.Mutex
	ctx        *context.Context
}

// Init create connecting to database
func Init(ctx *context.Context) *DataBase {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("no find os env value MONGODB_URI")
	}
	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetRetryWrites(false)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	checkListCollection := client.Database("heroku_lqwvt43d").Collection("test")
	mu := new(sync.Mutex)

	return &DataBase{
		client:     client,
		checkLists: checkListCollection,
		dbMutex:    mu,
		ctx:        ctx,
	}
}

// CreateUser create document in database
func (db *DataBase) CreateUser(userName string) {
	user := User{
		UserName:      userName,
		RootMessageID: 1,
		CheckLists:    nil,
	}
	userIsCreated := db.checkUserInDataBase(userName)
	if !userIsCreated {
		ctx := *db.ctx
		insertOneResult, err := db.checkLists.InsertOne(ctx, user)
		log.Printf(": fn -> db.CreateUser : %v :: %v", insertOneResult, err)
	}
}

func (db *DataBase) checkUserInDataBase(userName string) bool {
	filter := bson.D{{Key: "UserName", Value: userName}}
	ret := new(User)
	ctx := *db.ctx
	err := db.checkLists.FindOne(ctx, filter).Decode(ret)
	if err != nil && err.Error() == "mongo: no documents in result" {
		return false
	} else if err == nil {
		return true
	} else {
		panic(err)
	}
}

// DeleteUser delete basic document in collection
func (db *DataBase) DeleteUser(userName string) {
	filter := bson.D{{Key: "UserName", Value: userName}}
	ctx := *db.ctx
	delRes, err := db.checkLists.DeleteOne(ctx, filter)
	log.Printf(": fn -> db.DeleteUser : %v :: %v", delRes, err)
}

// UpdateUser update base document
func (db *DataBase) UpdateUser(userName string) {
	filter := bson.D{{Key: "UserName", Value: userName}}
	ctx := *db.ctx
	db.checkLists.UpdateOne(ctx, filter, bson.D{{}})
}
