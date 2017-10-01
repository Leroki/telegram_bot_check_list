package main

import (
	"encoding/json"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"os"
	"strings"
	// "strconv"
)

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

type User struct {
	ID    int64
	State string
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

	// главный цикл
	for update := range updates {
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет "+update.Message.From.UserName+"! Я телеграм бот.")
			bot.Send(msg)
		case "stop":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пока "+update.Message.From.UserName+"!")
			bot.Send(msg)
			Users[update.Message.From.UserName] = User{
				ID:    update.Message.Chat.ID,
				State: "Stopped",
			}
		}

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
		}

		switch Users[update.Message.From.UserName].State {
		case "Main":
			keyRow1 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Листы"))
			keyRow2 := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Шаблоны"))
			keyboard := tgbotapi.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Вы в главном меню")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case "Stopped":
			var keys []tgbotapi.KeyboardButton
			keys = append(keys, tgbotapi.NewKeyboardButton("В главное меню"))
			keys = append(keys, tgbotapi.NewKeyboardButton("Что-то :)"))
			keyboard := tgbotapi.NewReplyKeyboard(keys)
			msg := tgbotapi.NewMessage(Users[update.Message.From.UserName].ID, "Вы в подменю")
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

		case "Add tmp item":

		}
	}
}

/*func RemoveUser(u []User, name string) []User  {
	var id int
	id = -1
	for i := range u {
		if u[i].Name == name {
			id = i
			break
		}
	}

	return append(u[:id], u[id + 1:]...)
}*/

func ShowTemplates(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	file, _ := os.Open("AppData/template.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	templ := CheckListTemplate{}
	err := decoder.Decode(&templ)
	if err != nil {
		log.Panic(err)
	}

	reply := ""

	for i := range templ.CheckLists {
		reply += templ.CheckLists[i].Name + ":\n"
		for y := range templ.CheckLists[i].Items {
			reply += templ.CheckLists[i].Items[y].Name + "\n"
		}
		reply += "\n"
	}

	if reply == "" {
		reply = "Функция показа шаблонов"
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}

func AddNewTemplate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	reply := "Функция добавления новых шаблонов"
	if update.Message.CommandArguments() == "" {
		reply = "Нет аргументов"
	} else {
		filePath := "AppData/template.json"
		rawDataIn, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal("Cannot load settings:", err)
		}

		var templ CheckListTemplate
		err = json.Unmarshal(rawDataIn, &templ)
		if err != nil {
			log.Fatal("Invalid settings format:", err)
		}

		arguments := strings.Split(update.Message.CommandArguments(), " ")

		var item []Item
		for i := 1; i < len(arguments); i++ {
			item = append(item, Item{
				Name:  arguments[i],
				State: false,
			})

		}

		newList := CheckList{
			Name:  arguments[0],
			Items: item,
		}

		templ.CheckLists = append(templ.CheckLists, newList)

		rawDataOut, err := json.MarshalIndent(&templ, "", "  ")
		if err != nil {
			log.Fatal("JSON marshaling failed:", err)
		}

		err = ioutil.WriteFile(filePath, rawDataOut, 0)
		if err != nil {
			log.Fatal("Cannot write updated settings file:", err)
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}
