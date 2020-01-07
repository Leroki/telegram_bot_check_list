package main

import (
	"context"

	"github.com/Leroki/telegram_bot_check_list/db"
)

func main() {
	ctx := context.Background()
	dataBase := db.Init(&ctx)
	userName := "Nikita"
	dataBase.CreateUser(&userName)
	dataBase.DeleteUser(&userName)
}
