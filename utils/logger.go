package utils

import (
	"io"
	"os"

	logrus "github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitNewLogger() *logrus.Logger {
	Logger = logrus.New()

	level := logrus.InfoLevel
	if os.Getenv("DEBUG") == "true" {
		level = logrus.DebugLevel
	}
	Logger.SetLevel(level)

	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			multiWriter := io.MultiWriter(os.Stdout, file)
			Logger.SetOutput(multiWriter)
		} else {
			Logger.Warnf("Не вдалося відкрити файл логів %s: %v. Використовую стандартний stderr", logFile, err)
		}
	}

	return Logger
}


// func InitNewLogger() *logrus.Logger {
// 	logger := logrus.New()

// 	level := logrus.InfoLevel
// 	if os.Getenv("DEBUG") == "true" {
// 		level = logrus.DebugLevel
// 	}
// 	logger.SetLevel(level)

// 	logger.SetFormatter(&logrus.TextFormatter{
// 		FullTimestamp:   true,
// 		TimestampFormat: "2006-01-02 15:04:05",
// 		ForceColors:     true,
// 	})

	// logFile := os.Getenv("LOG_FILE")
	// if logFile != "" {
	// 	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// 	if err == nil {
	// 		multiWriter := io.MultiWriter(os.Stdout, file)
	// 		logger.SetOutput(multiWriter)
	// 	} else {
	// 		logger.Warnf("Не вдалося відкрити файл логів %s: %v. Використовую стандартний stderr", logFile, err)
	// 	}
	// }

// 	return logger
// }
