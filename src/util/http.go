// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/coreos/ignition/src/log"
)

// HttpClient is a simple wrapper around the Go HTTP client that standardizes
// the process and logging of fetching payloads.
type HttpClient struct {
	client *http.Client
	logger log.Logger
}

// NewHttpClient creates a new client with the given logger.
func NewHttpClient(logger log.Logger) HttpClient {
	return HttpClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Get performs an HTTP GET on the provided URL and returns the response body,
// HTTP status code, and error (if any).
func (c HttpClient) Get(url string) ([]byte, int, error) {
	var body []byte
	var status int

	err := c.logger.LogOp(func() error {
		resp, err := c.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		status = resp.StatusCode
		c.logger.Debug("GET result: %s", http.StatusText(status))
		body, err = ioutil.ReadAll(resp.Body)

		return err
	}, "GET %q", url)

	return body, status, err
}

// FetchConfig fetches a raw config from the provided URL and returns the
// response body on success or nil on failure. The caller must also provide a
// list of acceptable HTTP status codes. If the response's status code is not in
// the provided list, it is considered a failure.
func (c HttpClient) FetchConfig(url string, acceptedStatuses ...int) []byte {
	var config []byte

	c.logger.LogOp(func() error {
		data, status, err := c.Get(url)
		if err != nil {
			return err
		}

		for _, acceptedStatus := range acceptedStatuses {
			if status == acceptedStatus {
				config = data
				return nil
			}
		}

		return fmt.Errorf("%s", http.StatusText(status))
	}, "fetching config from %q", url)

	return config
}
