package metric

import (
	"fmt"
	"time"
)

type HostMetric struct {
	Name         string     `json:"name" csv:"name" rw:"r"`
	PropertyList []Property `json:"property_list" csv:"property_list" rw:"r"`
}

type Property struct {
	RoundTripStartTime time.Time     `json:"start_time" csv:"start_time" rw:"r"`
	RoundTripEndTime   time.Time     `json:"end_time" csv:"end_time" rw:"r"`
	RoundTripDuration  time.Duration `json:"round_trip_duration"`
}

type ClusterMetric struct {
	MasterName     string
	HostMetricList []HostMetric `json:"host_metric_list" csv:"host_metric_list" rw:"r"`
}

func (p Property) validate() error {
	var invalidFields []string
	if p.RoundTripStartTime.IsZero() {
		invalidFields = append(invalidFields, "start_time")
	}
	if p.RoundTripEndTime.IsZero() {
		invalidFields = append(invalidFields, "end_time")
	}
	if len(invalidFields) != 0 {
		return fmt.Errorf("following fields are not valid. %v", invalidFields)
	}
	return nil
}

func (p Property) marshalToCSV() []string {
	return []string{
		p.RoundTripStartTime.Format(time.RFC3339),
		p.RoundTripEndTime.Format(time.RFC3339),
		p.RoundTripDuration.String(),
	}
}

func parseProperty(csvRecord []string) (Property, error) {
	if len(csvRecord) < 3 {
		return Property{}, fmt.Errorf("insufficient columns. column count %d", len(csvRecord))
	}
	parsedStartTime, err := time.Parse(time.RFC3339, csvRecord[0])
	if err != nil {
		return Property{}, err
	}
	parsedEndTime, err := time.Parse(time.RFC3339, csvRecord[1])
	if err != nil {
		return Property{}, err
	}
	parsedDuration, err := time.ParseDuration(csvRecord[2])
	if err != nil {
		return Property{}, err
	}
	return Property{
		RoundTripStartTime: parsedStartTime,
		RoundTripEndTime:   parsedEndTime,
		RoundTripDuration:  parsedDuration,
	}, nil
}
