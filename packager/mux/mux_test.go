// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mux

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgt "github.com/ampproject/amppackager/packager/testing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promtest "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// expand propagates hardcoded values into template test url.
func expand(templateURL string) string {
	templateURL = strings.Replace(templateURL, "$HOST", "http://www.publisher_amp_server.com", 1)
	templateURL = strings.Replace(templateURL, "$FETCH", "http://www.publisher_main_server.com/some_page", 1)
	templateURL = strings.Replace(templateURL, "$SIGN", "https://www.publisher_main_server.com/some_page", 1)
	templateURL = strings.Replace(templateURL, "$CERT", pkgt.CertName, 1)
	return templateURL
}

// mockedHandler mocks mux' underlying http handlers - signer, cert etc.
type mockedHandler struct {
	mock.Mock
}

func (m *mockedHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	m.Called(Params(req))
}

func TestServeHTTPSuccess(t *testing.T) {
	templateTests := []struct {
		testName      string
		testURL       string
		expectHandler string
		expectParams  map[string]string
	}{
		{
			testName:      `Signer - empty`,
			testURL:       `$HOST/priv/doc`,
			expectHandler: `signer`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Signer - with query, empty`,
			testURL:       `$HOST/priv/doc?`,
			expectHandler: `signer`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Signer - with query, regular`,
			testURL:       `$HOST/priv/doc?fetch=$FETCH&sign=$SIGN`,
			expectHandler: `signer`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Signer - with query, escaping`,
			testURL:       `$HOST/priv/doc?fetch=$FETCH&sign=$SIGN%2A\`,
			expectHandler: `signer`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Signer - with path, empty`,
			testURL:       `$HOST/priv/doc/`,
			expectHandler: `signer`,
			expectParams:  map[string]string{`signURL`: ``},
		}, {
			testName:      `Signer -  with path, regular`,
			testURL:       `$HOST/priv/doc/$FETCH`,
			expectHandler: `signer`,
			expectParams:  map[string]string{`signURL`: `$FETCH`},
		}, {
			testName:      `Signer - with path, escaping`,
			testURL:       `$HOST/priv/doc/$FETCH%2A\`,
			expectHandler: `signer`,
			expectParams:  map[string]string{`signURL`: `$FETCH%2A%5C`},
		}, {
			testName:      `Signer - with path and query, regular`,
			testURL:       `$HOST/priv/doc/$FETCH?amp=1`,
			expectHandler: `signer`,
			expectParams:  map[string]string{`signURL`: `$FETCH?amp=1`},
		}, {
			testName:      `Signer - with path and query, escaping`,
			testURL:       `$HOST/priv/doc/$FETCH%2A\?amp=1%2A\`,
			expectHandler: `signer`,
			expectParams:  map[string]string{`signURL`: `$FETCH%2A%5C?amp=1%2A\`},
		}, {
			testName:      `Cert - empty`,
			testURL:       `$HOST/amppkg/cert/`,
			expectHandler: `cert`,
			expectParams:  map[string]string{`certName`: ``},
		}, {
			testName:      `Cert - regular`,
			testURL:       `$HOST/amppkg/cert/$CERT`,
			expectHandler: `cert`,
			expectParams:  map[string]string{`certName`: `$CERT`},
		}, {
			testName:      `Cert - escaping`,
			testURL:       `$HOST/amppkg/cert/$CERT%2A\`,
			expectHandler: `cert`,
			expectParams:  map[string]string{`certName`: `$CERT*\`},
		}, {
			testName:      `ValidityMap - regular`,
			testURL:       `$HOST/amppkg/validity`,
			expectHandler: `validityMap`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Healthz - regular`,
			testURL:       `$HOST/healthz`,
			expectHandler: `healthz`,
			expectParams:  map[string]string{},
		}, {
			testName:      `Metrics - regular`,
			testURL:       `$HOST/metrics`,
			expectHandler: `metrics`,
			expectParams:  map[string]string{},
		},
	}
	for _, tt := range templateTests {
		testName := tt.testName
		t.Run(testName, func(t *testing.T) {
			// Defer validation to ensure it does happen.
			mocks := map[string](*mockedHandler){"signer": &mockedHandler{}, "healthz": &mockedHandler{}, "cert": &mockedHandler{}, "validityMap": &mockedHandler{}, "metrics": &mockedHandler{}}
			var actualResp *http.Response
			defer func() {
				// Expect no errors.
				assert.Equal(t, 200, actualResp.StatusCode, "No error expected: %#v", actualResp)

				// Expect the right call to the right handler, and no calls to the rest.
				for _, mockedHandler := range mocks {
					mockedHandler.AssertExpectations(t)
				}
			}()

			// expand template URL and expectParams.
			tt.testURL = expand(tt.testURL)
			for v, k := range tt.expectParams {
				tt.expectParams[v] = expand(k)
			}

			// Set expectation.
			expectMockedHandler := mocks[tt.expectHandler]
			expectMockedHandler.On("ServeHTTP", tt.expectParams)

			// Run.
			mux := New(mocks["cert"], mocks["signer"], mocks["validityMap"], mocks["healthz"], mocks["metrics"])
			actualResp = pkgt.NewRequest(t, mux, tt.testURL).Do()
		})
	}
}

