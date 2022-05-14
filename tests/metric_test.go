package main_test

import (
	"lb-experts/golbd/metric"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMetricReadWriteRecord(t *testing.T) {
	hostName, _ := os.Hostname()
	logic := metric.NewLogic("", hostName)
	curTime := time.Now()
	property := metric.Property{
		RoundTripStartTime: curTime,
		RoundTripEndTime:   curTime.Add(5 * time.Second),
		RoundTripDuration:  5 * time.Second,
	}
	err := logic.WriteRecord(property)
	if err != nil {
		t.Fail()
		t.Errorf("error while recoding metric. error:%v", err)
		return
	}
	hostMetric, err := logic.ReadHostMetric()
	if err != nil {
		t.Fail()
		t.Errorf("error while reading metric.error:%v", err)
		return
	}

	if hostMetric.PropertyList == nil || len(hostMetric.PropertyList) == 0 {
		t.Fail()
		t.Errorf("property list is empty. expected length :%d", 1)
		return
	}
	expectedStarttime := property.RoundTripStartTime.Format(time.RFC3339)
	expectedEndtime := property.RoundTripEndTime.Format(time.RFC3339)
	actualStartTime := hostMetric.PropertyList[0].RoundTripStartTime.Format(time.RFC3339)
	actualEndTime := hostMetric.PropertyList[0].RoundTripEndTime.Format(time.RFC3339)
	if !strings.EqualFold(expectedStarttime, actualStartTime) {
		t.Fail()
		t.Errorf("start time value mismatch expected: %v, actual:%v", expectedStarttime, actualStartTime)
	}
	if !strings.EqualFold(expectedEndtime, actualEndTime) {
		t.Fail()
		t.Errorf("end time value mismatch expected: %v, actual:%v", expectedEndtime, actualEndTime)
	}
	if hostMetric.PropertyList[0].RoundTripDuration != property.RoundTripDuration {
		t.Fail()
		t.Errorf("duration value mismatch expected: %v, actual:%v", property.RoundTripDuration, hostMetric.PropertyList[0].RoundTripDuration)
	}
	err = os.Remove(logic.GetFilePath())
	if err != nil {
		t.Fail()
		t.Errorf("error deleting file.error %v", err)
	}
}
