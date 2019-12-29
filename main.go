package main

import "github.com/Leroki/telegram_bot_check_list/db"

func main() {
	dataBase := db.Init()
	userName := "Nikita"
	dataBase.CreateUser(&userName)
	dataBase.DeleteUser(&userName)
}
