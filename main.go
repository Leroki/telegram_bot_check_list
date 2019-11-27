package telegrambot

import (
	"encoding/json"
	"log"
	"os"

	tg "gopkg.in/telegram-bot-api.v4"
)

type userMap map[string]*user

func main() {
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

	Users := make(userMap)

	TData := make(chan transactData)
	go dataBase(TData)

	// главный цикл
	for update := range updates {
		// обработка callback'ов
		if update.CallbackQuery != nil {
			UserName := update.CallbackQuery.From.UserName
			query := update.CallbackQuery

			var cbData callbackData
			err := json.Unmarshal([]byte(query.Data), &cbData)
			if err != nil {
				log.Fatal("Invalid settings format:", err)
			}

			switch cbData.Command {
			case cbShowTemp: // call back на показ шаблона
				Users[UserName].State = stShowTemp
				Users[UserName].Data = cbData.ListID

				showTemplateList(cbData.ListID, Users[UserName], bot, &TData)
			case cbAddToList: // call back на добавление шаблона в чек лист
				TData <- transactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  trAddFromTemp,
				}
				Users[UserName].State = stList
				<-TData
				Users[UserName].MsgID = showCheckList(Users[UserName], bot, &TData, false)
			case cbCheckList: // call back для показа действий над чек листом
				Users[UserName].State = stDeleteFromList
				Users[UserName].Data = cbData.ListID

				keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btListDelete))
				keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btBack))
				keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
				msg := tg.NewMessage(Users[UserName].ID, "Что сделать с листом?")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case cbCheckItem: // call back для отметки элемента листа
				TData <- transactData{
					Data:     cbData.ListID,
					UserName: UserName,
					Command:  trCheckItem,
				}

				Users[UserName].State = stList
				Users[UserName].Data = cbData.ListID
				<-TData
				showCheckList(Users[UserName], bot, &TData, true)
			}
		}

		if update.Message == nil {
			continue
		}

		UserName := update.Message.From.UserName
		UserID := int64(update.Message.From.ID)

		if Users[UserName] == nil {
			Users[UserName] = &user{
				Name:  UserName,
				ID:    UserID,
				State: stMain,
			}
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", UserName, update.Message.Text)
		// debug info

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case cStart:
			msg := tg.NewMessage(UserID, "Привет "+update.Message.From.FirstName+" Я телеграм бот.")
			bot.Send(msg)
			TData <- transactData{
				UserName: UserName,
				Data:     "",
				Command:  trInitUser,
			}
		case cStop:
			msg := tg.NewMessage(UserID, "Пока "+update.Message.From.FirstName+"!")
			bot.Send(msg)
			delete(Users, UserName)
		}

		// обработка кнопок
		switch update.Message.Text {
		case btMain:
			Users[UserName].State = stMain
		case btLists:
			Users[UserName].State = stList
		case btTemplates:
			Users[UserName].State = stTemplates
		case btAddTemplate:
			Users[UserName].State = stAddTmp
		case btCancel:
			Users[UserName].State = stTemplates
		case btBack:
			if Users[UserName].State == stShowTemp {
				Users[UserName].State = stTemplates
			} else if Users[UserName].State == stDeleteFromList {
				Users[UserName].State = stList
			}
		case btFinish:
			if Users[UserName].State == stAddTmpItem {
				TData <- transactData{
					UserName: UserName,
					Data:     "",
					Command:  trSave,
				}
				Users[UserName].State = stTemplates
			} else if Users[UserName].State == stEditTmpItem {
				TData <- transactData{
					UserName: UserName,
					Data:     Users[UserName].Data,
					Command:  trEditTemp,
				}
				Users[UserName].State = stTemplates
			}
		case btEdit:
			if Users[UserName].State == stShowTemp {
				Users[UserName].State = stEditTmp
			}
		case btTemplDelete:
			TData <- transactData{
				UserName: UserName,
				Data:     Users[UserName].Data,
				Command:  trDelTemp,
			}
			Users[UserName].State = stTemplates
			<-TData
		case btListDelete:
			TData <- transactData{
				UserName: UserName,
				Data:     Users[UserName].Data,
				Command:  trDelList,
			}
			Users[UserName].State = stList
			<-TData
		case btAddListFromTemplate:
			showTemplates(Users[UserName], bot, cAdd, &TData)
			Users[UserName].State = stAddFromTemp
		}

		// обработка положения в меню
		switch Users[UserName].State {
		case stMain:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btLists))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btTemplates))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в главном меню")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		case stList:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btAddListFromTemplate))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btMain))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в меню чек листов")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName].MsgID = showCheckList(Users[UserName], bot, &TData, false)
		case stTemplates:
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btAddTemplate))
			keyRow3 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btMain))
			keyboard := tg.NewReplyKeyboard(keyRow2, keyRow3)
			msg := tg.NewMessage(Users[UserName].ID, "Вы в меню шаблонов")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			showTemplates(Users[UserName], bot, cShow, &TData)
		case stAddTmpName:
			TData <- transactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  trAddName,
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btFinish))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для формирования листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = &user{
				Name:  UserName,
				ID:    UserID,
				State: stAddTmpItem,
			}
		case stAddTmp:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btCancel))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btMain))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = &user{
				Name:  UserName,
				ID:    UserID,
				State: stAddTmpName,
			}
		case stAddTmpItem:
			TData <- transactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  trAddItem,
			}
		case stEditTmpName:
			TData <- transactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  trAddName,
			}
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btFinish))
			keyboard := tg.NewReplyKeyboard(keyRow1)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название элемента или нажмите кнопку завершить для изменения листа")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = &user{
				Name:  UserName,
				ID:    UserID,
				State: stEditTmpItem,
				Data:  Users[UserName].Data,
			}
		case stEditTmp:
			keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btCancel))
			keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btMain))
			keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2)
			msg := tg.NewMessage(Users[UserName].ID, "Введите название шаблона")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			Users[UserName] = &user{
				Name:  UserName,
				ID:    UserID,
				State: stEditTmpName,
				Data:  Users[UserName].Data,
			}
		case stEditTmpItem:
			TData <- transactData{
				UserName: UserName,
				Data:     update.Message.Text,
				Command:  trAddItem,
			}
		}
	}
}

