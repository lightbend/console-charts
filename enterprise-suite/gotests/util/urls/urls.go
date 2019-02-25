package urls

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lightbend/gotests/util"
)

type Result struct {
	Body    string
	Headers http.Header
}

// Get200 retries until it gets a 200, otherwise it returns an error.
func Get200(url string) (Result, error) {
	var res Result

	err := util.WaitUntilSuccess(util.SmallWait, func() error {
		resp, err := http.Get(url)
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
		return nil
	})

	return res, err
}
