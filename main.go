package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArtemHvozdov/tg-game-bot.git/config"
	"github.com/ArtemHvozdov/tg-game-bot.git/handlers"
	"github.com/ArtemHvozdov/tg-game-bot.git/internal/msgmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/pkg/btnmanager"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"

	"gopkg.in/telebot.v3"
)
func main() {
	cfg := config.LoadConfig() // Loading the configuration from a file or environment variable

	utils.InitNewLogger()

	// Initialize the database
	// dataDir := cfg.DatabaseDir
	// dataFile := cfg.DatabaseFile
	// dataDir := "../data/"
	// dataFile := "tg-game-bot.db"
	// dbPath := dataDir + dataFile

	// if err := os.MkdirAll(dataDir, 0755); err != nil {
	// 	utils.Logger.WithError(err).Fatal("Error creating folder:")
	// }

	dataDir := "/app/data"
	dataFile := "tg-game-bot.db"
	dbPath := fmt.Sprintf("%s/%s", dataDir, dataFile)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Println("Error creating folder:")
		//log.Fatalf("Error creating folder %s: %v", dataDir, err)
	}

	
	db, err := storage_db.InitDB(dbPath)
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed ti initialize database")
		//log.Fatalf("Failed to initialize database: %v", err)
	}
	defer storage_db.CloseDB(db)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	pref := telebot.Settings{
		Token: cfg.TelegramToken,
		Poller: &telebot.LongPoller{
			Timeout: 10,
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Failed to create a new bot: %v", err)
	}

	//handlers.InitMessageManager(bot)
	msgmanager.Init(bot)

	if err := btnmanager.Init("internal/data/buttons/buttons.json"); err != nil {
		log.Fatalf("Failed to initialize button manager: %v", err)
	}

	err = bot.SetCommands([]telebot.Command{})
	if err != nil {
		utils.Logger.Errorf("Failed to clear commands: %v", err)
	}

	// err = bot.SetCommands([]telebot.Command{
	// 	{Text: "start", Description: "Запустити бота"},
	// 	{Text: "help", Description: "Хелп мі"},
	// 	//{Text: "check_admin_bot", Description: "Перевірити права бота"},
	// })
	// if err != nil {
	// 	utils.Logger.Error("Failed to set bot commands")
	// 	//log.Printf("Failed to set bot commands: %v", err)
	// }

	// Create buttons
	//btnCreateGame := telebot.Btn{Text: "Створити гру"}
	//btnStartGame := telebot.Btn{Text: "Почати гру"}
	//btnJoinGame := telebot.Btn{Text: "Доєднатися до гри"}
	btnHelpMe := telebot.Btn{Text: "Help me!"}

	// Button handlers
	bot.Handle(&btnHelpMe, handlers.HelpMeHandler(bot))
	bot.Handle(&telebot.Btn{Unique: "answer_task"}, handlers.OnAnswerTaskBtnHandler(bot))
	bot.Handle(&telebot.Btn{Unique: "skip_task"}, handlers.OnSkipTaskBtnHandler(bot))
	bot.Handle(&telebot.Btn{Unique: "join_game_btn"}, handlers.JoinBtnHandler(bot))
	bot.Handle(&telebot.Btn{Unique: "start_game"}, handlers.StartGameHandlerFoo(bot))
	
	//bot.Handle(telebot.OnUserJoined, handlers.HandleUserJoined(bot))
	//bot.Handle(telebot.OnText, handlers.OnTextMsgHandler(bot))

	bot.Handle(telebot.OnText, handlers.HandlerPlayerResponse(bot))
	bot.Handle(telebot.OnPhoto, handlers.HandlerPlayerResponse(bot))
	bot.Handle(telebot.OnVideo, handlers.HandlerPlayerResponse(bot))
	bot.Handle(telebot.OnVoice, handlers.HandlerPlayerResponse(bot))
	bot.Handle(telebot.OnVideoNote, handlers.HandlerPlayerResponse(bot))
	
	bot.Handle(telebot.OnAddedToGroup, handlers.HandleBotAddedToGroup(bot))

	// Команда для создания опроса
	//bot.Handle("/color", handlers.SendColorQuestion(bot))
	bot.Handle("/test", handlers.TestRunHandler(bot))
	bot.Handle("/test_start", handlers.SendStartGameMessages(bot))
	bot.Handle("/test_finish", handlers.FinishTestHandler(bot))
	bot.Handle("/test_referal", handlers.SendReferalMsg(bot))
	bot.Handle("/test_feedback", handlers.SendFeedbackMsg(bot))
	bot.Handle("/test_coffee", handlers.SendBuyMeCoffeeMsg(bot))
	//bot.Handle("/test_ref_link", handlers.GetReferalLinkHandler(bot))
	//bot.Handle("/photo_task", handlers.SendPhotoTask(bot))
	bot.Handle("/create", handlers.CreateCollageFromResultsImageNine(bot))
	// Register handler for showing results
	bot.Handle("/subtask_results", handlers.SendSubtaskResultsToChat(bot))
	bot.Handle("/subtask10_results", handlers.CreateSubtask10Collage(bot))
	

	// bot.Handle(&telebot.InlineButton{Data: "color_answer_1"}, handlers.HandleColorAnswer(bot))
    // bot.Handle(&telebot.InlineButton{Data: "color_answer_2"}, handlers.HandleColorAnswer(bot))
    // bot.Handle(&telebot.InlineButton{Data: "color_answer_3"}, handlers.HandleColorAnswer(bot))
    // bot.Handle(&telebot.InlineButton{Data: "color_answer_4"}, handlers.HandleColorAnswer(bot))

	// bot.Handle(&telebot.InlineButton{Data: "photo_task_start"}, handlers.HandlePhotoTaskStart(bot))
    //bot.Handle(&telebot.InlineButton{Data: "photo_task_skip"}, handlers.HandlePhotoTaskSkip(bot))
    
	// bot.Handle("\fphoto_task_start", handlers.HandlePhotoTaskStart(bot))
	// bot.Handle("\fphoto_task_skip", handlers.HandlePhotoTaskSkip(bot))
	// bot.Handle("\fphoto_choice_", handlers.HandlePhotoChoice(bot))
	
	handlers.RegisterCallbackHandlers(bot)

	handlers.InitLoaderMessages()
	

	//bot.Handle("/start", handlers.StartHandler(bot, btnCreateGame, btnHelpMe))
	bot.Handle("/start", handlers.StartHandler(bot))
	bot.Handle("/help", handlers.HelpMeHandler(bot))
	//bot.Handle("/check_admin_bot", handlers.CheckAdminBotHandler(bot, btnStartGame))
	

	go func() {
		bot.Start()
	}()
  
  utils.Logger.Info("Bot started successfully")

	<-stop // Wait Ctrl+C or SIGTERM

	// if cfg.Mode == "dev" {
	// 	err := os.RemoveAll(dataDir)
	// 	if err != nil {
	// 		utils.Logger.Errorf("Failed to remove DB dir: %v", err)
	// 	} else {
	// 		utils.Logger.Info("DB dir removed (dev mode).")
	// 	}
	// } else {
	// 	utils.Logger.Info("Prod mode — DB dir not removed.")
	// }
}
