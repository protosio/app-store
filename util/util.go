package util

import "github.com/sirupsen/logrus"

// Log is the global logger
var log = logrus.New()

// SetLogLevel sets the log level for the application
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	return log
}

// StringInSlice checks if provided string is in provided string list
func StringInSlice(a string, list []string) (bool, int) {
	for i, b := range list {
		if b == a {
			return true, i
		}
	}
	return false, 0
}
