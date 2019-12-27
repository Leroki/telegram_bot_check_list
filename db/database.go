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

// CreateUser create base item in database
func (db *DataBase) CreateUser(userName *string) {
	user := Item{
		UserName:      *userName,
		RootMessageID: 1,
		CheckLists:    nil,
	}
	userIsCreated := db.checkUserInDataBase(userName)
	if !userIsCreated {
		insertOneResult, err := db.checkLists.InsertOne(context.TODO(), user)
		log.Printf("%v :: %v", insertOneResult, err)
	}
}

func (db *DataBase) checkUserInDataBase(userName *string) bool {

	filter := bson.D{{Key: "UserName", Value: *userName}}
	ret := new(Item)
	err := db.checkLists.FindOne(context.TODO(), filter).Decode(ret)
	if err != nil && err.Error() == "mongo: no documents in result" {
		return false
	} else if err == nil && ret != nil {
		return true
	}
	return false
}

// func (db *DataBase) DeleteUser(userName string) {
// 	err := db.checkListsTemplates.Find(bson.M{"user_name": userName}).One(&CheckListJSON{})
// 	if err == nil {
// 		err = db.checkListsTemplates.Update(bson.M{"user_name": userName}, &CheckListJSON{
// 			UserName:   userName,
// 			CheckLists: nil,
// 		})
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	} else {
// 		log.Fatal(err)
// 	}

// 	// user list
// 	err = db.checkLists.Find(bson.M{"user_name": userName}).One(&CheckListJSON{})
// 	if err == nil {
// 		err = db.checkLists.Update(bson.M{"user_name": userName}, &CheckListJSON{
// 			UserName:   userName,
// 			CheckLists: nil,
// 		})
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	} else {
// 		log.Fatal(err)
// 	}
// }

// func DataBaseasdf(CheckListJson) {
// 	var CL = make(map[string]CheckList)
// 	for {
// 		switch locData.Command {
// 		case TRAddName: // Добав CheckListJsonление названия чек листа
// 			CL[locData.UserName] = CheckList{
// 				Name: locData.Data,
// 				ID:   xid.New().String(),
// 			}

// 		case TRAddItem: // Добавление пунктов чек листа
// 			if len(CL[locData.UserName].Items) == 0 {
// 				CL[locData.UserName] = CheckList{
// 					Name: CL[locData.UserName].Name,
// 					ID:   CL[locData.UserName].ID,
// 					Items: []Item{{
// 						Name:  locData.Data,
// 						ID:    xid.New().String(),
// 						State: false,
// 					}},
// 				}
// 			} else {
// 				var lItems []Item
// 				for i := 0; i < len(CL[locData.UserName].Items); i++ {
// 					lItems = append(lItems, CL[locData.UserName].Items[i])
// 				}
// 				lItems = append(lItems, Item{
// 					Name:  locData.Data,
// 					ID:    xid.New().String(),
// 					State: false,
// 					CheckListJson})
// 				CL[locData.UserName] = CheckList{
// 					Name:  CL[locData.UserName].Name,
// 					ID:    CL[locData.UserName].ID,
// 					Items: lItems,
// 				}
// 			}

// 		case TREditTemp: // Изменение шаблона
// 			var temp CheckListJSON
// 			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			for i := range temp.CheckLists {
// 				if temp.CheckLists[i].ID == locData.Data {
// 					temp.CheckLists[i] = CL[locData.UserName]
// 					break
// 				}
// 			}

// 			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			delete(CL, locData.UserName)

// 		case TRSave: // Запись изменений или новых данных в БД
// 			var temp CheckListJSON
// 			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			temp.CheckLists = append(temp.CheckLists, CL[locData.UserName])

// 			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			delete(CL, locData.UserName)

// 		case TRDelTemp: // Удаление шаблона
// 			var temp CheckListJSON
// 			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)

// 			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{}

// 		case TRAddFromTemp: // Добавление Листа из шаблона
// 			var cl1 CheckListJSON
// 			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&cl1)
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			var cl2 CheckListJSON
// 			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&cl2)
// 			if err != nil {
// 				log.Fatal(err)
// 				CheckListJson
// 			}

// 			var cl3 CheckList
// 			for i := range cl1.CheckLists {
// 				if cl1.CheckLists[i].ID == locData.Data {
// 					cl3 = cl1.CheckLists[i]
// 				}
// 			}

// 			cl2.CheckLists = append(cl2.CheckLists, cl3)
// 			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &cl2)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{}

// 		case TRDelList: // Удаление листа
// 			var temp CheckListJSON
// 			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)
// 			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{}

// 		case TRCheckItem: // Отметка пункта листа
// 			var temp CheckListJSON
// 			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}

// 			for i := range temp.CheckLists {
// 				for j := range temp.CheckLists[i].Items {
// 					if locData.Data == temp.CheckLists[i].Items[j].ID {
// 						temp.CheckLists[i].Items[j].State = !temp.CheckLists[i].Items[j].State
// 					}
// 				}
// 			}

// 			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{}

// 		case TRReturnTemp: // возвращение шаблонов пользователя
// 			var temp CheckListJSON
// 			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{
// 				UserName: locData.UserName,
// 				DataCL:   temp,
// 			}

// 		case TRReturnList: // возвращение чек листов пользователя
// 			var temp CheckListJSON
// 			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			TData <- TransactData{
// 				UserName: locData.UserName,
// 				DataCL:   temp,
// 			}

// 		}
// 	}
// }

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
