package main

// состояние пользователя во время исольнения кода
const stMain byte = 0
const stList byte = 1
const stShowTemp byte = 2
const stTemplates byte = 3
const stAddTmpName byte = 4
const stAddTmp byte = 5
const stAddTmpItem byte = 6
const stEditTmpName byte = 7
const stEditTmpItem byte = 8
const stEditTmp byte = 9
const stAddFromTemp byte = 10
const stDeleteFromList byte = 11

// команда из call back'а
const cbShowTemp byte = 0
const cbAddToList byte = 1
const cbCheckList byte = 2
const cbCheckItem byte = 3

// команда для DB
const trAddFromTemp byte = 0
const trAddName byte = 1
const trAddItem byte = 2
const trCheckItem byte = 3
const trSave byte = 4
const trEditTemp byte = 5
const trDelTemp byte = 6
const trDelList byte = 7
const trInitUser byte = 8
const trReturnTemp byte = 9
const trReturnList byte = 10

// команда от пользователя в чате
const cStart = "start"
const cStop = "stop"
const cShow byte = 0
const cAdd byte = 1

// кнопки у бота в чате
const btMain = "В главное меню"
const btLists = "Листы"
const btTemplates = "Шаблоны"
const btAddTemplate = "Добавить новый шаблон"
const btCancel = "Отмена"
const btBack = "Назад"
const btFinish = "Завершить"
const btEdit = "Изменить"
const btTemplDelete = "Удалить шаблон"
const btListDelete = "Удалить лист"
const btAddListFromTemplate = "Добавить новый лист из шаблона"
