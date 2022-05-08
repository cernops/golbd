package main_test

import (
	"io/ioutil"
	"lb-experts/golbd/lbcluster"
	"os"
	"strings"
	"testing"
)

const testLogFilePath = "../tests/sample.log"

func TestLBClusterLoggerForInitFailure(t *testing.T)  {
	_, err := lbcluster.NewLoggerFactory("")
	if err==nil {
		t.Errorf("expected error not thrown")
	}
}

func TestLBClusterLoggerForInitSuccess(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	if logger==nil{
		t.Fail()
		t.Errorf("logger instance is nil")
	}
}

func TestLBClusterLoggerForDebugDisabled(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	logger.Debug("sample info")
	if isLogPresentInFile(t, testLogFilePath, "sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "sample info")
	}
}

func TestLBClusterLoggerForDebugEnabled(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	logger.EnableDebugMode()
	logger.Debug("sample info")
	if !isLogPresentInFile(t, testLogFilePath, "sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "sample info")
	}
}

func TestLBClusterLoggerForInfo(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	logger.Info("sample info")
	if !isLogPresentInFile(t, testLogFilePath, "INFO: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "INFO: sample info")
	}
}

func TestLBClusterLoggerForWarning(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	logger.Warning("sample info")
	if !isLogPresentInFile(t, testLogFilePath, "WARNING: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "WARNING: sample info")
	}
}

func TestLBClusterLoggerForError(t *testing.T)  {
	defer deleteFile(t, testLogFilePath)
	logger, err := lbcluster.NewLoggerFactory(testLogFilePath)
	if err!=nil {
		t.Fail()
		t.Errorf("unexpected error thrown. error: %v",err)
	}
	logger.Error("sample info")
	if !isLogPresentInFile(t, testLogFilePath, "ERROR: sample info") {
		t.Fail()
		t.Errorf("log file does not contain the expected debug info. Expected Info: %s", "ERROR: sample info")
	}
}

func deleteFile(t *testing.T, filePath string){
	err := os.Remove(filePath)
	if err!=nil{
		 t.Errorf("error whil deleting log file error: %v", err)
	}
}

func isLogPresentInFile(t *testing.T, filePath string, stringToCheck string) bool{
	data, err := ioutil.ReadFile(filePath)
	if err!=nil {
		t.Errorf("error while reading log file. error: %v", err)
		return false
	}
	return strings.Contains(string(data),stringToCheck)
}


