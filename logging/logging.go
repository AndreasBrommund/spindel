package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	OFF     LogLevel = 0
	FATAL   LogLevel = 1
	ERROR   LogLevel = 2
	WARNING LogLevel = 3
	INFO    LogLevel = 4
	DEBUG   LogLevel = 5
	ALL     LogLevel = 6
)

func (l LogLevel) String() string {
	switch l {
	case OFF:
		return "OFF"
	case FATAL:
		return "FATAL"
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case ALL:
		return "ALL"
	}
	return "UNKNOWN"
}

type Logging struct {
	CurrentLogLevel LogLevel
	LogToFile       bool
	File            string
}

func NewLogger(level LogLevel, path string) *Logging {
	logToFile := false

	if path != "" {
		logToFile = true
	}

	return &Logging{
		CurrentLogLevel: level,
		File:            path,
		LogToFile:       logToFile,
	}
}

func (l *Logging) LogFatal(msg string, err error) {
	l.log(msg, FATAL, err)
	os.Exit(1)
}

func (l *Logging) LogError(msg string, err error) {
	l.log(msg, ERROR, err)
}

func (l *Logging) LogWarning(msg string, err error) {
	l.log(msg, WARNING, err)
}

func (l *Logging) LogInfo(msg string) {
	l.log(msg, INFO, nil)
}

func (l *Logging) LogDebug(msg string) {
	l.log(msg, DEBUG, nil)
}

func (l *Logging) log(msg string, level LogLevel, err error) {
	var logString string

	delimiter := "########################################"

	msg = strings.Trim(msg, "\n")

	if err != nil {
		logString = fmt.Sprintf("%s (%s) \n%s\n%s\n", msg, level.String(), err.Error(), delimiter)
	} else {
		logString = fmt.Sprintf("%s (%s)\n%s\n", msg, level.String(), delimiter)
	}

	if l.CurrentLogLevel >= level {
		log.Print(logString)
		if l.LogToFile {
			f, err := os.OpenFile(l.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("Could not create file: %s \n%s", l.File, err.Error())
			}
			if _, err := f.Write([]byte(logString)); err != nil {
				log.Fatalf("Could not write to file: %s \n%s", l.File, err.Error())
			}
			if err := f.Close(); err != nil {
				log.Fatalf("Could not close file %s \n%s", l.File, err.Error())
			}
		}
	}
}