func showTemplateList(ListID string, user *user, bot *tg.BotAPI, TData *chan transactData) {
	*TData <- transactData{
		UserName: user.Name,
		Command:  trReturnTemp,
	}
	var locTData = <-*TData
	var temp = locTData.DataCL

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
	keyRow1 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btEdit))
	keyRow2 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btTemplDelete))
	keyRow3 := tg.NewKeyboardButtonRow(tg.NewKeyboardButton(btBack))
	keyboard := tg.NewReplyKeyboard(keyRow1, keyRow2, keyRow3)
	msg := tg.NewMessage(user.ID, reply)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showCheckList(user *user, bot *tg.BotAPI, TData *chan transactData, toEdit bool) int {
	*TData <- transactData{
		UserName: user.Name,
		Command:  trReturnList,
	}

	var locTData = <-*TData
	var temp = locTData.DataCL

	reply := "Ваши листы"
	var keys [][]tg.InlineKeyboardButton
	var count = 0
	for i := range temp.CheckLists {
		lName := temp.CheckLists[i].Name
		lID := temp.CheckLists[i].ID
		var cbData = callbackData{
			ListID:  lID,
			Command: cbCheckList,
		}
		outData, _ := json.Marshal(&cbData)
		keys = append(keys, []tg.InlineKeyboardButton{})
		keys[count] = append(keys[count], tg.NewInlineKeyboardButtonData("⏬  "+lName+"  ⏬", string(outData)))
		count++
		for j := range temp.CheckLists[i].Items {
			iName := temp.CheckLists[i].Items[j].Name
			iID := temp.CheckLists[i].Items[j].ID
			var cbData = callbackData{
				ListID:  iID,
				Command: cbCheckItem,
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
	if toEdit {
		msg := tg.NewEditMessageReplyMarkup(user.ID, user.MsgID, keyboard)
		infoMsg, _ := bot.Send(msg)
		return infoMsg.MessageID
	}
	msg := tg.NewMessage(user.ID, reply)
	msg.ReplyMarkup = keyboard
	infoMsg, _ := bot.Send(msg)
	return infoMsg.MessageID
}

func showTemplates(user *user, bot *tg.BotAPI, command byte, TData *chan transactData) {
	*TData <- transactData{
		UserName: user.Name,
		Command:  trReturnTemp,
	}
	var locTData = <-*TData
	var temp = locTData.DataCL

	if len(temp.CheckLists) == 0 {
		reply := "У вас нет шаблонов, вы можете их добавить"
		msg := tg.NewMessage(user.ID, reply)
		bot.Send(msg)
	} else {
		reply := "Ваши шаблоны"
		var keys [][]tg.InlineKeyboardButton
		for i := range temp.CheckLists {
			name := temp.CheckLists[i].Name
			id := temp.CheckLists[i].ID
			var cbData callbackData
			if command == cShow {
				cbData = callbackData{
					ListID:  id,
					Command: cbShowTemp,
				}
			} else if command == cAdd {
				cbData = callbackData{
					ListID:  id,
					Command: cbAddToList,
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
