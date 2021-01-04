package db

import "github.com/google/uuid"

// User base struct db storage.
type User struct {
	UserName      string      `bson:"user_name"`
	RootMessageID int         `bson:"root_message_id"`
	CheckLists    []CheckList `bson:"check_lists"`
}

// CheckList list with elements.
type CheckList struct {
	Name     string    `bson:"name"`
	UUID     uuid.UUID `bson:"uuid"`
	Status   bool      `bson:"status"`
	Elements []Element `bson:"elements"`
}

// Element single element check list.
type Element struct {
	Name   string    `bson:"name"`
	UUID   uuid.UUID `bson:"uuid"`
	Status bool      `bson:"status"`
}
