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

// Make adds a simple threshold metric with non-configurable parameters.
// TODO: Make this more configurable!
func (m *Connection) Make(name string, metric string) error {
	url := fmt.Sprintf("%v/monitors/%v", m.url, name)

	json := fmt.Sprintf(`
	{
		"monitorVersion": "1",
		"model": "threshold",
		"parameters": {
		  "metric": "%v",
		  "window": "5m",
		  "confidence": "1",
		  "severity": {
			"warning": {
			  "comparator": "!=",
			  "threshold": "1"
			}
		  },
		  "summary": "summ",
		  "description": "desc"
		}
	  }`, metric)

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

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
