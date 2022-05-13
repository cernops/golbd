package logger

import (
	"fmt"
	"log/syslog"
	"os"
	"regexp"
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

//Log struct for the log
type Log struct {
	logWriter         *syslog.Writer
	shouldWriteToSTD  bool
	isDebugAllowed    bool
	filePath          string
	logMu             sync.Mutex
	logStartTime      time.Time
	isSnapShotEnabled bool
	snapShotCycleTime time.Duration
	logFileBasePath   string
	logFileExtension  string
}

//Logger struct for the Logger interface
type Logger interface {
	EnableDebugMode()
	EnableWriteToSTd()
	StartSnapshot(d time.Duration)
	GetLogFilePath() string
	Info(s string)
	Warning(s string)
	Debug(s string)
	Error(s string)
}

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
	basePath, extension, err := getLogFilePathAndExtension(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("error while validating log file path. error: %v", err)
	}

	if !isLogFilePathValid(basePath, extension) {
		return nil, fmt.Errorf("invalid log file path. log path: %s", logFilePath)
	}
	return &Log{
			logWriter:        log,
			logStartTime:     time.Now(),
			filePath:         logFilePath,
			logFileBasePath:  basePath,
			logFileExtension: extension,
		},
		nil
}

func isLogFilePathValid(basePath, extension string) bool {
	return !strings.EqualFold(basePath, "") && !strings.EqualFold(extension, "")
}

func getLogFilePathAndExtension(logFilePath string) (string, string, error) {
	matcher := regexp.MustCompile(`(.*)\.(.*)`)
	matchGroups := matcher.FindAllStringSubmatch(logFilePath, -1)
	if len(matchGroups) == 0 || len(matchGroups[0]) < 3 {
		return "", "", fmt.Errorf("log file path is not in the right format. path: %s", logFilePath)
	}
	return matchGroups[0][1], matchGroups[0][2], nil
}

func (l *Log) EnableDebugMode() {
	l.isDebugAllowed = true
}

func (l *Log) EnableWriteToSTd() {
	l.shouldWriteToSTD = true
}

func (l *Log) StartSnapshot(d time.Duration) {
	if !l.isSnapShotEnabled {
		l.isSnapShotEnabled = true
		l.snapShotCycleTime = d
		l.startSnapShot()
	}
}

func (l *Log) GetLogFilePath() string {
	return l.filePath
}
func (l *Log) startSnapShot() {
	l.logStartTime = time.Now()
	l.filePath = fmt.Sprintf("%s_%s.%s", l.logFileBasePath, l.getFileNameTimeSuffix(), l.logFileExtension)
}

func (l *Log) getFileNameTimeSuffix() string {
	return fmt.Sprintf("%v_%v_%v-%v_%v_%v", l.logStartTime.Year(), l.logStartTime.Month(), l.logStartTime.Day(), l.logStartTime.Hour(), l.logStartTime.Minute(), l.logStartTime.Second())
}

func (l *Log) shouldStartNewSnapshot() bool {
	if l.isSnapShotEnabled {
		return time.Now().Sub(l.logStartTime) >= l.snapShotCycleTime
	}
	return false
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
	if l.shouldWriteToSTD {
		fmt.Printf(msg)
	}
	l.logMu.Lock()
	defer l.logMu.Unlock()

	if l.shouldStartNewSnapshot() {
		l.startSnapShot()
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