func expectError(t *testing.T, url string, expectErrorMessage string, expectErrorCode int, body io.Reader) {
	// Defer validation to ensure it does happen.
	mockedHandler := new(mockedHandler)
	var actualResp *http.Response
	var actualErrorMessage string
	defer func() {
		// Expect the right error.
		assert.Equal(t, expectErrorCode, actualResp.StatusCode)
		assert.Equal(t, expectErrorMessage, actualErrorMessage)

		// Expect no calls to mocks.
		mockedHandler.AssertExpectations(t)
	}()

	// Initialize mux with 4 identical mocked handlers, because no calls are expect to any of them.
	mux := New(mockedHandler, mockedHandler, mockedHandler, mockedHandler, mockedHandler)

	// Run and extract error.
	actualResp = pkgt.NewRequest(t, mux, url).SetBody(body).Do()
	actualErrorMessageBuffer, _ := ioutil.ReadAll(actualResp.Body)
	actualErrorMessage = fmt.Sprintf("%s", actualErrorMessageBuffer)

}

func TestServeHTTPexpect404s(t *testing.T) {
	templateTests := []struct {
		testName string
		URL      string
	}{
		{"No such endpoint                      ", "$HOST/abc"},
		{"Signer - unexpected extra char        ", "$HOST/priv/doc1"},
		{"Cert - no closing slash               ", "$HOST/amppkg/cert"},
		{"ValidityMap - unexpected closing slash", "$HOST/amppkg/validity/"},
		{"Healthz - unexpected closing slash    ", "$HOST/healthz/"},
		{"Healthz - unexpected extra char       ", "$HOST/healthz1"},
		{"Metrics - unexpected closing slash    ", "$HOST/metrics/"},
	}
	for _, tt := range templateTests {
		t.Run(tt.testName, func(t *testing.T) {
			expectError(t, expand(tt.URL), "404 page not found\n", http.StatusNotFound, nil)
		})
	}
}

func TestServeHTTPexpect405(t *testing.T) {
	body := strings.NewReader("Non empty body so this sends a POST request")
	expectError(t, expand("$HOST/healthz"), "405 method not allowed\n", http.StatusMethodNotAllowed, body)
}

func TestParamsIncorrectValueType(t *testing.T) {
	req := httptest.NewRequest("", "http://abc.com", nil)

	// Pass string instead of expected map[string]string.
	req = req.WithContext(context.WithValue(req.Context(), paramsKey, "Some string"))

	// Expect Params to handle invalid input gracefully.
	assert.Equal(t, Params(req), map[string]string{})
}

const promExpectedHeaderRequestsTotal = `
	# HELP total_requests_by_code_and_url Total number of requests by HTTP code and URL.
	# TYPE total_requests_by_code_and_url counter
	`

