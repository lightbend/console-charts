package monitor

import (
	"bytes"
	"fmt"
	"net/http"
)

// Connection to es-monitor-api server
type Connection struct {
	url string
}

// MakeMonitor creates a new threshold monitor with given values
func (m *Connection) MakeMonitor(name string, metric string, window string, confidence float32, threshold float32) error {
	url := fmt.Sprintf("%v/monitors/%v", m.url, name)

	json := fmt.Sprintf(`
	{
		"monitorVersion": "1",
		"model": "threshold",
		"parameters": {
		  "metric": "%v",
		  "window": "%v",
		  "confidence": "%v",
		  "severity": {
				"warning": {
					"comparator": "!=",
					"threshold": "%v"
				}
		  },
		  "summary": "summ",
		  "description": "desc"
		}
	}`, metric, window, confidence, threshold)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(json))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Author-Name", "me")
	req.Header.Set("Author-Email", "me@lightbend.com")
	req.Header.Set("Message", "testing")

	httpClient := &http.Client{}
	_, err = httpClient.Do(req)

	return err
}

// MakeSimpleMonitor creates a threshold monitor with non-configurable parameters.
func (m *Connection) MakeSimpleMonitor(name string, metric string) error {
	return m.MakeMonitor(name, metric, "5m", 1.0, 3.0)
}

// MakeAlertingMonitor creates a threshold monitor with low confidence
func (m *Connection) MakeAlertingMonitor(name string, metric string, threshold float32) error {
	return m.MakeMonitor(name, metric, "1m", 0.0001, threshold)
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
