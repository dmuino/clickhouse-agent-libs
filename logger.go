package chagent

import (
	"encoding/json"
	"fmt"
	"github.com/mattn/go-isatty"
	"net"
	"os"
	"time"
)

type Level int

const (
	ERROR Level = iota
	WARNING
	INFO
	DEBUG
)

var socketForOtel net.Conn

type Logger struct {
	name      string
	level     Level
	colorizer *Colorizer
}

var loggerMap = make(map[string]*Logger)

func GetLogger(name string) *Logger {
	if logger, ok := loggerMap[name]; ok {
		return logger
	}
	logger := NewLogger(name, INFO)
	loggerMap[name] = logger
	return logger
}

func createSocket() (conn net.Conn, err error) {
	delay := 250 * time.Millisecond
	for attempt := 1; attempt < 5; attempt++ {
		conn, err = net.Dial("tcp", "localhost:1552")
		if err == nil {
			return
		}
		fmt.Printf("WARN: Unable to connect to localhost:1552 %v\n", err)
		if attempt < 4 {
			delay *= 2
			fmt.Printf("INFO: Will retry connection to otel-collector after %v\b", delay)
			time.Sleep(delay)
		}
	}
	return
}

func isTerminal() bool {
	return isatty.IsTerminal(os.Stderr.Fd())
}

func NewLogger(name string, level Level) *Logger {
	var err error
	if !IsLocalEnvironment() && socketForOtel == nil {
		socketForOtel, err = createSocket()
	}

	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		socketForOtel = nil
	}
	return &Logger{level: level,
		colorizer: NewColorizer(isTerminal()), name: name}
}

type LogPayload struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	Logger  string `json:"logger"`
	Logts   int64  `json:"logts"`
}

func (l *Logger) sendToCollector(level string, msg string) {
	if socketForOtel == nil {
		return
	}

	payload := LogPayload{Message: msg, Level: level, Logger: l.name, Logts: time.Now().UnixNano()}
	encoder := json.NewEncoder(socketForOtel)
	err := encoder.Encode(payload)
	if err != nil {
		fmt.Printf("Error encoding payload. Reconnecting: %v\n", err)
		_ = socketForOtel.Close()
		socketForOtel, _ = createSocket()
	}
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) log(format string, tag string, colorFun func(string, string) string, args ...interface{}) string {
	userMsg := fmt.Sprintf(format, args...)

	taggedMsg := fmt.Sprintf("%s %s", tag, userMsg)
	msg := colorFun(taggedMsg, ColorsReset)
	// add a timestamp if writing to the terminal, otherwise assume it's provided by the platform
	// like journald
	dateTime := ""
	if isTerminal() {
		dateTime = fmt.Sprintf("%s ", time.Now().Format("2006-01-02T15:04:05.000"))
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s%-20s %s\n", dateTime, l.name, msg)
	return userMsg
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level >= DEBUG {
		userMsg := l.log(format, "", l.colorizer.Cyan, args...)
		l.sendToCollector("debug", userMsg)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= INFO {
		userMsg := l.log(format, ColorsCheck, l.colorizer.Green, args...)
		l.sendToCollector("info", userMsg)
	}
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	if l.level >= WARNING {
		userMsg := l.log(format, "[!]", l.colorizer.Yellow, args...)
		l.sendToCollector("warning", userMsg)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= ERROR {
		userMsg := l.log(format, ColorsCross, l.colorizer.Red, args...)
		l.sendToCollector("error", userMsg)
	}
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Errorf(format, args...)
	os.Exit(1)
}

func (l *Logger) CheckErr(err error) {
	if err != nil {
		l.Fatalf("%v", err)
	}
}
