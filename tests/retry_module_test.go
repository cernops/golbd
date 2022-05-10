package main_test

import (
	"fmt"
	"testing"
	"time"

	"lb-experts/golbd/lbcluster"
)

func TestRetryWithNoErrorsShouldExitAfterFirstAttempt(t *testing.T) {
	lg := &lbcluster.Log{Stdout: true, Debugflag: false}
	currentTime := time.Now()
	retryModule := lbcluster.NewRetryModule(10*time.Second, lg)
	err := retryModule.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	if time.Now().Sub(currentTime) > 1*time.Second {
		t.Fail()
		t.Errorf("should quit after first try")
	}
}

func TestRetryWithErrorShouldQuitAfterMultipleAttempts(t *testing.T) {
	lg := &lbcluster.Log{Stdout: true, Debugflag: false}
	currentTime := time.Now()
	counter := 0
	retryModule := lbcluster.NewRetryModule(1*time.Second, lg)
	err := retryModule.Execute(func() error {
		if counter == 4 {
			return nil
		}
		counter += 1
		return fmt.Errorf("sample error")
	})
	if err != nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	if time.Now().Sub(currentTime) > 12*time.Second {
		t.Fail()
		t.Errorf("should quit after expected: %v, actual:%v", "11 sec", time.Now().Sub(currentTime))
	}
}

func TestRetryWithErrorShouldQuitAfterMaxCount(t *testing.T) {
	lg := &lbcluster.Log{Stdout: true, Debugflag: false}
	currentTime := time.Now()
	counter := 0
	retryModule := lbcluster.NewRetryModule(1*time.Second, lg)
	err := retryModule.SetMaxCount(3)
	if err != nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	err = retryModule.Execute(func() error {
		if counter == 4 {
			return nil
		}
		counter += 1
		return fmt.Errorf("sample error")
	})
	if err == nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	if time.Now().Sub(currentTime) > 4*time.Second {
		t.Fail()
		t.Errorf("should quit after expected: %v, actual:%v", "3 sec", time.Now().Sub(currentTime))
	}
}

func TestRetryWithErrorShouldQuitAfterMaxDuration(t *testing.T) {
	lg := &lbcluster.Log{Stdout: true, Debugflag: false}
	currentTime := time.Now()
	counter := 0
	retryModule := lbcluster.NewRetryModule(1*time.Second, lg)
	err := retryModule.SetMaxDuration(4 * time.Second)
	if err != nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	err = retryModule.Execute(func() error {
		if counter == 4 {
			return nil
		}
		counter += 1
		return fmt.Errorf("sample error")
	})
	if err == nil {
		t.Fail()
		t.Errorf("error should be nil")
	}
	if time.Now().Sub(currentTime) > 5*time.Second {
		t.Fail()
		t.Errorf("should quit after expected: %v, actual:%v", "3 sec", time.Now().Sub(currentTime))
	}
}
