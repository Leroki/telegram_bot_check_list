package main

import (
	"encoding/json"
	"github.com/rs/xid"
	"io/ioutil"
	"log"
)

func DataBase(TData chan TransactData) {
	var CL = make(map[string]CheckList)
	for {
		locData := <-TData
		switch locData.Command {
		case TRAddName:
			CL[locData.UserName] = CheckList{
				Name: locData.Data,
				ID:   xid.New().String(),
			}

		case TRAddItem:
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

		case TREditTemp:
			filePath := "AppData/" + locData.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var temp CheckListJson
			err = json.Unmarshal(rawDataIn, &temp)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			for i := range temp.CheckLists {
				if temp.CheckLists[i].ID == locData.Data {
					temp.CheckLists[i] = CL[locData.UserName]
					break
				}
			}

			rawDataOut, err := json.MarshalIndent(&temp, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			delete(CL, locData.UserName)

		case TRSave:
			filePath := "AppData/" + locData.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var temp CheckListJson
			if len(rawDataIn) == 0 {
				temp.CheckLists = append(temp.CheckLists, CL[locData.UserName])
			} else {
				err = json.Unmarshal(rawDataIn, &temp)
				if err != nil {
					log.Fatal("Invalid settings format:", err)
				}
				temp.CheckLists = append(temp.CheckLists, CL[locData.UserName])
			}

			rawDataOut, err := json.MarshalIndent(&temp, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			delete(CL, locData.UserName)

		case TRDelTemp:
			filePath := "AppData/" + locData.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var temp CheckListJson
			err = json.Unmarshal(rawDataIn, &temp)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)
			rawDataOut, err := json.MarshalIndent(&temp, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}

		case TRAddFromTemp:
			file1 := "AppData/" + locData.UserName + ".tem.json"
			file2 := "AppData/" + locData.UserName + ".json"

			rawDataIn1, err := ioutil.ReadFile(file1)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}
			rawDataIn2, err := ioutil.ReadFile(file2)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var cl1 CheckListJson
			err = json.Unmarshal(rawDataIn1, &cl1)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			var cl2 CheckListJson
			err = json.Unmarshal(rawDataIn2, &cl2)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}

			var cl3 CheckList
			for i := range cl1.CheckLists {
				if cl1.CheckLists[i].ID == locData.Data {
					cl3 = cl1.CheckLists[i]
				}
			}
			cl2.CheckLists = append(cl2.CheckLists, cl3)
			rawDataOut, err := json.MarshalIndent(&cl2, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(file2, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			TData <- TransactData{}
		case TRDelList:
			filePath := "AppData/" + locData.UserName + ".json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var temp CheckListJson
			err = json.Unmarshal(rawDataIn, &temp)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			temp.CheckLists = RemoveCheckList(temp.CheckLists, locData.Data)
			rawDataOut, err := json.MarshalIndent(&temp, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			TData <- TransactData{}
		case TRCheckItem:
			filePath := "AppData/" + locData.UserName + ".json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var temp CheckListJson
			err = json.Unmarshal(rawDataIn, &temp)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}

			count := 0
			for i := range temp.CheckLists {
				for j := range temp.CheckLists[i].Items {
					if temp.CheckLists[i].Items[j].State == true {
						count++
					}
					if locData.Data == temp.CheckLists[i].Items[j].ID {
						temp.CheckLists[i].Items[j].State = true
					}
				}
			}
			rawDataOut, err := json.MarshalIndent(&temp, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			TData <- TransactData{}
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
