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
}

// Init create connecting to database
func Init() *DataBase {
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
	}
}

// CreateUser create document in database
func (db *DataBase) CreateUser(userName *string) {
	user := Item{
		UserName:      *userName,
		RootMessageID: 1,
		CheckLists:    nil,
	}
	userIsCreated := db.checkUserInDataBase(userName)
	if !userIsCreated {
		insertOneResult, err := db.checkLists.InsertOne(context.TODO(), user)
		log.Printf(": fn -> db.CreateUser : %v :: %v", insertOneResult, err)
	}
}

func (db *DataBase) checkUserInDataBase(userName *string) bool {

	filter := bson.D{{Key: "UserName", Value: *userName}}
	ret := new(Item)
	err := db.checkLists.FindOne(context.TODO(), filter).Decode(ret)
	if err != nil && err.Error() == "mongo: no documents in result" {
		return false
	} else if err == nil {
		return true
	}
	return false
}

// DeleteUser delete basic document in collection
func (db *DataBase) DeleteUser(userName *string) {
	filter := bson.D{{Key: "UserName", Value: userName}}
	delRes, err := db.checkLists.DeleteOne(context.TODO(), filter)
	log.Printf(": fn -> db.DeleteUser : %v :: %v", delRes, err)
}
// // RemoveCheckList remove check list from runtime
// func RemoveCheckList(u []CheckList, listID string) []CheckList {
// 	var id int
// 	id = -1
// 	for i := range u {
// 		if u[i].ID == listID {
// 			id = i
// 			break
// 		}
// 	}

// 	return append(u[:id], u[id+1:]...)
// }
