package main

// сруктура хранения пользователей в рантайме
type User struct {
	Name  string
	ID    int64
	State byte
	Data  string
	MsgId int
}

func (u *User) Init() {

}

