package logging

import (
	"log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Debug *log.Logger
var Error *log.Logger

func InitLogger() {
	debugWriter := &lumberjack.Logger{
		Filename:   "log/debug.log",
		MaxSize:    10,   // 10 MB max
		MaxBackups: 3,    // keep 3 old files
		MaxAge:     7,    // delete after 7 days
		Compress:   true, // gzip old files
	}

	errorWriter := &lumberjack.Logger{
		Filename:   "log/error.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	Debug = log.New(debugWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}