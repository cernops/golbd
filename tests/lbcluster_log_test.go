package main_test

import (
	"fmt"
	"io/ioutil"
	"lb-experts/golbd/logger"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	testLogDirPath = "../tests"
	testFileName   = "sample"
)

func getLogFilePath() string {
	return fmt.Sprintf("%s/%s.%s", testLogDirPath, testFileName, "log")
}
func TestLBClusterLoggerForInitFailure(t *testing.T) {
	_, err := logger.NewLoggerFactory("")
	if err == nil {
		t.Errorf("expected error not thrown")
	}
}

func TestLBClusterLoggerForInitSuccess(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	if logger == nil {
		t.Fail()
		t.Errorf("logger instance is nil")
	}
}

func TestLBClusterLoggerForSnapshot(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	if logger == nil {
		t.Fail()
		t.Errorf("logger instance is nil")
	}
	logger.StartSnapshot(1 * time.Minute)
	if !strings.Contains(logger.GetLogFilePath(),
		fmt.Sprintf("%v_%v_%v-%v_%v", time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute())) {
		t.Fail()
		t.Errorf("error while setting snapshot")
	}
}

func TestLBClusterLoggerForNewSnapshot(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	if logger == nil {
		t.Fail()
		t.Errorf("logger instance is nil")
	}
	curTime := time.Now()
	logger.StartSnapshot(5 * time.Second)
	time.Sleep(5 * time.Second)
	curTime = curTime.Add(5 * time.Second)
	logger.Info("sample info")
	if !strings.Contains(logger.GetLogFilePath(),
		fmt.Sprintf("%v_%v_%v-%v_%v_%v", curTime.Year(), curTime.Month(), curTime.Day(), curTime.Hour(), curTime.Minute(), curTime.Second())) {
		t.Fail()
		t.Errorf("error while setting snapshot")
	}
}

func TestLBClusterLoggerForDebugDisabled(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	logger.Debug("sample info")
	if isLogPresentInFile(t, getLogFilePath(), "sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "sample info")
	}
}

func TestLBClusterLoggerForDebugEnabled(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	logger.EnableDebugMode()
	logger.Debug("sample info")
	if !isLogPresentInFile(t, getLogFilePath(), "sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "sample info")
	}
}

func TestLBClusterLoggerForInfo(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	logger.Info("sample info")
	if !isLogPresentInFile(t, getLogFilePath(), "INFO: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "INFO: sample info")
	}
}

func TestLBClusterLoggerForWarning(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	logger.Warning("sample info")
	if !isLogPresentInFile(t, getLogFilePath(), "WARNING: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "WARNING: sample info")
	}
}

func TestLBClusterLoggerForError(t *testing.T) {
	defer deleteFile(t)
	logger, err := logger.NewLoggerFactory(getLogFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v", err)
	}
	logger.Error("sample info")
	if !isLogPresentInFile(t, getLogFilePath(), "ERROR: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "ERROR: sample info")
	}
}

func deleteFile(t *testing.T) {
	files, err := ioutil.ReadDir(testLogDirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), testFileName) {
			err = os.Remove(f.Name())
			if err != nil {
				t.Errorf("error whil deleting log file error: %v", err)
			}
		}
	}
}

func isLogPresentInFile(t *testing.T, filePath string, stringToCheck string) bool {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Errorf("error while reading log file. error: %v", err)
		return false
	}
	return strings.Contains(string(data), stringToCheck)
}
