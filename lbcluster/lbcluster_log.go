package lbcluster

import (
	"fmt"
	"log/syslog"
	"os"
	"sync"
	"strings"
	"time"
)
type Log struct {
	Writer     syslog.Writer
	Syslog     bool
	Stdout     bool
	Debugflag  bool
	TofilePath string
	logMu      sync.Mutex
}

type Logger interface {
	Info(s string) error
	Warning(s string) error
	Debug(s string) error
	Error(s string) error
}

func (self *LBCluster) Write_to_log(level string, msg string) error {

	my_message := "cluster: " + self.Cluster_name + " " + msg

	if level == "INFO" {
		self.Slog.Info(my_message)
	} else if level == "DEBUG" {
		self.Slog.Debug(my_message)
	} else if level == "WARNING" {
		self.Slog.Warning(my_message)
	} else if level == "ERROR" {
		self.Slog.Error(my_message)
	} else {
		self.Slog.Error("LEVEL " + level + " NOT UNDERSTOOD, ASSUMING ERROR " + my_message)
	}

	//We send the logs to timber, and in that one, it is quite easy to filter by cluster. We don't need the dedicated logs anymore
	/*
		f, err := os.OpenFile(self.Per_cluster_filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer f.Close()
		tag := "lbd"
		nl := ""
		if !strings.HasSuffix(msg, "\n") {
			nl = "\n"
		}
		timestamp := time.Now().Format(time.Stamp)
		_, err = fmt.Fprintf(f, "%s %s[%d]: cluster: %s %s: %s%s",
			timestamp,
			tag, os.Getpid(), self.Cluster_name, level, msg, nl)
		return err */
	return nil
}


func (l *Log) Info(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Info(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("INFO: " + s)
	}
	return err

}

func (l *Log) Warning(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Warning(s)
	}
	if l.Stdout || (l.TofilePath != "") {
		err = l.writefilestd("WARNING: " + s)
	}
	return err

}

func (l *Log) Debug(s string) error {
	var err error
	if l.Debugflag {
		if l.Syslog {
			err = l.Writer.Debug(s)
		}
		if l.Stdout || (l.TofilePath != "") {
			err = l.writefilestd("DEBUG: " + s)
		}
	}
	return err

}

func (l *Log) Error(s string) error {
	var err error
	if l.Syslog {
		err = l.Writer.Err(s)
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
	timestamp := time.Now().Format(time.Stamp)
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
