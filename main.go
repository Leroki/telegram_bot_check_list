package main

import (
	"encoding/json"
	"github.com/rs/xid"
	tg "gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"os"

	/* "strings" */ /* "strconv" */)

type Config struct {
	TelegramBotToken string
}

type Item struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State bool   `json:"state"`
}

type CheckList struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Items []Item `json:"items"`
}

type CheckListJson struct {
	CheckLists []CheckList `json:"lists"`
}

type User struct {
	Name  string
	ID    int64
	State string
	Data  string
}

type TransactData struct {
	UserName string
	Data     string
	Command  string
}

type CallbackData struct {
	ListID  string `json:"list_id"`
	Command string `json:"command"`
}

func main() {
	os.Mkdir("AppData", os.ModePerm)
	file, _ := os.Open("configs/config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	file.Close()

	bot, err := tg.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// for long pooling
	/*u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)*/

	// http
	updates := bot.ListenForWebhook("/" + bot.Token)

	var Users map[string]User = make(map[string]User)

	var TData chan TransactData = make(chan TransactData)
	go DataBase(TData)

	// главный цикл
	for update := range updates {
		// обработка callback'ов
		if update.CallbackQuery != nil {
			UserName := update.CallbackQuery.From.UserName
			UserID := int64(update.CallbackQuery.From.ID)
			query := update.CallbackQuery
			var cbData CallbackData
			err := json.Unmarshal([]byte(query.Data), &cbData)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			switch cbData.Command {
			case "show templ":
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Show templ",
					Data:  cbData.ListID,
				}
				go ShowList(cbData.ListID, Users[UserName], bot)
			case "add to list":
				TData <- TransactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  "add from templ",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Lists",
				}
				<-TData
				ShowCheckList(Users[UserName], bot)

			case "check list":
				Users[UserName] = User{
					Name:  UserName,
					Data:  cbData.ListID,
					ID:    UserID,
					State: "Delete from list",
				}
				keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Удалить"))
				keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Назад"))
				keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
				msg := tg.NewMessage(Users[UserName].ID, "Вы в главном меню")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			case "check item":
				TData <- TransactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  "check item",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					Data:  cbData.ListID,
					State: "Lists",
				}
				<-TData
				ShowCheckList(Users[UserName], bot)
			}
		}

		if update.Message == nil {
			continue
		}

		UserName := update.Message.From.UserName
		UserID := int64(update.Message.From.ID)

		if Users[UserName].ID == 0 {
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Main",
			}
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		// debug info

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			msg := tg.NewMessage(UserID, "Привет "+update.Message.From.FirstName+"! Я телеграм бот.")
			bot.Send(msg)
			InitUser(UserName)
		case "stop":
			msg := tg.NewMessage(UserID, "Пока "+update.Message.From.FirstName+"!")
			bot.Send(msg)
			delete(Users, UserName)
		}

		// обработка кнопок
		switch update.Message.Text {
		case "В главное меню":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Main",
			}
		case "Листы":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Lists",
			}
			ShowCheckList(Users[UserName], bot)
		case "Шаблоны":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Templates",
			}
		case "Показать мои шаблоны":
			go ShowTemplates(Users[UserName], bot, "show")
			continue
		case "Добавить новый шаблон":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Add tmp",
			}
		case "Отмена":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Templates",
			}
		case "Назад":
			if Users[UserName].State == "Show templ" {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Templates",
				}
			} else if Users[UserName].State == "Delete from list" {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Lists",
				}
			}
		case "Завершить":
			if Users[UserName].State == "Add tmp item" {
				TData <- TransactData{
					UserName: UserName,
					Data:     "",
					Command:  "save tmp data",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Templates",
				}
			} else if Users[UserName].State == "Edit tmp item" {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  "edit templ",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Templates",
				}
			}
		case "Изменить":
			if Users[UserName].State == "Show templ" {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Edit tmp",
					Data:  Users[UserName].Data,
				}
			}
		case "Удалить":
			if Users[UserName].State == "Show templ" {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  "delete templ",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Templates",
				}
			} else if Users[UserName].State == "Delete from list" {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  "delete List",
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: "Lists",
				}
				<-TData
				ShowCheckList(Users[UserName], bot)
			}
		case "Добавить новый лист из шаблона":
			ShowTemplates(Users[UserName], bot, "add")
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Add from template",
			}
		}

		// обработка положения в меню
		switch Users[update.Message.From.UserName].State {
		case "Main":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Листы"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Шаблоны"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в главном меню")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Lists":
			// ShowCheckList(Users[UserName], bot)
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Добавить новый лист из шаблона"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Что сделать?")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Templates":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Показать мои шаблоны"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Добавить новый шаблон"))
			keyRow3 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в меню шаблонов")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Add tmp name":
			TData <- TransactData{
				UserName: update.Message.From.UserName,
				Data:     update.Message.Text,
				Command:  "add name",
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Завершить"))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для формирования листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Add tmp item",
			}

		case "Add tmp":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Add tmp name",
			}
		case "Add tmp item":
			TData <- TransactData{
				UserName: update.Message.From.UserName,
				Data:     update.Message.Text,
				Command:  "add item",
			}
		case "Edit tmp name":
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  "add name",
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Завершить"))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для изменения листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Edit tmp item",
				Data:  Users[UserName].Data,
			}
		case "Edit tmp":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: "Edit tmp name",
				Data:  Users[UserName].Data,
			}
		case "Edit tmp item":
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  "add item",
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

