package lbcluster

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

//Log struct for the log
type Log struct {
	SyslogWriter *syslog.Writer
	Stdout       bool
	Debugflag    bool
	TofilePath   string
	logMu        sync.Mutex
}

//Logger struct for the Logger interface
type Logger interface {
	Info(s string) error
	Warning(s string) error
	Debug(s string) error
	Error(s string) error
}

//Write_to_log put something in the log file
func (lbc *LBCluster) Write_to_log(level string, msg string) error {

	myMessage := "cluster: " + lbc.Cluster_name + " " + msg

	if level == "INFO" {
		lbc.Slog.Info(myMessage)
	} else if level == "DEBUG" {
		lbc.Slog.Debug(myMessage)
	} else if level == "WARNING" {
		lbc.Slog.Warning(myMessage)
	} else if level == "ERROR" {
		lbc.Slog.Error(myMessage)
	} else {
		lbc.Slog.Error("LEVEL " + level + " NOT UNDERSTOOD, ASSUMING ERROR " + myMessage)
	}

	return nil
}

//Info write as Info
func (l *Log) Info(s string) error {
	var err error
	if l.SyslogWriter != nil {
		err = l.SyslogWriter.Info(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("INFO: " + s)
	}
	return err

}

//Warning write as Warning
func (l *Log) Warning(s string) error {
	var err error
	if l.SyslogWriter != nil {
		err = l.SyslogWriter.Warning(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("WARNING: " + s)
	}
	return err

}

//Debug write as Debug
func (l *Log) Debug(s string) error {
	var err error
	if l.Debugflag {
		if l.SyslogWriter != nil {
			err = l.SyslogWriter.Debug(s)
		}
		if l.Stdout || (l.TofilePath != "") {
			err = l.writefilestd("DEBUG: " + s)
		}
	}
	return err

}

//Error write as Error
func (l *Log) Error(s string) error {
	var err error
	if l.SyslogWriter != nil {
		err = l.SyslogWriter.Err(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("ERROR: " + s)
	}
	return err

}

func (l *Log) writefilestd(s string) error {
	var err error
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
	if l.Stdout {
		_, err = fmt.Printf(msg)
	}
	if l.TofilePath != "" {
		f, err := os.OpenFile(l.TofilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = fmt.Fprintf(f, msg)
	}
	return err
}
