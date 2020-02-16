package main

import (
	"log"
	"os"

	tg "gopkg.in/telegram-bot-api.v4"
)

func main() {
	token := os.Getenv("bot_token")
	bot, err := tg.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	// главный цикл
	for update := range updates {
		log.Printf("%v\n", update)
	}
}
