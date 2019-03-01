package alertmanager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

type Alert struct {
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	StartsAt     time.Time         `json:"startsAt,omitempty"`
	EndsAt       time.Time         `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL,omitempty"`
	Status       struct {
		State       string        `json:"state,omitempty"`
		SilencedBy  []interface{} `json:"silencedBy,omitempty"`
		InhibitedBy []interface{} `json:"inhibitedBy,omitempty"`
	} `json:"status,omitempty"`
	Receivers   []string `json:"receivers,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

type AlertsResponse struct {
	Status string  `json:"status"`
	Data   []Alert `json:"data,omitempty"`
}

// Connection to alermanager server
type Connection struct {
	url string
}

func (a *Connection) Alerts() ([]Alert, error) {
	addr := fmt.Sprintf("%v/api/v1/alerts", a.url)
	resp, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	defer util.Close(resp.Body)

	if resp.StatusCode == 200 {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var alerts AlertsResponse
		if err := json.Unmarshal(content, &alerts); err != nil {
			return nil, err
		}

		if alerts.Status != "success" {
			return nil, fmt.Errorf("unexpected status in alertmanager response: %v", alerts.Status)
		}

		return alerts.Data, nil
	} else {
		return nil, fmt.Errorf("expected 200 alertmanager response status code, got %v", resp.StatusCode)
	}
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
