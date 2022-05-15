package metric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const defaultPort = ":8081"

type serverCtx struct {
	metricLogic Logic
}

func NewMetricServer(metricDirectoryPath string) error {
	currentHostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to fetch current host name. error: %v", err)
	}
	logic := NewLogic(metricDirectoryPath, currentHostname)
	server := serverCtx{
		metricLogic: logic,
	}
	http.HandleFunc("/metric", server.fetchHostMetric)
	err = http.ListenAndServe(defaultPort, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *serverCtx) fetchHostMetric(resp http.ResponseWriter, req *http.Request) {
	if !strings.EqualFold(req.Method, http.MethodGet) {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("request method not supported"))
		return
	}
	hostMetric, err := s.metricLogic.ReadHostMetric()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(fmt.Sprintf("error while reading host metric. error: %v", err)))
		return
	}
	responseData, err := json.Marshal(hostMetric)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(fmt.Sprintf("error while marshalling data. error:%v", err)))
		return
	}
	resp.Header().Set("content-type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write(responseData)
}
