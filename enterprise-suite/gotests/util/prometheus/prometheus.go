package prometheus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
)

type Data struct {
	ResultType string      `json:"resultType,omitempty"`
	Result     interface{} `json:"result,omitempty"`
}

type Response struct {
	Status    string   `json:"status"`
	Data      Data     `json:"data,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
	Error     string   `json:"error,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// Connection to prometheus server
type Connection struct {
	url string
}

func (p *Connection) Query(query string) (*Response, error) {
	addr := fmt.Sprintf("%v/api/v1/query?query=%v", p.url, url.QueryEscape(query))

	resp, err := http.Get(addr)
	if err != nil {
		return nil, err
	}

	// Prometheus docs say 2XX codes are used for success, not just 200
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var resp Response
		if err := json.Unmarshal(content, &resp); err != nil {
			return nil, err
		} else {
			return &resp, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("prometheus response status %v", resp.StatusCode))
}

func (p *Connection) HasData(query string) error {
	resp, err := p.Query(query)

	if err != nil {
		return fmt.Errorf("%q returned an error: %v", query, err)
	}

	// Cast result to array of anything
	arr, ok := resp.Data.Result.([]interface{})
	if !ok {
		return fmt.Errorf("%q - expected array of values, but was %v", query, reflect.TypeOf(resp.Data.Result))
	}

	if len(arr) == 0 {
		return fmt.Errorf("%q returned 0 results", query)
	}

	return nil
}

func (p *Connection) HasModel(model string) error {
	return p.HasData(fmt.Sprintf("model{name=\"%v\"}", model))
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
