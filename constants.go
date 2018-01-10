package main

// состояние пользователя во время исольнения кода
const (
	STMain           byte = iota
	STList
	STShowTemp
	STTemplates
	STAddTmpName
	STAddTmp
	STAddTmpItem
	STEditTmpName
	STEditTmpItem
	STEditTmp
	STAddFromTemp
	STDeleteFromList
)

// команда из call back'а
const (
	CBShowTemp  byte = iota
	CBAddToList
	CBCheckList
	CBCheckItem
)

// команда для DB
const (
	TRAddFromTemp byte = iota
	TRAddName
	TRAddItem
	TRCheckItem
	TRSave
	TREditTemp
	TRDelTemp
	TRDelList
	TRInitUser
	TRReturnTemp
	TRReturnList
)

// команда от пользователя в чате
const (
	CMDStart      = "start"
	CMDStop       = "stop"
	CMDShow  byte = 0
	CMDAdd   byte = 1
)

// кнопки у бота в чате
const (
	BTNMain                = "В главное меню"
	BTNLists               = "Листы"
	BTNTemplates           = "Шаблоны"
	BTNAddTemplate         = "Добавить новый шаблон"
	BTNCancel              = "Отмена"
	BTNBack                = "Назад"
	BTNFinish              = "Завершить"
	BTNEdit                = "Изменить"
	BTNTemplDelete         = "Удалить шаблон"
	BTNListDelete          = "Удалить лист"
	BTNAddListFromTemplate = "Добавить новый лист из шаблона"
)