func ShowList(ListID string, user User, bot *tg.BotAPI) {
	filePath := "AppData/" + user.Name + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var templ CheckListJson
	err = json.Unmarshal(rawDataIn, &templ)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}

	reply := ""
	for i := range templ.CheckLists {
		if templ.CheckLists[i].ID == ListID {
			for y := range templ.CheckLists[i].Items {
				reply += templ.CheckLists[i].Items[y].Name + "\n"
			}
			break
		}
	}
	if reply == "" {
		reply = "Этот шаблон пуст"
	}
	keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Изменить"))
	keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Удалить"))
	keyRow3 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Назад"))
	keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
	msg := tg.NewMessage(user.ID, reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ShowCheckList(user User, bot *tg.BotAPI) {
	filePath := "AppData/" + user.Name + ".json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var templ CheckListJson
	err = json.Unmarshal(rawDataIn, &templ)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}
	reply := "Ваши листы"
	var keys [][]tg.InlineKeyboardButton
	var count int = 0
	for i := range templ.CheckLists {
		lName := templ.CheckLists[i].Name
		lID := templ.CheckLists[i].ID
		var cbData CallbackData = CallbackData{
			ListID:  lID,
			Command: "check list",
		}
		outData, _ := json.Marshal(&cbData)
		keys = append(keys, []tg.InlineKeyboardButton{})
		keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData("⏬   "+lName+"   ⏬", string(outData)))
		count++
		for j := range templ.CheckLists[i].Items {
			iName := templ.CheckLists[i].Items[j].Name
			iID := templ.CheckLists[i].Items[j].ID
			var cbData CallbackData = CallbackData{
				ListID:  iID,
				Command: "check item",
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})
			if templ.CheckLists[i].Items[j].State {
				keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData("☑   "+iName, string(outData)))
			} else {
				keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData(iName, string(outData)))
			}
			count++
		}
	}
	keyboard := tg.NewInlineKeyboardMarkup(keys...)
	msg := tg.NewMessage(user.ID, reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ShowTemplates(user User, bot *tg.BotAPI, command string) {
	filePath := "AppData/" + user.Name + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	if len(rawDataIn) == 17 || len(rawDataIn) == 19 {
		reply := "У вас нет шаблонов, вы можете их добавить"
		msg := tg.NewMessage(user.ID, reply)
		bot.Send(msg)
	} else {
		var templ CheckListJson
		err = json.Unmarshal(rawDataIn, &templ)
		if err != nil {
			log.Fatal("Invalid settings format:", err)
		}
		reply := "Ваши шаблоны"
		var keys [][]tg.InlineKeyboardButton
		for i := range templ.CheckLists {
			name := templ.CheckLists[i].Name
			id := templ.CheckLists[i].ID
			var cbData CallbackData
			if command == "show" {
				cbData = CallbackData{
					ListID:  id,
					Command: "show templ",
				}
			} else if command == "add" {
				cbData = CallbackData{
					ListID:  id,
					Command: "add to list",
				}
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})

			var rerrr string = string(outData)
			var qqqqq *string = &rerrr
			keys[i] = append(keys[i], tg.InlineKeyboardButton{
				CallbackData: qqqqq,
				Text:         name,
			})
		}
		keyboard := tg.NewInlineKeyboardMarkup(keys...)
		msg := tg.NewMessage(user.ID, reply)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

func InitUser(UserName string) {
	// user lists
	file1 := "AppData/" + UserName + ".json"
	os.Create(file1)

	clu := CheckListJson{}
	var rawDataOut []byte
	var err error
	rawDataOut, err = json.MarshalIndent(&clu, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}

	err = ioutil.WriteFile(file1, rawDataOut, 0664)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}

	// user templates
	file2 := "AppData/" + UserName + ".tem.json"
	os.Create(file2)

	clt := CheckListJson{}
	rawDataOut, err = json.MarshalIndent(&clt, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}

	err = ioutil.WriteFile(file2, rawDataOut, 0664)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}
}

