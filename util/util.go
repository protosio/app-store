package util

import "github.com/sirupsen/logrus"

// Config is a struct that is used to share config params all over the code
type Config struct {
	Port   int
	DBHost string
	DBName string
	DBUser string
	DBPass string
	DBPort int
}

// Log is the global logger
var log = logrus.New()

var config = Config{
	DBName: "installers",
	DBUser: "root",
	DBPort: 26257,
}

// SetLogLevel sets the log level for the application
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	return log
}

// GetConfig returns the global config struct
func GetConfig() *Config {
	return &config
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
