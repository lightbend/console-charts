package monitor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Connection to console-api server
type Connection struct {
	url string
}

// MakeMonitor creates a new threshold monitor with given values.
// Confidence is a string because console-api accepts only certain values - 5e-324, 0.25, 0.5, 0.75, 0.95 and 1.
func (m *Connection) MakeMonitor(name string, metric string, window string, confidence string, threshold float32) error {
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

	return makeRequest(req)
}

// MakeSimpleMonitor creates a threshold monitor with non-configurable parameters.
func (m *Connection) MakeSimpleMonitor(name string, metric string) error {
	return m.MakeMonitor(name, metric, "5m", "1", 3.0)
}

// MakeAlertingMonitor creates a threshold monitor with low confidence
func (m *Connection) MakeAlertingMonitor(name string, metric string, threshold float32) error {
	return m.MakeMonitor(name, metric, "1m", "5e-324", threshold)
}

func (m *Connection) DeleteMonitor(name string) error {
	url := fmt.Sprintf("%v/monitors/%v", m.url, name)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Author-Name", "me")
	req.Header.Set("Author-Email", "me@lightbend.com")
	req.Header.Set("Message", "testing")

	return makeRequest(req)
}

func (m *Connection) CheckHealth() error {
	url := fmt.Sprintf("%v/status", m.url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	return makeRequest(req)
}

func makeRequest(req *http.Request) error {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("console-api replied with status code %v", resp.StatusCode)
	}

	return nil
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
