package prometheus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

type PromData struct {
	ResultType string      `json:"resultType,omitempty"`
	Result     interface{} `json:"result,omitempty"`
}

type PromResponse struct {
	Original  string
	Status    string   `json:"status"`
	Data      PromData `json:"data,omitempty"`
	ErrorType string   `json:"errorType,omitempty"`
	Error     string   `json:"error,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// Connection to prometheus server
type Connection struct {
	url string
}

func (p *Connection) Query(query string) (*PromResponse, error) {
	addr := fmt.Sprintf("%v/api/v1/query?query=%v", p.url, url.QueryEscape(query))
	// Some of the tests with openshift clusters timed out in 20 seconds, so adding a timeout for 45 seconds.
	timeout := time.Duration(45 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(addr)
	if err != nil {
		return nil, err
	}

	defer util.Close(resp.Body)
	// Prometheus docs say 2XX codes are used for success, not just 200
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var promResp PromResponse
		if err := json.Unmarshal(content, &promResp); err != nil {
			return nil, err
		} else {
			promResp.Original = string(content)
			return &promResp, nil
		}
	}

	return nil, fmt.Errorf("prometheus response status %v", resp.StatusCode)
}

func (p *Connection) HasData(query string, formatArgs ...interface{}) error {
	return p.checkForData(fmt.Sprintf(query, formatArgs...), true)
}

func (p *Connection) HasNoData(query string, formatArgs ...interface{}) error {
	return p.checkForData(fmt.Sprintf(query, formatArgs...), false)
}

func (p *Connection) checkForData(query string, expectResults bool) error {
	resp, err := p.Query(query)

	if err != nil {
		return fmt.Errorf("%s returned an error: %v", query, err)
	}

	// Cast result to array of anything
	arr, ok := resp.Data.Result.([]interface{})
	if !ok {
		return fmt.Errorf("expected array of values, but was %q: %s ->\n%s", reflect.TypeOf(resp.Data.Result),
			query, util.IndentJson(resp.Original))
	}

	if len(arr) == 0 && expectResults {
		return fmt.Errorf("expected >1 result, got 0: %s ->\n%s", query, util.IndentJson(resp.Original))
	}

	if len(arr) > 0 && !expectResults {
		return fmt.Errorf("expected 0 results, got %d: %s ->\n%s", len(arr), query, util.IndentJson(resp.Original))
	}

	return nil
}

// find any instance of query over past 10 minutes
func (p *Connection) AnyData(query string) error {
	return p.HasData(fmt.Sprintf("count_over_time( (%v) [10m:] )", query))
}

func (p *Connection) HasNScrapes(metric string, n int) error {
	if n < 1 {
		return fmt.Errorf("HasNScrapes: n must be 1 or higher, was %v", n)
	}
	return p.HasData(fmt.Sprintf("count_over_time(%v[10m]) > %v", metric, n-1))
}

func (p *Connection) HasModel(model string) error {
	return p.HasData(fmt.Sprintf("model{name=\"%v\"}", model))
}

func NewConnection(url string) (*Connection, error) {
	// TODO: verify connection works
	return &Connection{url}, nil
}
