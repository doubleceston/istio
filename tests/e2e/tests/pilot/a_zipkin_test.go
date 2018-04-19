// Copyright 2017 Istio Authors
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

// TODO(nmittler): Named file with prefix "a_" to force golang to run it first.  Running it last results in flakes.

package pilot

import (
	"fmt"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
)

const (
	traceHeader = "X-Client-Trace-Key"
	numTraces   = 5
)

func TestZipkin(t *testing.T) {
	for i := 0; i < numTraces; i++ {
		testName := fmt.Sprintf("index_%d", i)
		traceSent := false
		var id string

		runRetriableTest(t, testName, defaultRetryBudget, func() error {
			if !traceSent {
				// Send a request with a trace header.
				id = uuid.NewV4().String()
				response := ClientRequest("a", "http://b", 1,
					fmt.Sprintf("-key %v -val %v", traceHeader, id))
				if !response.IsHTTPOk() {
					// Keep retrying until we successfully send a trace request.
					return errAgain
				}

				traceSent = true
			}

			// Check the zipkin server to verify the trace was received.
			response := ClientRequest(
				"t",
				fmt.Sprintf("http://zipkin.%s:9411/api/v1/traces?annotationQuery=guid:x-client-trace-id=%s",
					tc.Kube.IstioSystemNamespace(), id),
				1, "",
			)

			if !response.IsHTTPOk() {
				return errAgain
			}

			// Check that the trace contains the id value (must occur more than once, as the
			// response also contains the request URL with query parameter).
			if strings.Count(response.Body, id) == 1 {
				return errAgain
			}

			return nil
		})
	}
}
