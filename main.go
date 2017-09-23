package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
)

type Config struct {
	TelegramBotToken string
}

type CheckList struct {
	Name string		`json:"name"`
	Items []string	`json:"items"`
}

type CheckListTemplate struct {
	CheckLists []CheckList `json:"lists"`
}

func main() {
	os.Mkdir("./AppData", os.ModePerm)
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)

	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(configuration.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		// универсальный ответ на любое сообщение
		reply := "Не знаю что сказать"
		if update.Message == nil {
			continue
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			reply = "Привет. Я телеграм-бот"
		case "hello":
			reply = "world"
		case "showtemplates":
			go ShowTemplates(update, bot)
			continue
		case "addnewtemplate":
			go AddNewTemplate(update, bot)
			continue
		}

		// создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		// отправляем
		bot.Send(msg)
	}
}

func ShowTemplates(update tgbotapi.Update, bot *tgbotapi.BotAPI)  {
	file, _ := os.Open("AppData/template.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	clt := CheckListTemplate{}
	err := decoder.Decode(&clt)
	if err != nil {
		log.Panic(err)
	}

	reply := ""

	for i := range clt.CheckLists {
		reply += clt.CheckLists[i].Name + ":\n"
		for y := range clt.CheckLists[i].Items {
			reply += clt.CheckLists[i].Items[y] + "\n"
		}
		reply += "\n"
	}

	if reply == "" {
		reply = "Функция показа шаблонов"
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}

func AddNewTemplate(update tgbotapi.Update, bot *tgbotapi.BotAPI)  {
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

	newList := CheckList{
		Name:  arguments[0],
		Items: arguments[1:],
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

	reply := "Функция добавления новых шаблонов"
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	bot.Send(msg)
}