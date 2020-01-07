package db

// User base struct db storage
type User struct {
	UserName      string      `bson:"UserName"`
	RootMessageID int         `bson:"RootMessageID"`
	CheckLists    []CheckList `bson:"CheckLists"`
}

// CheckList list with elemts
type CheckList struct {
	Name     string    `bson:"Name"`
	UUID     string    `bson:"UUID"`
	Status   bool      `bson:"Status"`
	Elements []Element `bson:"Elements"`
}

// Element single element check list
type Element struct {
	Name   string `bson:"Name"`
	UUID   string `bson:"UUID"`
	Status bool   `bson:"Status"`
}
