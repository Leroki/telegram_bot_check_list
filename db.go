package main

import (
	"log"
	"os"

	"github.com/rs/xid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func DataBase(TData chan TransactData) {
	mongoUri := os.Getenv("MONGODB_URI")
	session, err := mgo.Dial(mongoUri)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	checkListTempDB := session.DB("heroku_lqwvt43d").C("CheckListTemplate")
	checkListDB := session.DB("heroku_lqwvt43d").C("CheckList")

	var CL = make(map[string]CheckList)
	for {
		locData := <-TData
		switch locData.Command {
		case TRInitUser: // Создание пользователя в БД
			// user check lists
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&CheckListJson{})
			if err == nil {
				err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &CheckListJson{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else if err.Error() == "not found" {
				err = checkListTempDB.Insert(&CheckListJson{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}

			// user list
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&CheckListJson{})
			if err == nil {
				err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &CheckListJson{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else if err.Error() == "not found" {
				err = checkListDB.Insert(&CheckListJson{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}

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

		case TRReturnTemp:
			var temp CheckListJson
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- TransactData{
				UserName: locData.UserName,
				DataCL:   temp,
			}

		case TRReturnList:
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
