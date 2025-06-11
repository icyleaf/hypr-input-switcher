package main

import (
	"log"
	"os"

	"hypr-input-switcher/internal/app"
	"hypr-input-switcher/pkg/logger"
)

func main() {
	// Set up logging
	logFile, err := os.OpenFile("hypr-input-switcher.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logFile.Close()

	logger.SetOutput(logFile)

	// Initialize and run the application
	application := app.NewApplication()
	if err := application.Run(); err != nil {
		logger.Error("Application failed to run", err)
		os.Exit(1)
	}
}
