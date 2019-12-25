package main

import (
	"log"
	"os"

	"github.com/rs/xid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func dataBase(TData chan transactData) {
	mongoURI := os.Getenv("MONGODB_URI")
	session, err := mgo.Dial(mongoURI)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	checkListTempDB := session.DB("heroku_lqwvt43d").C("checkListTemplate")
	checkListDB := session.DB("heroku_lqwvt43d").C("checkList")

	var CL = make(map[string]checkList)
	for {
		locData := <-TData
		switch locData.Command {
		case trInitUser: // Создание пользователя в БД
			// user check lists
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&checkListJSON{})
			if err == nil {
				err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &checkListJSON{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else if err.Error() == "not found" {
				err = checkListTempDB.Insert(&checkListJSON{
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
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&checkListJSON{})
			if err == nil {
				err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &checkListJSON{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else if err.Error() == "not found" {
				err = checkListDB.Insert(&checkListJSON{
					UserName:   locData.UserName,
					CheckLists: nil,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}

		case trAddName: // Добавление названия чек листа
			CL[locData.UserName] = checkList{
				Name: locData.Data,
				ID:   xid.New().String(),
			}

		case trAddItem: // Добавление пунктов чек листа
			if len(CL[locData.UserName].Items) == 0 {
				CL[locData.UserName] = checkList{
					Name: CL[locData.UserName].Name,
					ID:   CL[locData.UserName].ID,
					Items: []item{{
						Name:  locData.Data,
						ID:    xid.New().String(),
						State: false,
					}},
				}
			} else {
				var lItems []item
				for i := 0; i < len(CL[locData.UserName].Items); i++ {
					lItems = append(lItems, CL[locData.UserName].Items[i])
				}
				lItems = append(lItems, item{
					Name:  locData.Data,
					ID:    xid.New().String(),
					State: false,
				})
				CL[locData.UserName] = checkList{
					Name:  CL[locData.UserName].Name,
					ID:    CL[locData.UserName].ID,
					Items: lItems,
				}
			}

		case trEditTemp: // Изменение шаблона
			var temp checkListJSON
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

		case trSave: // Запись изменений или новых данных в БД
			var temp checkListJSON
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

		case trDelTemp: // Удаление шаблона
			var temp checkListJSON
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}

			temp.CheckLists = removecheckList(temp.CheckLists, locData.Data)

			err = checkListTempDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- transactData{}

		case trAddFromTemp: // Добавление Листа из шаблона
			var cl1 checkListJSON
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&cl1)
			if err != nil {
				log.Fatal(err)
			}

			var cl2 checkListJSON
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&cl2)
			if err != nil {
				log.Fatal(err)
			}

			var cl3 checkList
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
			TData <- transactData{}

		case trDelList: // Удаление листа
			var temp checkListJSON
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}

			temp.CheckLists = removecheckList(temp.CheckLists, locData.Data)
			err = checkListDB.Update(bson.M{"user_name": locData.UserName}, &temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- transactData{}

		case trCheckItem: // Отметка пункта листа
			var temp checkListJSON
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
			TData <- transactData{}

		case trReturnTemp:
			var temp checkListJSON
			err = checkListTempDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- transactData{
				UserName: locData.UserName,
				DataCL:   temp,
			}

		case trReturnList:
			var temp checkListJSON
			err = checkListDB.Find(bson.M{"user_name": locData.UserName}).One(&temp)
			if err != nil {
				log.Fatal(err)
			}
			TData <- transactData{
				UserName: locData.UserName,
				DataCL:   temp,
			}

		}
	}
}

func removecheckList(u []checkList, listID string) []checkList {
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
