package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var ErrLogLevel = errors.New("unrecognized log_level")

type Conf struct {
	Level string `toml:"level"`
}

type Logger struct {
	level  int
	writer io.Writer
	mu     *sync.Mutex
}

func NewLogger(level string, writer io.Writer) *Logger {
	switch strings.ToUpper(level) {
	case "ERROR":
		return &Logger{level: LevelError, mu: &sync.Mutex{}, writer: writer}
	case "WARN":
		return &Logger{level: LevelWarn, mu: &sync.Mutex{}, writer: writer}
	case "INFO":
		return &Logger{level: LevelInfo, mu: &sync.Mutex{}, writer: writer}
	case "DEBUG":
		return &Logger{level: LevelDebug, mu: &sync.Mutex{}, writer: writer}
	default:
		fmt.Fprintln(os.Stderr, "unrecognized log_level")
		os.Exit(1)
	}
	return nil
}

func (l *Logger) printf(format string, a ...interface{}) {
	l.mu.Lock()
	_, err := fmt.Fprintf(l.writer, format, a...)
	l.mu.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: Fprintf : %v", err)
		os.Exit(1)
	}
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.printf("Fatal:"+format, a)
	os.Exit(1)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	if l.level >= LevelError {
		l.printf("ERROR:"+format, a...)
	}
}

func (l *Logger) Warningf(format string, a ...interface{}) {
	if l.level >= LevelWarn {
		l.printf("WARN:"+format, a...)
	}
}

func (l *Logger) Infof(format string, a ...interface{}) {
	if l.level >= LevelInfo {
		l.printf("INFO:"+format, a...)
	}
}

func (l *Logger) Debugf(format string, a ...interface{}) {
	if l.level >= LevelDebug {
		l.printf("DEBUG:"+format, a...)
	}
}
