package util

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/sirupsen/logrus"
)

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

// String2SHA1 converts a string to a SHA1 hash, formatted as a hex string
func String2SHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
