package main

import (
	"encoding/json"
	tg "gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	os.Mkdir("AppData", os.ModePerm)

	token := os.Getenv("bot_token")

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var Users = make(map[string]User)

	var TData = make(chan TransactData)
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
			case CBShowTemp:
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STShowTemp,
					Data:  cbData.ListID,
				}
				go ShowList(cbData.ListID, Users[UserName], bot)
			case CBAddToList:
				TData <- TransactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  TRAddFromTemp,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STList,
				}
				<-TData
				ShowCheckList(Users[UserName], bot)

			case CBCheckList:
				Users[UserName] = User{
					Name:  UserName,
					Data:  cbData.ListID,
					ID:    UserID,
					State: STDeleteFromList,
				}
				keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Удалить"))
				keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Назад"))
				keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
				msg := tg.NewMessage(Users[UserName].ID, "Вы в главном меню")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			case CBCheckItem:
				TData <- TransactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  TRCheckItem,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					Data:  cbData.ListID,
					State: STList,
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
				State: STMain,
			}
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", UserName, update.Message.Text)
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
				State: STMain,
			}
		case "Листы":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STList,
			}
			ShowCheckList(Users[UserName], bot)
		case "Шаблоны":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STTemplates,
			}
		case "Показать мои шаблоны":
			go ShowTemplates(Users[UserName], bot, "show")
			continue
		case "Добавить новый шаблон":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STAddTmp,
			}
		case "Отмена":
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STTemplates,
			}
		case "Назад":
			if Users[UserName].State == STShowTemp {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STTemplates,
				}
			} else if Users[UserName].State == STDeleteFromList {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STList,
				}
			}
		case "Завершить":
			if Users[UserName].State == STAddTmpItem {
				TData <- TransactData{
					UserName: UserName,
					Data:     "",
					Command:  TRSave,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STTemplates,
				}
			} else if Users[UserName].State == STEditTmpItem {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  TREditTemp,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STTemplates,
				}
			}
		case "Изменить":
			if Users[UserName].State == STShowTemp {
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STEditTmp,
					Data:  Users[UserName].Data,
				}
			}
		case "Удалить":
			if Users[UserName].State == STShowTemp {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  TRDelTemp,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STTemplates,
				}
			} else if Users[UserName].State == STDeleteFromList {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  TRDelList,
				}
				Users[UserName] = User{
					Name:  UserName,
					ID:    UserID,
					State: STList,
				}
				<-TData
				ShowCheckList(Users[UserName], bot)
			}
		case "Добавить новый лист из шаблона":
			ShowTemplates(Users[UserName], bot, "add")
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STAddFromTemp,
			}
		}

		// обработка положения в меню
		switch Users[UserName].State {
		case STMain:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Листы"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Шаблоны"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в главном меню")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case STList:
			// ShowCheckList(Users[UserName], bot)
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Добавить новый лист из шаблона"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Что сделать?")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case STTemplates:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Показать мои шаблоны"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Добавить новый шаблон"))
			keyRow3 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в меню шаблонов")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case STAddTmpName:
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  TRAddName,
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Завершить"))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для формирования листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STAddTmpItem,
			}

		case STAddTmp:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STAddTmpName,
			}
		case STAddTmpItem:
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  TRAddItem,
			}
		case STEditTmpName:
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  TRAddName,
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Завершить"))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для изменения листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STEditTmpItem,
				Data:  Users[UserName].Data,
			}
		case STEditTmp:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				Name:  UserName,
				ID:    UserID,
				State: STEditTmpName,
				Data:  Users[UserName].Data,
			}
		case STEditTmpItem:
			TData <- TransactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  TRAddItem,
			}
		}
	}
}

func ShowList(ListID string, user User, bot *tg.BotAPI) {
	filePath := "AppData/" + user.Name + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var temp CheckListJson
	err = json.Unmarshal(rawDataIn, &temp)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}

	reply := ""
	for i := range temp.CheckLists {
		if temp.CheckLists[i].ID == ListID {
			for y := range temp.CheckLists[i].Items {
				reply += temp.CheckLists[i].Items[y].Name + "\n"
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

	var temp CheckListJson
	err = json.Unmarshal(rawDataIn, &temp)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}
	reply := "Ваши листы"
	var keys [][]tg.InlineKeyboardButton
	var count = 0
	for i := range temp.CheckLists {
		lName := temp.CheckLists[i].Name
		lID := temp.CheckLists[i].ID
		var cbData = CallbackData{
			ListID:  lID,
			Command: CBCheckList,
		}
		outData, _ := json.Marshal(&cbData)
		keys = append(keys, []tg.InlineKeyboardButton{})
		keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData("⏬   "+lName+"   ⏬", string(outData)))
		count++
		for j := range temp.CheckLists[i].Items {
			iName := temp.CheckLists[i].Items[j].Name
			iID := temp.CheckLists[i].Items[j].ID
			var cbData = CallbackData{
				ListID:  iID,
				Command: CBCheckItem,
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})
			if temp.CheckLists[i].Items[j].State {
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
		var temp CheckListJson
		err = json.Unmarshal(rawDataIn, &temp)
		if err != nil {
			log.Fatal("Invalid settings format:", err)
		}
		reply := "Ваши шаблоны"
		var keys [][]tg.InlineKeyboardButton
		for i := range temp.CheckLists {
			name := temp.CheckLists[i].Name
			id := temp.CheckLists[i].ID
			var cbData CallbackData
			if command == "show" {
				cbData = CallbackData{
					ListID:  id,
					Command: CBShowTemp,
				}
			} else if command == "add" {
				cbData = CallbackData{
					ListID:  id,
					Command: CBAddToList,
				}
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})
			keys[i] = append(keys[i], tg.NewInlineKeyboardButtonData(name, string(outData)))
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
