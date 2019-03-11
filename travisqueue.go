// Copyright 2016-2018, Pulumi Corporation.
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
//
// Original source at https://github.com/pulumi/travisqueue.
//
// Modified by james.ravn@lightbend.com to queue project wide, so no more than 1 concurrent build at a time.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func mustGetenv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("ERROR: %v is not set\n", key)
	}

	return value
}

func mustParseURL(v string) *url.URL {
	url, err := url.Parse(v)
	if err != nil {
		log.Fatalf("can't parse %v as URL: %v", v, err)
	}
	return url
}

func mustAtoi(v string) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("can't convert %v to int: %v", v, err)
	}
	return i
}

var (
	travisEndpoint = mustParseURL(mustGetenv("TRAVIS_ENDPOINT"))
	travisToken    = mustGetenv("TRAVIS_TOKEN")

	// https://docs.travis-ci.com/user/environment-variables/#Default-Environment-Variables
	travisBuildID = mustAtoi(mustGetenv("TRAVIS_BUILD_ID"))
	travisRepoSlug  = mustGetenv("TRAVIS_REPO_SLUG")
)

// https://developer.travis-ci.org/resource/build#Build
// This definition only includes fields we need.
type Build struct {
	ID int

	Number string
	State  string

	// e.g. "2006-01-02T15:04:05Z" or nil if not started
	StartedAt *string `json:"started_at"`
}

// https://developer.travis-ci.org/resource/builds#Builds
type Builds struct {
	Builds []Build
}

// If bodyValue is non-nil, decodes body as JSON into it.
// Exits on error.
func callTravisAPI(method, path string, expectStatus int, bodyValue interface{}) {
	url := travisEndpoint.ResolveReference(mustParseURL(path))
	req, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		log.Fatalf("couldn't create request to %v", url)
	}

	req.Header.Add("Travis-API-Version", "3")
	req.Header.Add("Authorization", "token "+travisToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("request to %v failed: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != expectStatus {
		log.Fatalf("request to %v failed: %v", url, res.Status)
	}

	if bodyValue != nil {
		err = json.NewDecoder(res.Body).Decode(bodyValue)
		if err != nil {
			log.Fatalf("can't decode response as %T: %v", bodyValue, err)
		}
	}
}

// Return the build
// - in this repository
// - of this branch
// - started by a `push` event
// - with a state in `states`, or in any state if `states` is empty
// - that sorts first by `sortBy`, as interpreted by the Travis API.
// Exits on error or if no matching build is found.
// https://developer.travis-ci.com/resource/builds#find
func firstMatchingBuild(states, sortBy string) Build {
	vs := url.Values{}
	vs.Add("sort_by", sortBy)
	if states != "" {
		vs.Add("build.state", states)
	}
	vs.Add("limit", "1")

	var builds Builds

	path := fmt.Sprintf("/repo/%v/builds?%v", url.PathEscape(travisRepoSlug), vs.Encode())
	callTravisAPI("GET", path, http.StatusOK, &builds)

	if len(builds.Builds) == 0 {
		log.Fatal("found no builds")
	}

	return builds.Builds[0]
}

func earliestStartedBuild() Build {
	return firstMatchingBuild("started", "started_at")
}

func newestFinishedBuild() Build {
	return firstMatchingBuild("passed,failed,errored", "id:desc")
}

func newestBuild() Build {
	return firstMatchingBuild("", "id:desc")
}

func cancelThisBuild() {
	log.Print("Cancelling this build...")

	path := fmt.Sprintf("/build/%v/cancel", travisBuildID)
	callTravisAPI("POST", path, http.StatusAccepted, nil)

	// Wait for the build to be cancelled. Travis' build timeout is 2 hours.
	time.Sleep(3 * time.Hour)
}

func restartBuild(id int) {
	path := fmt.Sprintf("/build/%v/restart", id)
	callTravisAPI("POST", path, http.StatusAccepted, nil)
}

func main() {
	command := ""
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "start":
		// Check we're the running build with the earliest start time.
		earliest := earliestStartedBuild()
		if earliest.ID != travisBuildID {
			log.Printf("Found an older build running: %v (%v) started at %v\n", earliest.Number, earliest.ID, *earliest.StartedAt)
			cancelThisBuild()
		}

		// Check there are no newer, finished builds.
		finished := newestFinishedBuild()
		if finished.ID > travisBuildID {
			log.Printf("Found a newer finished build: %v (%v), state %v\n", finished.Number, finished.ID, finished.State)
			cancelThisBuild()
		}

		// Okay to proceed.

	case "finish":
		// Restart the newest queued build if it is cancelled.
		newest := newestBuild()
		if newest.State == "canceled" /* [sic] */ {
			log.Printf("Restarting cancelled build %v (%v)\n", newest.Number, newest.ID)
			restartBuild(newest.ID)
		}

	default:
		log.Fatalf("Usage: %v {start|finish}\n", os.Args[0])
	}
}
