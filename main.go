package main

import (
	// "log"
	// "os"
	"github.com/Leroki/telegram_bot_check_list/db"
	// tg "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	dataBase := db.Init()
	userName := "Nikita"
	dataBase.CreateUser(&userName)
	// bot, err := tg.NewBotAPI(os.Getenv("bot_token"))
	// if err != nil {
	// 	log.Panic(err)
	// }

	// bot.Debug = true

	// log.Printf("Authorized on account %s", bot.Self.UserName)

	// u := tg.NewUpdate(0)
	// u.Timeout = 60

	// updates, err := bot.GetUpdatesChan(u)
	// if err != nil {
	// 	log.Panic(err)
	// }

	// // users := map[string]*User{}

	// // главный цикл
	// for update := range updates {
	// 	messageID := update.Message.MessageID
	// 	log.Printf("message %v", messageID)
	// }
}
