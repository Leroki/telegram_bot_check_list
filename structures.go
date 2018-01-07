package main

import "time"

// Структура элемента листа
type Item struct {
	Name  string `bson:"name"`
	ID    string `bson:"id"`
	State bool   `bson:"state"`
}

// структура листа
type CheckList struct {
	Name      string    `bson:"name"`
	ID        string    `bson:"id"`
	TimeStart time.Time `bson:"time_start"`
	FlagStart bool      `bson:"flag_start"`
	Items     []Item    `bson:"items"`
}

// структура хранения листов
type CheckListJson struct {
	UserName   string      `bson:"user_name"`
	CheckLists []CheckList `bson:"lists"`
}

// сруктура хранения пользователей в рантайме
type User struct {
	Name  string
	ID    int64
	State byte
	Data  string
	MsgId int
}

// структура данных передаваемых в chanel к DB
type TransactData struct {
	UserName string
	Data     string
	Command  byte
	DataCL   CheckListJson
}

// структура данный передаваемых черз call back
type CallbackData struct {
	ListID  string `json:"list_id"`
	Command byte   `json:"command"`
}
