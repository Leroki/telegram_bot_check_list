package main

import (
	"encoding/json"
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
	State bool   `json:"state"`
}

type CheckList struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

type CheckListTemplate struct {
	CheckLists []CheckList `json:"lists"`
}

type CheckListUser struct {
	CheckLists []CheckList `json:"lists"`
}

type User struct {
	ID         int64
	State      string
	PickedList string
}

type TransactData struct {
	UserName string
	Data     string
	Command  string
}

type CallbackData struct {
	ListName string `json:"listname"`
	Command  string `json:"command"`
}

func main() {
	os.Mkdir("./AppData", os.ModePerm)
	file, _ := os.Open("config.json")
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

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

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
					ID:         UserID,
					State:      "Show templ",
					PickedList: cbData.ListName,
				}
				go ShowList(cbData.ListName, update, bot)
			}
		}

		if update.Message == nil {
			continue
		}

		UserName := update.Message.From.UserName
		UserID := int64(update.Message.From.ID)

		if Users[UserName].ID == 0 {
			Users[UserName] = User{
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
				ID:    UserID,
				State: "Main",
			}
		case "Листы":
			Users[UserName] = User{
				ID:    UserID,
				State: "Lists",
			}
		case "Шаблоны":
			Users[UserName] = User{
				ID:    UserID,
				State: "Templates",
			}
		case "Показать мои шаблоны":
			go ShowTemplates(update, bot)
			continue
		case "Добавить новый шаблон":
			Users[UserName] = User{
				ID:    UserID,
				State: "Add tmp",
			}
		case "Отмена":
			Users[UserName] = User{
				ID:    UserID,
				State: "Templates",
			}
		case "Назад":
			if Users[UserName].State == "Show templ" {
				Users[UserName] = User{
					ID:    UserID,
					State: "Templates",
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
					ID:    UserID,
					State: "Templates",
				}
			} else if Users[UserName].State == "Edit tmp item" {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].PickedList,
					Command:  "edit templ",
				}
				Users[UserName] = User{
					ID:    UserID,
					State: "Templates",
				}
			}
		case "Изменить":
			if Users[UserName].State == "Show templ" {
				Users[UserName] = User{
					ID:         UserID,
					State:      "Edit tmp",
					PickedList: Users[UserName].PickedList,
				}
			}
		case "Удалить":
			if Users[UserName].State == "Show templ" {
				TData <- TransactData{
					UserName: UserName,
					Data:     Users[UserName].PickedList,
					Command:  "delete templ",
				}
				Users[UserName] = User{
					ID:    UserID,
					State: "Templates",
				}
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
			ShowCheckList(update, bot)
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Добавить новый лист из шаблона"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "")
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
			if len(update.Message.Text) > 26 {
				Users[UserName] = User{
					ID:    UserID,
					State: "Add tmp name",
				}
				msg := tg.NewMessage(Users[UserName].ID, "Название слишком длинное, попробуйте другое")
				bot.Send(msg)
			} else {
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
					ID:    UserID,
					State: "Add tmp item",
				}
			}
		case "Add tmp":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона (26 символов)")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
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
			if len(update.Message.Text) > 26 {
				Users[UserName] = User{
					ID:         UserID,
					State:      "Edit tmp name",
					PickedList: Users[UserName].PickedList,
				}
				msg := tg.NewMessage(Users[UserName].ID, "Название слишком длинное, попробуйте другое")
				bot.Send(msg)
			} else {
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
					ID:         UserID,
					State:      "Edit tmp item",
					PickedList: Users[UserName].PickedList,
				}
			}
		case "Edit tmp":
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("Отмена"))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton("В главное меню"))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона (26 символов)")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = User{
				ID:         UserID,
				State:      "Edit tmp name",
				PickedList: Users[UserName].PickedList,
			}
		case "Edit tmp item":
			TData <- TransactData{
				UserName: update.Message.From.UserName,
				Data:     update.Message.Text,
				Command:  "add item",
			}
		}
	}
}

func RemoveCheckList(u []CheckList, listName string) []CheckList {
	var id int
	id = -1
	for i := range u {
		if u[i].Name == listName {
			id = i
			break
		}
	}

	return append(u[:id], u[id+1:]...)
}

