package jet

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	Logger      *log.Logger
	ActionColor string
	QueryColor  string
	ResetColor  string
}

func NewLogger(file *os.File) *Logger {
	if file == nil {
		file = os.Stdout
	}
	return &Logger{
		Logger:      log.New(file, "SQL: ", log.LstdFlags),
		ActionColor: "\x1b[36m",
		QueryColor:  "\x1b[35m",
		ResetColor:  "\x1b[0m",
	}
}

func (l *Logger) Actionf(format string, args ...interface{}) {
	l.Logger.Printf("%s%s%s", l.ActionColor, fmt.Sprintf(format, args...), l.ResetColor)
}

func (l *Logger) Queryf(format string, args ...interface{}) {
	l.Logger.Printf("%s%s%s", l.QueryColor, fmt.Sprintf(format, args...), l.ResetColor)
}