func DataBase(TData chan TransactData) {
	var CL map[string]CheckList = make(map[string]CheckList)
	for {
		ldata := <-TData
		switch ldata.Command {
		case "add name":
			CL[ldata.UserName] = CheckList{
				Name: ldata.Data,
				ID:   xid.New().String(),
			}

		case "add item":
			if len(CL[ldata.UserName].Items) == 0 {
				CL[ldata.UserName] = CheckList{
					Name: CL[ldata.UserName].Name,
					ID:   CL[ldata.UserName].ID,
					Items: []Item{{
						Name:  ldata.Data,
						ID:    xid.New().String(),
						State: false,
					}},
				}
			} else {
				var lItems []Item
				for i := 0; i < len(CL[ldata.UserName].Items); i++ {
					lItems = append(lItems, CL[ldata.UserName].Items[i])
				}
				lItems = append(lItems, Item{
					Name:  ldata.Data,
					ID:    xid.New().String(),
					State: false,
				})
				CL[ldata.UserName] = CheckList{
					Name:  CL[ldata.UserName].Name,
					ID:    CL[ldata.UserName].ID,
					Items: lItems,
				}
			}

		case "edit templ":
			filePath := "AppData/" + ldata.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListJson
			err = json.Unmarshal(rawDataIn, &templ)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			for i := range templ.CheckLists {
				if templ.CheckLists[i].ID == ldata.Data {
					templ.CheckLists[i] = CL[ldata.UserName]
					break
				}
			}

			rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			delete(CL, ldata.UserName)

		case "save tmp data":
			filePath := "AppData/" + ldata.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListJson
			if len(rawDataIn) == 0 {
				templ.CheckLists = append(templ.CheckLists, CL[ldata.UserName])
			} else {
				err = json.Unmarshal(rawDataIn, &templ)
				if err != nil {
					log.Fatal("Invalid settings format:", err)
				}
				templ.CheckLists = append(templ.CheckLists, CL[ldata.UserName])
			}

			rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			delete(CL, ldata.UserName)

		case "delete templ":
			filePath := "AppData/" + ldata.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListJson
			err = json.Unmarshal(rawDataIn, &templ)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			templ.CheckLists = RemoveCheckList(templ.CheckLists, ldata.Data)
			rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}

		case "add from templ":
			file1 := "AppData/" + ldata.UserName + ".tem.json"
			file2 := "AppData/" + ldata.UserName + ".json"

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
				if cl1.CheckLists[i].ID == ldata.Data {
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
		case "delete List":
			filePath := "AppData/" + ldata.UserName + ".json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListJson
			err = json.Unmarshal(rawDataIn, &templ)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			templ.CheckLists = RemoveCheckList(templ.CheckLists, ldata.Data)
			rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
			if err != nil {
				log.Fatal("JSON marshaling failed:", err)
			}

			err = ioutil.WriteFile(filePath, rawDataOut, 0664)
			if err != nil {
				log.Fatal("Cannot write updated settings file:", err)
			}
			TData <- TransactData{}
		case "check item":
			filePath := "AppData/" + ldata.UserName + ".json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListJson
			err = json.Unmarshal(rawDataIn, &templ)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}

			count := 0
			for i := range templ.CheckLists {
				for j := range templ.CheckLists[i].Items {
					if templ.CheckLists[i].Items[j].State == true {
						count++
					}
					if ldata.Data == templ.CheckLists[i].Items[j].ID {
						templ.CheckLists[i].Items[j].State = true
					}
				}
			}
			rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
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

func CheckItem(user User, bot *tg.BotAPI) {
	filePath := "AppData/" + user.Name + ".json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var templ CheckListJson
	err = json.Unmarshal(rawDataIn, &templ)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}

	for i := range templ.CheckLists {
		for j := range templ.CheckLists[i].Items {
			if user.Data == templ.CheckLists[i].Items[j].ID {
				templ.CheckLists[i].Items[j].State = true
			}
		}
	}
	rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling failed:", err)
	}

	err = ioutil.WriteFile(filePath, rawDataOut, 0664)
	if err != nil {
		log.Fatal("Cannot write updated settings file:", err)
	}
}
