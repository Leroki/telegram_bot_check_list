package main

type Item struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	State bool   `json:"state"`
}

type CheckList struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Items []Item `json:"items"`
}

type CheckListJson struct {
	UserName   string      `json:"user_name"`
	CheckLists []CheckList `json:"lists"`
}

type User struct {
	Name  string
	ID    int64
	State byte
	Data  string
}

type TransactData struct {
	UserName string
	Data     string
	Command  byte
	DataCL   CheckListJson
}

type CallbackData struct {
	ListID  string `json:"list_id"`
	Command byte   `json:"command"`
}
