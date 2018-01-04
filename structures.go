package main

// Структура элемента листа
type Item struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State bool   `json:"state"`
}

// структура листа
type CheckList struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Items []Item `json:"items"`
}

// структура хранения листов
type CheckListJson struct {
	UserName   string      `json:"user_name"`
	CheckLists []CheckList `json:"lists"`
}

// сруктура хранения пользователей в рантайме
type User struct {
	Name  string
	ID    int64
	State byte
	Data  string
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