// TestPrometheusMetricRequestsTotal tests the respective Prometheus metric.
// Test each scenario in isolation to make sure each of them works, then test
// them altogether to make sure they don't interfere with each other.
func TestPrometheusMetricRequestsTotal(t *testing.T) {
	nopHandler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	tests := []struct {
		testName        string
		testHint        string
		testFunc        func()
		expectedMetrics string
	}{
		{
			/* testName= */ `AllHandlersNOP200`,
			/* testHint= */ `
				Make requests to mux with all handlers being NOPs returning 200.
				Request Healthz twice to test aggregation of identical requests.
			`,
			/* testFunc= */ func() {
				mux := New(nopHandler, nopHandler, nopHandler, nopHandler, nopHandler)
				pkgt.NewRequest(t, mux, expand(`$HOST/priv/doc?fetch=$FETCH&sign=$SIGN`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/amppkg/cert/$CERT`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/amppkg/validity`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/healthz`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/healthz`)).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="200",handler="signer"} 1
				total_requests_by_code_and_url{code="200",handler="certCache"} 1
				total_requests_by_code_and_url{code="200",handler="validityMap"} 1
				total_requests_by_code_and_url{code="200",handler="healthz"} 2
			`,
		},
		{
			/* testName= */ `ErrorsReturnedByMuxDirectly`,
			/* testHint= */ `
				Test counting requests to same handler that returned different codes.
				Trigger a 404 attributed to healthz by adding an unexpected suffix to path.
			`,
			/* testFunc= */ func() {
				mux := New(nopHandler, nopHandler, nopHandler, nopHandler, nopHandler)
				pkgt.NewRequest(t, mux, expand(`$HOST/healthzSOME_SUFFIX`)).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="404",handler="healthz"} 1
			`,
		},
		{
			/* testName= */ `UnassignedRequests`,
			/* testHint= */ `
				Test counting request not assigned to a handler.
			`,
			/* testFunc= */ func() {
				mux := New(nopHandler, nopHandler, nopHandler, nopHandler, nopHandler)
				pkgt.NewRequest(t, mux, expand(`$HOST/abc`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/def`)).Do()
				pkgt.NewRequest(t, mux, expand(`$HOST/ghi`)).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="404",handler="handler_not_assigned"} 3
			`,
		},
		{
			/* testName= */ `ForbiddenMethod`,
			/* testHint= */ `
				Special case: forbidden method.
			`,
			/* testFunc= */ func() {
				mux := New(nopHandler, nopHandler, nopHandler, nopHandler, nopHandler)
				body := strings.NewReader("Non empty body so this will be a POST request")
				pkgt.NewRequest(t, mux, expand("$HOST/healthz")).SetBody(body).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="405",handler="healthz"} 1
			`,
		},
		{
			/* testName= */ `ErrorReturnedByHandler`,
			/* testHint= */ `
				Some of the above requests generated errors, but those errors were thrown
				by mux, not by handlers. Handlers were no-ops. Now let's simulate a
				request that triggers a handler-generated error.
				Specifically let's simulate signer returning a 400.
			`,
			/* testFunc= */ func() {
				signerMockReturning400 := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "Bad Request", 400) }))
				mux := New(nopHandler, signerMockReturning400, nopHandler, nopHandler, nopHandler)
				pkgt.NewRequest(t, mux, expand("$HOST/priv/doc/abc")).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="400",handler="signer"} 1
			`,
		},
		{
			/* testName= */ `FetchMetricsEndpoint`,
			/* testHint= */ `
				Special case: send a request to "metrics" endpoint, which results in two actions:
				1) Previously collected metrics are returned in response with status 200.
				2) Prometheus requests counter is incremented for respective handler and code ("metric", 200).
				Let's test that these actions work fine together.
				Let's send a "metric" request to a new mux instance that has a real, non-mocked
				metric handler. Such request, along with downstream validation, checks
				that the "metric" endpoint's underlying logic doesn't interfere
				with accounting for the actual metric request.
			`,
			/* testFunc= */ func() {
				mux := New(nopHandler, nopHandler, nopHandler, nopHandler, promhttp.Handler())
				pkgt.NewRequest(t, mux, expand(`$HOST/metrics`)).Do()
			},
			/* expectedMetrics = */ `
				total_requests_by_code_and_url{code="200",handler="metrics"} 1
			`,
		},
	}

	// Test each scenario in isolation.
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			promRequestsTotal.Reset()
			expectedMetrics := promExpectedHeaderRequestsTotal + tt.expectedMetrics
			tt.testFunc()
			expectation := strings.NewReader(expectedMetrics)
			if err := promtest.CollectAndCompare(promRequestsTotal, expectation, "total_requests_by_code_and_url"); err != nil {
				t.Errorf("TestPrometheusMetricRequestsTotal - "+tt.testName+": unexpected collecting result:\n%s", err)
			}
		})
	}

	// Test all scenarios together.
	promRequestsTotal.Reset()
	expectedMetrics := promExpectedHeaderRequestsTotal
	for _, tt := range tests {
		expectedMetrics += tt.expectedMetrics
		tt.testFunc()
	}
	expectation := strings.NewReader(expectedMetrics)
	if err := promtest.CollectAndCompare(promRequestsTotal, expectation, "total_requests_by_code_and_url"); err != nil {
		t.Errorf("TestPrometheusMetricRequestsTotal - all scenarios in single run: unexpected collecting result:\n%s", err)
	}
}

