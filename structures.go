package telegrambot

import "time"

// Структура элемента листа
type item struct {
	Name  string `bson:"name"`
	ID    string `bson:"id"`
	State bool   `bson:"state"`
}

// структура листа
type checkList struct {
	Name      string    `bson:"name"`
	ID        string    `bson:"id"`
	TimeStart time.Time `bson:"time_start"`
	FlagStart bool      `bson:"flag_start"`
	Items     []item    `bson:"items"`
}

// структура хранения листов
type checkListJSON struct {
	UserName   string      `bson:"user_name"`
	CheckLists []checkList `bson:"lists"`
}

// сруктура хранения пользователей в рантайме
type user struct {
	Name  string
	ID    int64
	State byte
	Data  string
	MsgID int
}

// структура данных передаваемых в chanel к DB
type transactData struct {
	UserName string
	Data     string
	Command  byte
	DataCL   checkListJSON
}

// структура данный передаваемых черз call back
type callbackData struct {
	ListID  string `json:"list_id"`
	Command byte   `json:"command"`
}
