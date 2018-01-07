package main

// состояние пользователя во время исольнения кода
const STMain byte = 0
const STList byte = 1
const STShowTemp byte = 2
const STTemplates byte = 3
const STAddTmpName byte = 4
const STAddTmp byte = 5
const STAddTmpItem byte = 6
const STEditTmpName byte = 7
const STEditTmpItem byte = 8
const STEditTmp byte = 9
const STAddFromTemp byte = 10
const STDeleteFromList byte = 11

// команда из call back'а
const CBShowTemp byte = 0
const CBAddToList byte = 1
const CBCheckList byte = 2
const CBCheckItem byte = 3

// команда для DB
const TRAddFromTemp byte = 0
const TRAddName byte = 1
const TRAddItem byte = 2
const TRCheckItem byte = 3
const TRSave byte = 4
const TREditTemp byte = 5
const TRDelTemp byte = 6
const TRDelList byte = 7
const TRInitUser byte = 8
const TRReturnTemp byte = 9
const TRReturnList byte = 10

// команда от пользователя в чате
const CMDStart = "start"
const CMDStop = "stop"

// кнопки у бота в чате
const BTNMain = "В главное меню"
const BTNLists = "Листы"
const BTNTemplates = "Шаблоны"
const BTNShowTemplates = "Показать мои шаблоны"
const BTNAddTemplate = "Добавить новый шаблон"
const BTNCancel = "Отмена"
const BTNBack = "Назад"
const BTNFinish = "Завершить"
const BTNEdit = "Изменить"
const BTNDelete = "Удалить"
const BTNAddListFromTemplate = "Добавить новый лист из шаблона"
