package main

import (
	"encoding/json"
	"gopkg.in/telegram-bot-api.v4"
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
	ID    int64
	State string
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

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var Users map[string]User
	Users = make(map[string]User)

	var TData chan TransactData = make(chan TransactData)
	go DataBase(TData)

	// главный цикл
	for update := range updates {
		if update.CallbackQuery != nil {
			query := update.CallbackQuery
			var cbData CallbackData
			err := json.Unmarshal([]byte(query.Data), &cbData)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}
			log.Println(cbData)
			switch cbData.Command {
			case "show":
				go ShowList(cbData.ListName, update, bot)
			}
		}

		if update.Message == nil {
			continue
		}

		if Users[update.Message.From.UserName].ID == 0 {
			log.Println(Users[update.Message.From.UserName].ID)
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Main",
			}
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		// debug info
		log.Println(Users[update.Message.From.UserName].State)

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет "+update.Message.From.FirstName+"! Я телеграм бот.")
			bot.Send(msg)
			InitUser(update.Message.From.UserName)
		case "stop":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пока "+update.Message.From.UserName+"!")
			bot.Send(msg)
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Stopped",
			}
		}

		// обработка кнопок
		switch update.Message.Text {
		case "В главное меню":
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Main",
			}
		case "Шаблоны":
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Templates",
			}
		case "Показать мои шаблоны":
			go ShowTemplates(update, bot)
			continue
		case "Добавить новый шаблон":
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Add tmp",
			}
		case "Отмена":
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Templates",
			}
		case "Завершить":
			if Users[update.Message.From.UserName].State == "Add tmp item" {
				TData <- TransactData{
					UserName: update.Message.From.UserName,
					Data:     "",
					Command:  "save tmp data",
				}
				Users[update.Message.From.UserName] = User{
					ID:    update.Message.Chat.ID,
					State: "Templates",
				}
			}
		}

		// обработка положения в меню
		switch Users[update.Message.From.UserName].State {
		case "Main":
			keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Листы"))
			keyRow2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Шаблоны"))
			keyboard := tgbotapi.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Вы в главном меню")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Templates":
			keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Показать мои шаблоны"))
			keyRow2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Добавить новый шаблон"))
			keyRow3 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("В главное меню"))
			keyboard := tgbotapi.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Вы в меню шаблонов")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Add tmp":
			keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Отмена"))
			keyRow2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("В главное меню"))
			keyboard := tgbotapi.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Add tmp name",
			}
		case "Add tmp name":
			TData <- TransactData{
				UserName: update.Message.From.UserName,
				Data:     update.Message.Text,
				Command:  "add name",
			}
			keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Завершить"))
			keyboard := tgbotapi.NewReplyKeyboard(keyRow1)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Введите название элемента или нажмите кнопку завершить для формирования листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Add tmp item",
			}
		case "Add tmp item":
			TData <- TransactData{
				UserName: update.Message.From.UserName,
				Data:     update.Message.Text,
				Command:  "add item",
			}
		}
	}
}

/* func RemoveUser(u []User, name string) []User  {
 *     var id int
 *     id = -1
 *     for i := range u {
 *         if u[i].Name == name {
 *             id = i
 *             break
 *         }
 *     }
 *
 *     return append(u[:id], u[id + 1:]...)
 * } */
func ShowList(listName string, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
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
	log.Println(templ)

	// reply := ""
	// for i := range templ.CheckLists {
	// if templ.CheckLists[i].Name == listName {
	// for j := range templ.CheckLists[i].Items {
	// reply += templ.CheckLists[i].Items[j].Name + "\n"
	// }
	// }
	// }

	reply := ""
	for i := range templ.CheckLists {
		if templ.CheckLists[i].Name == listName {
			for y := range templ.CheckLists[i].Items {
				reply += templ.CheckLists[i].Items[y].Name + "\n"
			}
			break
		}
	}
	keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Изменить"))
	keyRow2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Удалить"))
	keyRow3 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Назад"))
	keyboard := tgbotapi.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
	msg := tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ShowTemplates(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	filePath := "AppData/" + update.Message.From.UserName + ".tem.json"
	rawDataIn, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Cannot load settings:", err)
	}

	if len(rawDataIn) == 0 {
		reply := "У вас нет шаблонов, вы можете их добавить"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)
	} else {
		var templ CheckListTemplate
		err = json.Unmarshal(rawDataIn, &templ)
		if err != nil {
			log.Fatal("Invalid settings format:", err)
		}
		reply := "Ваши листы"
		var keys [][]tgbotapi.InlineKeyboardButton
		for i := range templ.CheckLists {
			name := templ.CheckLists[i].Name
			var cbData CallbackData = CallbackData{
				ListName: name,
				Command:  "show",
			}
			outData, _ := json.Marshal(&cbData)
			keys = append(keys, []tgbotapi.InlineKeyboardButton{})
			keys[i] = append(keys[i], tgbotapi.NewInlineKeyboardButtonData(name, string(outData)))
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(keys...)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)
	}
}

func InitUser(UserName string) {
	file1 := "AppData/" + UserName + ".json"
	os.Create(file1)
	file2 := "AppData/" + UserName + ".tem.json"
	os.Create(file2)
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
		}
	}
}

/* func AddNewTemplate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
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
 *     msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
 *     bot.Send(msg)
 * } */
