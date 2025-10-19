[FLOW]:
    1.Создаётся игра (функция SetupGameHandler)
    2.Игроки добавляются в игру(функция AddPlayerToGame)
    3.Обновляется стутус игры на "playing"(функция UpdateGameStatus)
    4.Начинают отправляться таски (функция SendTasks)
    5.С таской отправляется две кнопки "Хочу відповісти" и "Пропустити". У каждой из этих кнопок есть свой обработчик - функции OnAnswerTaskBtnHandler и OnSkipTaskBtnHandler соответственно.
    6. Юзер может пропустить 3 таски, потом уже такой возможности нет, нужно только отвечать
    7.Когда юзер нажимает хочу ответить, он может отправить в чат игры любое сообщение и на это сообщение среагирует функция HandlerPlayerResponse, запишет факт ответа юзера в БД функция AddPlayerResponse и обновит статус юзера в БД функция UpdatePlayerStatus

[FUNCTIONS]:
    1.handlers.SetupGameHandler
    2.storage_db.CreateGame
    3.storage_db.AddPlayerToGame
    4.storage_db.UpdateGameStatus
    5.handlers.SendTasks 
    6.handlers.OnAnswerTaskBtnHandler
    7.handlers.OnSkipTaskBtnHandler
    8.handlers.HandlerPlayerResponse
    9.storage_db.AddPlayerResponse
    10.storage_db.UpdatePlayerStatus

[PACKAGES]:
    1.handlers
    2.storage_db
