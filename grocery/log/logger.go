package log

import (
	"fmt"
	"log"
	"os"
	"reflect"
)

var (
	colors = map[string]string{
		"DEF":   "\033[00;38;5;240m",
		"FAIL":  "\033[48;5;160;38;5;230m",
		"OK":    "\033[01;38;5;64m",
		"WARN":  "\033[00;38;5;136m",
		"INFO":  "\033[01;38;5;33m",
		"CLEAR": "\033[0m",
	}
)

type (
	Logger struct {
		*log.Logger
	}
)

func NewLogger(prefix string, flags int) *Logger {
	if flags == 0 {
		flags = log.Lmsgprefix
	}

	return &Logger{
		Logger: log.New(os.Stdout, prefix, flags),
	}
}

func (l *Logger) Print(msg string, args ...interface{}) {
	color := "DEF"

	if len(args) > 0 {
		if reflect.TypeOf(args[len(args)-1]).Kind() == reflect.String {
			if _, ok := colors[args[len(args)-1].(string)]; ok {
				color = args[len(args)-1].(string)
				args = args[:len(args)-1]
			}
		}
	}

	os.Stdout.WriteString(fmt.Sprintf("%s%s%s%s\n",
		colors[color],
		l.Logger.Prefix(),
		colors["CLEAR"],
		fmt.Sprintf(msg, args...),
	))
}
