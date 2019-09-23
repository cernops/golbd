package main_test

import (
	"testing"
	"time"
)

func TestTimeOfLastEvaluation(t *testing.T) {

	c := getTestCluster("test01.cern.ch")

	c.Time_of_last_evaluation = time.Now().Add(time.Duration(-c.Parameters.Polling_interval+2) * time.Second)
	if c.Time_to_refresh() {
		t.Errorf("e.Time_of_last_evaluation: got\n%v\nexpected\n%v", true, false)
	}
	c.Time_of_last_evaluation = time.Now().Add(time.Duration(-c.Parameters.Polling_interval-2) * time.Second)
	if !c.Time_to_refresh() {
		t.Errorf("e.Time_of_last_evaluation: got\n%v\nexpected\n%v", false, true)
	}
}
