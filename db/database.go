package db

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Структура элемента листа
type Item struct {
	Name  string `bson:"name"`
	ID    string `bson:"id"`
	State bool   `bson:"state"`
}

// структура листа
type CheckList struct {
	Name      string    `bson:"name"`
	ID        string    `bson:"id"`
	TimeStart time.Time `bson:"time_start"`
	FlagStart bool      `bson:"flag_start"`
	Items     []Item    `bson:"items"`
}

// структура хранения листов
type CheckListJson struct {
	UserName   string      `bson:"user_name"`
	CheckLists []CheckList `bson:"lists"`
}

type DataBase struct {
	session             *mongo.Client
	checkListsTemplates *mongo.Database
	checkLists          *mongo.Database
	mu                  *sync.Mutex
}

func Init() *DataBase {
	mongoUri := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	cltmp := client.Database("heroku_lqwvt43d").Collection("CheckListTemplate")
	cl := client.Database("heroku_lqwvt43d").Collection("CheckList")
	mu := &sync.Mutex{}

	return &DataBase{
		client:              client,
		checkLists:          cl,
		checkListsTemplates: cltmp,
		mu:                  mu,
	}
}

func (db *DataBase) CreateUser(userName string) {
	err := db.checkListsTemplates.Find(bson.M{"user_name": userName}).One(&CheckListJson{})
	if err.Error() == "not found" {
		err = db.checkListsTemplates.Insert(&CheckListJson{
			UserName:   userName,
			CheckLists: nil,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	// user list
	err = db.checkLists.Find(bson.M{"user_name": userName}).One(&CheckListJson{})
	if err.Error() == "not found" {
		err = db.checkLists.Insert(&CheckListJson{
			UserName:   userName,
			CheckLists: nil,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}

func (db *DataBase) DeleteUser(userName string) {
	err := db.checkListsTemplates.Find(bson.M{"user_name": userName}).One(&CheckListJson{})
	if err == nil {
		err = db.checkListsTemplates.Update(bson.M{"user_name": userName}, &CheckListJson{
			UserName:   userName,
			CheckLists: nil,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	// user list
	err = db.checkLists.Find(bson.M{"user_name": userName}).One(&CheckListJson{})
	if err == nil {
		err = db.checkLists.Update(bson.M{"user_name": userName}, &CheckListJson{
			UserName:   userName,
			CheckLists: nil,
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}

func DataBaseasdf() {
	var CL = make(map[string]CheckList)
	for {
		switch locData.Command {
		case TRAddName: // Добавление названия чек листа
			CL[locData.UserName] = CheckList{
				Name: locData.Data,
				ID:   xid.New().String(),
			}

		case TRAddItem: // Добавление пунктов чек листа
			if len(CL[locData.UserName].Items) == 0 {
				CL[locData.UserName] = CheckList{
					Name: CL[locData.UserName].Name,
					ID:   CL[locData.UserName].ID,
					Items: []Item{{
						Name:  locData.Data,
						ID:    xid.New().String(),
						State: false,
					}},
				}
			} else {
				var lItems []Item
				for i := 0; i < len(CL[locData.UserName].Items); i++ {
					lItems = append(lItems, CL[locData.UserName].Items[i])
				}
				lItems = append(lItems, Item{
					Name:  locData.Data,
					ID:    xid.New().String(),
					State: false,
				})
				CL[locData.UserName] = CheckList{
					Name:  CL[locData.UserName].Name,
					ID:    CL[locData.UserName].ID,
					Items: lItems,
				}
			}

		case TREditTemp: // Изменение шаблона
			var temp CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			for i := range temp.CheckLists {
				if temp.CheckLists[i].ID == locData.Data {
					temp.CheckLists[i] = CL[locData.UserName]
					break
				}
			}

			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			delete(CL, locData.UserName)

		case TRSave: // Запись изменений или новых данных в БД
			var temp CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			temp.CheckLists = append(temp.CheckLists, CL[locData.UserName])

			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			delete(CL, locData.UserName)

		case TRDelTemp: // Удаление шаблона
			var temp CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}

			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)

			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{}

		case TRAddFromTemp: // Добавление Листа из шаблона
			var cl1 CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&cl1)
			if err != nil {
				log.Fatal(err)
			}

			var cl2 CheckListJson
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&cl2)
			if err != nil {
				log.Fatal(err)
			}

			var cl3 CheckList
			for i := range cl1.CheckLists {
				if cl1.CheckLists[i].ID == locData.Data {
					cl3 = cl1.CheckLists[i]
				}
			}

			cl2.CheckLists = append(cl2.CheckLists, cl3)
			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &cl2)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{}

		case TRDelList: // Удаление листа
			var temp CheckListJson
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}

			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)
			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{}

		case TRCheckItem: // Отметка пункта листа
			var temp CheckListJson
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}

			for i := range temp.CheckLists {
				for j := range temp.CheckLists[i].Items {
					if locData.Data == temp.CheckLists[i].Items[j].ID {
						temp.CheckLists[i].Items[j].State = !temp.CheckLists[i].Items[j].State
					}
				}
			}

			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{}

		case TRReturnTemp: // возвращение шаблонов пользователя
			var temp CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{
				UserName: locData.UserName,
				DataCL:   temp,
			}

		case TRReturnList: // возвращение чек листов пользователя
			var temp CheckListJson
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{
				UserName: locData.UserName,
				DataCL:   temp,
			}

		}
	}
}

func RemoveCheckList(u []CheckList, listID string) []CheckList {
	var id int
	id = -1
	for i := range u {
		if u[i].ID == listID {
			id = i
			break
		}
	}

	return append(u[:id], u[id+1:]...)
}
