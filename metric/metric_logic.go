package metric

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultMetricFileNamePrefix = "metric"
	csvExtension                = "csv"
	defaultCycleDuration        = 24 * time.Hour
)

type Logic interface {
	GetFilePath() string
	ReadHostMetric() (HostMetric, error)
	WriteRecord(property Property) error
}
type BizLogic struct {
	isCurrentHostMaster bool
	hostName            string
	filePath            string
	dirPath             string
	rwLocker            sync.RWMutex
	recordStartTime     time.Time
}

func NewLogic(dirPath string, hostName string) Logic {
	logic := &BizLogic{
		dirPath:  dirPath,
		hostName: hostName,
	}
	logic.initNewFilePath()
	return logic
}

func (bl *BizLogic) ReadHostMetric() (HostMetric, error) {
	propertyList, err := bl.readAllRecords()
	if err != nil {
		return HostMetric{}, err
	}
	return HostMetric{
		Name:         bl.hostName,
		PropertyList: propertyList,
	}, nil
}

func (bl *BizLogic) GetFilePath() string {
	return bl.filePath
}

func (bl *BizLogic) readAllRecords() ([]Property, error) {
	var propertyResultList []Property
	bl.rwLocker.RLock()
	defer bl.rwLocker.RUnlock()
	fp, err := os.Open(bl.filePath)
	defer fp.Close()
	if err != nil {
		return propertyResultList, fmt.Errorf("error while reading metric. error: %v", err)
	}
	csvReader := csv.NewReader(fp)

	recordList, err := csvReader.ReadAll()
	if err != nil {
		return propertyResultList, fmt.Errorf("error while reading metric. error: %v", err)
	}
	for _, record := range recordList {
		property, err := parseProperty(record)
		if err != nil {
			return propertyResultList, fmt.Errorf("error while parsing a property record. error: %s", err)
		}
		propertyResultList = append(propertyResultList, property)
	}
	return propertyResultList, nil
}

func (bl *BizLogic) WriteRecord(property Property) error {
	if err := property.validate(); err != nil {
		return err
	}
	bl.rwLocker.Lock()
	defer bl.rwLocker.Unlock()
	if bl.shouldUpdateFilePath() {
		bl.initNewFilePath()
	}
	property.RoundTripDuration = property.RoundTripEndTime.Sub(property.RoundTripStartTime)
	if strings.EqualFold(bl.filePath, "") {
		return fmt.Errorf("filePath cannot be empty")
	}

	fp, err := os.OpenFile(bl.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	defer fp.Close()
	if err != nil {
		return fmt.Errorf("error while recording metric. error: %v", err)
	}
	csvWriter := csv.NewWriter(fp)
	err = csvWriter.Write(property.marshalToCSV())
	if err != nil {
		return err
	}
	csvWriter.Flush()

	return nil
}

func (bl *BizLogic) shouldUpdateFilePath() bool {
	return time.Now().Sub(bl.recordStartTime) > defaultCycleDuration
}

func (bl *BizLogic) initNewFilePath() {
	curTime := time.Now()
	bl.recordStartTime = curTime
	bl.filePath = fmt.Sprintf("%s%s_%d_%d_%d.%s", bl.dirPath, defaultMetricFileNamePrefix, curTime.Year(), curTime.Month(), curTime.Day(), csvExtension)
}
