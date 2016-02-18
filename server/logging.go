package server

import (
	"log"
	"os"
)

type HasLogger struct {
	logger *log.Logger
}

func (h *HasLogger) InitLog(filename string) error {
	if filename == "" {
		// Use a simple format for stdout: no prefix, date or time
		h.logger = log.New(os.Stdout, "", 0)
		return nil
	}

	fileOptions := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile(filename, fileOptions, 0666)
	if err != nil {
		return err
	}

	h.logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	return nil
}

func (h HasLogger) DefaultLogger() log.Logger {
	return *h.logger
}

// Output an ordinary ("info") message to the log.
func (h HasLogger) Log(message ...interface{}) {
	h.logger.Println("INFO:  ", message)
}

// Output a warning to the log.
func (h HasLogger) Warn(message ...interface{}) {
	h.logger.Println("WARN:  ", message)
}

// Output an error to the log.
func (h HasLogger) Error(message ...interface{}) {
	h.logger.Println("ERROR: ", message)
}

// Output an error to the log and terminate the process.
func (h HasLogger) FatalError(message ...interface{}) {
	h.logger.Fatal("FATAL: ", message)
}
