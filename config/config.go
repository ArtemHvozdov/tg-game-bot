package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Duratuions struct for settings bot
type TimeDurations struct {
	TimePauseMsgStartGameAndMsgJoinGame 	time.Duration // Pause before start game and join game
	TimeDeleteMsgUserIsAlreadyInGame 		time.Duration // Pause before delete message user is already in game
	TimeDeleteMsgOnlyAdmniCanStartGame 		time.Duration // Pause before delete message only admin can start game
	TimeDeleteMsgYouAlreadyStartedGame 		time.Duration // Pause before delete message you already started game
	TimePauseBeforeStartSendingTask 		time.Duration // Pause before sending task to user
	TimeDeleteMsgJoinGamerReminder 			time.Duration // Pause before delete message join gamer reminder
	TimeDeleteAlbumId 						time.Duration // Pause before delete album id
	TimePauseBetweenSendingTasks 			time.Duration // Pause between sending tasks to user
	TimeDeleteMsgYouAlreadyAnswered 		time.Duration // Pause before delete message you already answered
	TimeDeleteMsgYouAreNotInGame 			time.Duration // Pause before delete message you are not in game
	TimeDeleteMsgAwaitingAnswer 			time.Duration // Pause before delete message awaiting answer
	TimeDeleteMsgMaxSkipTasks 				time.Duration // Pause before delete message skip max tasks
	TimeDeleteMsgReturnToGame 				time.Duration // Pause before delete message return to game
	TimeDeleteMsgExitGame 					time.Duration // Pause before delete message exit game
	TimeDeleteMsgAdminExit 			time.Duration // Pause before delete message admin exit game
}

var (
	durations TimeDurations
)

var devDurations = TimeDurations{
	TimePauseMsgStartGameAndMsgJoinGame:   5 * time.Second,  // 5 seconds
	TimeDeleteMsgUserIsAlreadyInGame:      5 * time.Second, // 30 seconds Feature: change time to 5 seconds
	TimeDeleteMsgOnlyAdmniCanStartGame:	   5 * time.Second, // 30 seconds
	TimeDeleteMsgYouAlreadyStartedGame:	   5 * time.Minute,	 // 1 minute 		
	TimePauseBeforeStartSendingTask:	   1 * time.Minute,  // 1 minute				
	TimeDeleteAlbumId:					   2 * time.Minute,  // 2 minutes						
	TimePauseBetweenSendingTasks:		   10 * time.Minute,  // 3 minutes
	TimeDeleteMsgJoinGamerReminder:		   5 * time.Second,  // 5 seconds
	TimeDeleteMsgYouAlreadyAnswered:	   5 * time.Second,  // 5 seconds 			
	TimeDeleteMsgYouAreNotInGame: 		   5 * time.Second, // 5 seconds			
	TimeDeleteMsgAwaitingAnswer:		   5 * time.Second,  // 3 seconds 			
	TimeDeleteMsgMaxSkipTasks:			   5 * time.Second,  // 5 minutes
	TimeDeleteMsgReturnToGame:             5 * time.Second,  // 5 seconds	
	TimeDeleteMsgExitGame:                 5 * time.Second,  // 5 seconds
	TimeDeleteMsgAdminExit:         	   5 * time.Second,  // 5 seconds			 
}

var prodDurations = TimeDurations{
	TimePauseMsgStartGameAndMsgJoinGame:   5 * time.Second,  // 5 seconds
	TimeDeleteMsgUserIsAlreadyInGame:      30 * time.Second, // 30 seconds Feature: change time to 5 seconds
	TimeDeleteMsgOnlyAdmniCanStartGame:	   30 * time.Second, // 30 seconds
	TimeDeleteMsgYouAlreadyStartedGame:	   1 * time.Minute,	 // 1 minute 		
	TimePauseBeforeStartSendingTask:	   1 * time.Minute,  // 1 minute
	TimeDeleteMsgJoinGamerReminder:		   45 * time.Second,  // 5 seconds				
	TimeDeleteAlbumId:					   2 * time.Minute,  // 2 minutes						
	TimePauseBetweenSendingTasks:		   3 * time.Minute,  // 3 minutes
	TimeDeleteMsgYouAlreadyAnswered:	   5 * time.Second,  // 5 seconds 			
	TimeDeleteMsgYouAreNotInGame: 		   5 * time.Second, // 5 seconds			
	TimeDeleteMsgAwaitingAnswer:		   3 * time.Second,  // 3 seconds 			
	TimeDeleteMsgMaxSkipTasks:			   5 * time.Second,  // 5 minutes		
	TimeDeleteMsgReturnToGame:             5 * time.Second,  // 5 seconds	
	TimeDeleteMsgExitGame:                 5 * time.Second,  // 5 seconds
	TimeDeleteMsgAdminExit:           	   5 * time.Second,  // 5 seconds	 
}


// Config struct for settings bot
type Config struct {
	TelegramToken string // Token by telegram-bot
	DatabaseDir   string // name folder database
	DatabaseFile  string // Name database file
	Mode          string // Mode of bot (dev | prod)
	Durations     TimeDurations // Duratuions for settings bot
}

// LoadConfig load configuration from .env file
func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env, environment variables in use")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		panic("You need to set the TELEGRAM_TOKEN environment variable")
	}

	dbDir := os.Getenv("DATABASE_DIR")
	if dbDir == "" {
		panic("You need to set the DATABASE_DIR environment variable")
	}

	dbFile := os.Getenv("DATABASE_FILE")
	if dbFile == "" {
		panic("You need to set the DATABASE_FILE environment variable")
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		panic("You need to set the MODE environment variable")
	}

	switch mode {
	case "dev":
		durations = devDurations
	case "prod":
		durations = prodDurations
	default:
		panic("Invalid mode specified. Use 'dev' or 'prod'.")
	}

	return &Config{
		TelegramToken: token,
		DatabaseDir:   dbDir,
		DatabaseFile:  dbFile,
		Mode:		   mode,
		Durations: 	   durations,
	}
}
