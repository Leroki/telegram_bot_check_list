package main

// структура данный передаваемых черз call back
type CallbackData struct {
	ListID  string `json:"list_id"`
	Command byte   `json:"command"`
}
