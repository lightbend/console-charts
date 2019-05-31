package urls

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/lightbend/console-charts/enterprise-suite/gotests/util"
)

type Result struct {
	Body    string
	Headers http.Header
	Status  int
}

// Get200 retries until it gets a 200, otherwise it returns an error.
func Get200(url string) (Result, error) {
	var res Result

	err := util.WaitUntilSuccess(util.SmallWait, func() error {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("wanted 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}

		res.Body = string(bodyBytes)
		res.Headers = resp.Header
		res.Status = resp.StatusCode
		return nil
	})

	return res, err
}

// Get returns any response without expecting any particular status code
func Get(url string, followRedirects bool) (Result, error) {
	var res Result

	err := util.WaitUntilSuccess(util.SmallWait, func() error {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		if !followRedirects {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		res.Body = string(bodyBytes)
		res.Headers = resp.Header
		res.Status = resp.StatusCode
		return nil
	})

	return res, err
}
