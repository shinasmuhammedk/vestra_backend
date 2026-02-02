package logging

import (
	"log"
	"os"
)

// Debug logger writes debug-level logs to debug.log
var Debug *log.Logger

// Error logger writes error-level logs to error.log
var Error *log.Logger

// InitLogger initializes debug and error loggers
func InitLogger() {

	// Open or create debug.log file
	debugFile, err := os.OpenFile(
		"log/debug.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatalf("failed to open debug.log: %v", err)
	}

	// Open or create error.log file
	errorFile, err := os.OpenFile(
		"log/error.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatalf("failed to open error.log: %v", err)
	}

	// Initialize debug logger
	Debug = log.New(
		debugFile,
		"[DEBUG] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	// Initialize error logger
	Error = log.New(
		errorFile,
		"[ERROR] ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
}