// TestPrometheusMetricRequestsLatency tests the end-to-end latencies metrics.
// It checks that the right error codes and handlers are accounted for. It also
// checks that the latencies are positive, but doesn't expect exact values,
// because latencies are non-deterministic.
// It would be nice to mock time (e.g. patch the time.Since function) to test
// the exact latencies values produced, and to simulate slow execution, too.
// However, seems like there's no "native" way to monkey-patch in Go.
// There is an option that doesn't look safe enough:
// https://www.reddit.com/r/golang/comments/30try1/monkey_patching_in_go/
// https://news.ycombinator.com/item?id=22442170.
func TestPrometheusMetricRequestsLatency(t *testing.T) {
	hintPrefix := "TestPrometheusMetricRequestsLatency"

	type metricRecordKey struct {
		codeLabelPair, handlerLabelPair string
	}

	type scenarioRequests []struct {
		urlTemplate          string
		mockHandlerThrows404 bool
	}

	type scenarioExpectedSampleCountMap map[metricRecordKey]uint64

	scenarios := []struct {
		requests               scenarioRequests
		expectedSampleCountMap scenarioExpectedSampleCountMap
	}{
		{
			scenarioRequests{
				{urlTemplate: `$HOST/priv/doc?fetch=$FETCH&sign=$SIGN`, mockHandlerThrows404: true},
				{urlTemplate: `$HOST/priv/doc?fetch=$FETCH&sign=$SIGN`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"404" `, `name:"handler" value:"signer" `}: 1,
				{`name:"code" value:"200" `, `name:"handler" value:"signer" `}: 1,
			},
		},
		{
			scenarioRequests{
				{urlTemplate: `$HOST/amppkg/cert/$CERT`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"200" `, `name:"handler" value:"certCache" `}: 1,
			},
		},
		{
			scenarioRequests{
				{urlTemplate: `$HOST/amppkg/validity`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"200" `, `name:"handler" value:"validityMap" `}: 1,
			},
		},
		{
			scenarioRequests{
				{urlTemplate: `$HOST/healthz`, mockHandlerThrows404: false},
				{urlTemplate: `$HOST/healthz`, mockHandlerThrows404: false},
				{urlTemplate: `$HOST/healthz`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"200" `, `name:"handler" value:"healthz" `}: 3,
			},
		},
		{
			scenarioRequests{
				{urlTemplate: `$HOST/metrics`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"200" `, `name:"handler" value:"metrics" `}: 1,
			},
		},
		{
			scenarioRequests{
				{urlTemplate: `$HOST/abc`, mockHandlerThrows404: false},
				{urlTemplate: `$HOST/def`, mockHandlerThrows404: false},
			},
			scenarioExpectedSampleCountMap{
				{`name:"code" value:"404" `, `name:"handler" value:"handler_not_assigned" `}: 2,
			},
		},
	}

	for _, scenario := range scenarios {
		promRequestsLatency.Reset()

		for _, req := range scenario.requests {
			mockHandler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if req.mockHandlerThrows404 {
					http.Error(w, "404 page not found", 404)
				}
			}))
			mux := New(mockHandler, mockHandler, mockHandler, mockHandler, mockHandler)
			pkgt.NewRequest(t, mux, expand(req.urlTemplate)).Do()

		}

		expectedSampleCountMap := scenario.expectedSampleCountMap

		reg := prometheus.NewPedanticRegistry()
		if err := reg.Register(promRequestsLatency); err != nil {
			t.Errorf(hintPrefix+" - registering collector failed: %s", err)
		}

		actualMetricFamilyArr, err := reg.Gather()
		if err != nil {
			t.Errorf(hintPrefix+" - gathering metrics failed: %s", err)
		}

		assert.Equal(t, 1, len(actualMetricFamilyArr),
			hintPrefix+" expects exactly one metric family.")

		assert.Equal(t, "request_latencies_in_seconds", *actualMetricFamilyArr[0].Name,
			hintPrefix+" expects the right metric name.")

		assert.Equal(t, len(expectedSampleCountMap), len(actualMetricFamilyArr[0].Metric),
			hintPrefix+" expects the right amount of metrics collected and gathered.")

		for _, actualMetric := range actualMetricFamilyArr[0].Metric {
			// Expect the right sample count.
			code := actualMetric.Label[0].String()
			handler := actualMetric.Label[1].String()
			expectedSampleCount := expectedSampleCountMap[metricRecordKey{code, handler}]
			actualSampleCount := actualMetric.Summary.GetSampleCount()
			assert.Equal(t, expectedSampleCount, actualSampleCount, hintPrefix+" expects the right sample count for "+code+" "+handler)

			// Expect the right number of quantiles.
			assert.Equal(t, 3, len(actualMetric.Summary.Quantile), hintPrefix+" expects the right number of quantiles.")

			// Expect the right quantiles.
			// Expect positive quantile values, because latencies are non-zero.
			// Don't check the exact values, because latencies are non-deterministic.
			expectedQuantileKeys := []float64{0.5, 0.9, 0.99}
			for i, quantile := range actualMetric.Summary.Quantile {
				assert.Equal(t, expectedQuantileKeys[i], quantile.GetQuantile(), hintPrefix+" expects the right quantile.")
				assert.True(t, quantile.GetValue() > .0, hintPrefix+" expects non-zero quantile value (latency).")
			}
		}
	}
}
