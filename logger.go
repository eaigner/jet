package jet

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	Logger     *log.Logger
	QueryColor string
	TxnColor   string
	ResetColor string
	parts      []string
}

func NewLogger(file *os.File) *Logger {
	if file == nil {
		file = os.Stdout
	}
	return &Logger{
		Logger:     log.New(file, "SQL: ", log.LstdFlags),
		QueryColor: "\x1b[35m",
		TxnColor:   "\x1b[36m",
		ResetColor: "\x1b[0m",
		parts:      []string{},
	}
}

func (l *Logger) Queryf(format string, args ...interface{}) *Logger {
	l.parts = append(l.parts, fmt.Sprintf("%s%s%s", l.QueryColor, fmt.Sprintf(format, args...), l.ResetColor))
	return l
}

func (l *Logger) Txnf(format string, args ...interface{}) *Logger {
	l.parts = append(l.parts, fmt.Sprintf("%s%s%s", l.TxnColor, fmt.Sprintf(format, args...), l.ResetColor))
	return l
}

func (l *Logger) Println() {
	l.Logger.Println(strings.Join(l.parts, ""))
	l.parts = []string{}
}
