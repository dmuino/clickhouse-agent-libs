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

type Logger struct {
	level     Level
	colorizer *Colorizer
	socket    net.Conn
}

// global logger
var logger = NewLogger(DEBUG)

func GetLogger() *Logger {
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

func NewLogger(level Level) *Logger {
	var socket net.Conn
	var err error
	if !isLocalEnv() {
		socket, err = createSocket()
	}

	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		socket = nil
	}
	return &Logger{level: level,
		colorizer: NewColorizer(isatty.IsTerminal(os.Stderr.Fd())), socket: socket}
}

type LogPayload struct {
	Message string `json:"message"`
	Level   string `json:"level"`
	Logger  string `json:"logger"`
	Logts   int64  `json:"logts"`
}

func (l *Logger) sendToCollector(level string, msg string) {
	if l.socket == nil {
		return
	}

	payload := LogPayload{Message: msg, Level: level, Logger: "agent", Logts: time.Now().UnixNano()}
	encoder := json.NewEncoder(l.socket)
	err := encoder.Encode(payload)
	if err != nil {
		fmt.Printf("Error encoding payload. Reconnecting: %v\n", err)
		_ = l.socket.Close()
		l.socket, _ = createSocket()
	}
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level >= DEBUG {
		formattedMsg := fmt.Sprintf(format, args...)
		msg := l.colorizer.Cyan(formattedMsg, ColorsReset)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
		l.sendToCollector("debug", formattedMsg)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= INFO {
		userMsg := fmt.Sprintf(format, args...)
		infoMsg := fmt.Sprintf("%s %s", ColorsCheck, userMsg)
		msg := l.colorizer.Green(infoMsg, ColorsReset)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
		l.sendToCollector("info", userMsg)
	}
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	if l.level >= WARNING {
		userMsg := fmt.Sprintf(format, args...)
		taggedMsg := fmt.Sprintf("[!] %s", userMsg)
		msg := l.colorizer.Yellow(taggedMsg, ColorsReset)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
		l.sendToCollector("warning", userMsg)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= ERROR {
		userMsg := fmt.Sprintf(format, args...)
		taggedMsg := fmt.Sprintf("%s %s", ColorsCross, userMsg)
		msg := l.colorizer.Red(taggedMsg, ColorsReset)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", msg)
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
