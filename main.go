package main

import (
	"log"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
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

		msg := tg.NewMessage(update.Message.Chat.ID, "Что сделать с листом?")

		message, err := bot.Send(msg)
		if err != nil {
			log.Printf("error: %v\n", err)
		}

		_, err = bot.DeleteMessage(tg.DeleteMessageConfig{
			ChatID:    message.Chat.ID,
			MessageID: message.MessageID,
		})

		if err != nil {
			log.Printf("error: %v\n", err)
		}

		_, err = bot.DeleteMessage(tg.DeleteMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.MessageID,
		})

		if err != nil {
			log.Printf("error: %v\n", err)
		}
	}
}
