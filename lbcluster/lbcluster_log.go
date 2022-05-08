package lbcluster

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultLbdTag   = "lbd"
	logLevelInfo    = "INFO"
	logLevelDebug   = "DEBUG"
	logLevelWarning = "WARNING"
	logLevelError   = "ERROR"
)

func NewLoggerFactory(logFilePath string) (Logger, error) {
	log, err := syslog.New(syslog.LOG_NOTICE, DefaultLbdTag)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(logFilePath, "") {
		return nil, fmt.Errorf("empty log file path")
	}
	_, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return nil, err
	}
	return &Log{
			logWriter: log,
			filePath:  logFilePath},
		nil
}

//Log struct for the log
type Log struct {
	logWriter        *syslog.Writer
	shouldWriteToSTD bool
	isDebugAllowed   bool
	filePath         string
	logMu            sync.Mutex
}

//Logger struct for the Logger interface
type Logger interface {
	EnableDebugMode()
	EnableWriteToSTd()
	Info(s string)
	Warning(s string)
	Debug(s string)
	Error(s string)
}

func (l *Log) EnableDebugMode() {
	l.isDebugAllowed = true
}

func (l *Log) EnableWriteToSTd() {
	l.shouldWriteToSTD = true
}

//Info write as Info
func (l *Log) Info(s string) {
	if l.logWriter != nil {
		_ = l.logWriter.Info(s)
	}
	if l.shouldWriteToSTD || (l.filePath != "") {
		l.write(fmt.Sprintf("%s: %s", logLevelInfo, s))
	}
}

//Warning write as Warning
func (l *Log) Warning(s string) {
	if l.logWriter != nil {
		_ = l.logWriter.Warning(s)
	}
	if l.shouldWriteToSTD || (l.filePath != "") {
		l.write(fmt.Sprintf("%s: %s", logLevelWarning, s))
	}
}

//Debug write as Debug
func (l *Log) Debug(s string) {
	if l.isDebugAllowed {
		if l.logWriter != nil {
			_ = l.logWriter.Debug(s)
		}
		if l.shouldWriteToSTD || (l.filePath != "") {
			l.write(fmt.Sprintf("%s: %s", logLevelDebug, s))
		}
	}
}

//Error write as Error
func (l *Log) Error(s string) {
	if l.logWriter != nil {
		_ = l.logWriter.Err(s)
	}
	if l.shouldWriteToSTD || (l.filePath != "") {
		l.write(fmt.Sprintf("%s: %s", logLevelError, s))
	}
}

func (l *Log) write(s string) {
	tag := "lbd"
	nl := ""
	if !strings.HasSuffix(s, "\n") {
		nl = "\n"
	}
	timestamp := time.Now().Format(time.StampMilli)
	msg := fmt.Sprintf("%s %s[%d]: %s%s",
		timestamp,
		tag, os.Getpid(), s, nl)
	l.logMu.Lock()
	defer l.logMu.Unlock()
	if l.shouldWriteToSTD {
		fmt.Printf(msg)
	}

	f, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error while opening the log file. error: %v", err)
		return
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, msg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error while writing to the log file. error: %v", err)
	}
}