func ShowList(listName string, update tg.Update, bot *tg.BotAPI) {
	filePath := "AppData/" + update.CallbackQuery.From.UserName + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var templ CheckListTemplate
	err = json.Unmarshal(rawDataIn, &templ)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}

	reply := ""
	for i := range templ.CheckLists {
		if templ.CheckLists[i].Name == listName {
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
	msg := tg.NewMessage(int64(update.CallbackQuery.From.ID), reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ShowCheckList(update tg.Update, bot *tg.BotAPI) {
	filePath := "AppData/" + update.Message.From.UserName + ".json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	var templ CheckListUser
	err = json.Unmarshal(rawDataIn, &templ)
	if err != nil {
		log.Fatal("Invalid settings format:", err)
	}
	reply := "Ваши листы"
	var keys [][]tg.InlineKeyboardButton
	var count int = 0
	for i := range templ.CheckLists {
		lName := templ.CheckLists[i].Name
		var cbData CallbackData = CallbackData{
			ListName: lName,
			Command:  "show templ",
		}
		outData, _ := json.Marshal(&cbData)
		keys = append(keys, []tg.InlineKeyboardButton{})
		keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData(lName+" ⏬", string(outData)))
		count++
		for j := range templ.CheckLists[i].Items {
			iName := templ.CheckLists[i].Items[j].Name
			var cbData CallbackData = CallbackData{
				ListName: iName,
				Command:  "show templ",
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})
			keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData(iName, string(outData)))
			count++
		}
	}
	keyboard := tg.NewInlineKeyboardMarkup(keys...)
	msg := tg.NewMessage(update.Message.Chat.ID, reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ShowTemplates(update tg.Update, bot *tg.BotAPI) {
	filePath := "AppData/" + update.Message.From.UserName + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	if len(rawDataIn) == 17 || len(rawDataIn) == 19 {
		reply := "У вас нет шаблонов, вы можете их добавить"
		msg := tg.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	} else {
		var templ CheckListTemplate
		err = json.Unmarshal(rawDataIn, &templ)
		if err != nil {
			log.Fatal("Invalid settings format:", err)
		}
		reply := "Ваши листы"
		var keys [][]tg.InlineKeyboardButton
		for i := range templ.CheckLists {
			name := templ.CheckLists[i].Name
			var cbData CallbackData = CallbackData{
				ListName: name,
				Command:  "show templ",
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tg.InlineKeyboardButton{})
			keys[i] = append(keys[i], tg.NewInlineKeyboardButtonData(name, string(outData)))
		}
		keyboard := tg.NewInlineKeyboardMarkup(keys...)
		msg := tg.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

func InitUser(UserName string) {
	// user lists
	file1 := "AppData/" + UserName + ".json"
	os.Create(file1)

	clu := CheckListUser{}
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

	clt := CheckListTemplate{}
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
			}
		case "add item":
			if len(CL[ldata.UserName].Items) == 0 {
				CL[ldata.UserName] = CheckList{
					Name: CL[ldata.UserName].Name,
					Items: []Item{{
						Name:  ldata.Data,
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
					State: false,
				})
				CL[ldata.UserName] = CheckList{
					Name:  CL[ldata.UserName].Name,
					Items: lItems,
				}
			}
		case "edit templ":
			filePath := "AppData/" + ldata.UserName + ".tem.json"
			rawDataIn, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal("Cannot load settings:", err)
			}

			var templ CheckListTemplate
			err = json.Unmarshal(rawDataIn, &templ)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			for i := range templ.CheckLists {
				if templ.CheckLists[i].Name == ldata.Data {
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

			var templ CheckListTemplate
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

			var templ CheckListTemplate
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
		}
	}
}

/* func AddNewTemplate(update tg.Update, bot *tg.BotAPI) {
 *     reply := "Функция добавления новых шаблонов"
 *     if update.Message.CommandArguments() == "" {
 *         reply = "Нет аргументов"
 *     } else {
 *         filePath := "AppData/template.json"
 *         rawDataIn, err := ioutil.ReadFile(filePath)
 *         if err != nil {
 *             log.Fatal("Cannot load settings:", err)
 *         }
 *
 *         var templ CheckListTemplate
 *         err = json.Unmarshal(rawDataIn, &templ)
 *         if err != nil {
 *             log.Fatal("Invalid settings format:", err)
 *         }
 *
 *         arguments := strings.Split(update.Message.CommandArguments(), " ")
 *
 *         var item []Item
 *         for i := 1; i < len(arguments); i++ {
 *             item = append(item, Item{
 *                 Name:  arguments[i],
 *                 State: false,
 *             })
 *
 *         }
 *
 *         newList := CheckList{
 *             Name:  arguments[0],
 *             Items: item,
 *         }
 *
 *         templ.CheckLists = append(templ.CheckLists, newList)
 *
 *         rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
 *         if err != nil {
 *             log.Fatal("JSON marshaling failed:", err)
 *         }
 *
 *         err = ioutil.WriteFile(filePath, rawDataOut, 0)
 *         if err != nil {
 *             log.Fatal("Cannot write updated settings file:", err)
 *         }
 *     }
 *
 *     msg := tg.NewMessage(update.Message.Chat.ID, reply)
 *     bot.Send(msg)
 * } */
