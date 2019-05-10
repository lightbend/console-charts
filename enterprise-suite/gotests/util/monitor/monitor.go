package monitor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/onsi/ginkgo"
)

// Connection to console-api server
type Connection struct {
	url string
}

// MakeThresholdMonitor creates a new threshold monitor with given values.
// Confidence is a string because console-api accepts only certain values - 5e-324, 0.25, 0.5, 0.75, 0.95 and 1.
func (m *Connection) MakeThresholdMonitor(name, metric, window, confidence, comparator, threshold string) error {
	url := fmt.Sprintf("%v/monitors/%v", m.url, name)

	json := fmt.Sprintf(`
	{
		"monitorVersion": "1",
		"model": "threshold",
		"parameters": {
		  "metric": "%s",
		  "window": "%s",
		  "confidence": "%s",
		  "warmup": "1s",
		  "severity": {
				"warning": {
					"comparator": "%s",
					"threshold": "%s"
				}
		  },
		  "summary": "summ",
		  "description": "desc"
		}
	}`, metric, window, confidence, comparator, threshold)

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

// MakeSimpleMonitor creates a simple threshold monitor.
func (m *Connection) MakeSimpleMonitor(name string, metric string) error {
	return m.MakeThresholdMonitor(name, metric, "5m", "1", ">", "9999")
}

// MakeAlertingMonitor creates a monitor which is always alerting.
func (m *Connection) MakeAlertingMonitor(name string) error {
	if err := m.DeleteMonitor(name); err != nil {
		// Delete monitor in case it exists
		// ignore any error
	}
	return m.MakeThresholdMonitor(name, "up", "1m", "5e-324", "!=", "-1")
}

func (m *Connection) TryDeleteMonitor(name string) {
	if err := m.DeleteMonitor(name); err != nil {
		// ignore error
	}
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
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintf(ginkgo.GinkgoWriter, "response: %s\n", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("console-api replied with status code %v", resp.StatusCode)
	}

	return nil
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
