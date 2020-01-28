package scenario

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"time"

	restify "github.com/SpaceStock/go-restify"
	"github.com/SpaceStock/go-restify/enum/onfailure"
)

//implementation of restify.Scenario
type scenario struct {
	cache map[string]json.RawMessage

	id          string
	name        string
	description string
	environment string

	cases []restify.TestCase

	getter restify.ScenarioGetter
	setter restify.ScenarioSetter
}

//New Scenario create a new test scenario
func New() restify.Scenario {
	s := &scenario{
		cache: make(map[string]json.RawMessage),
		cases: []restify.TestCase{},
	}

	s.getter = &getter{s}
	s.setter = &setter{s}
	return s
}

func (s *scenario) String() string {
	b, err := s.MarshalJSON()
	if err != nil {
		return fmt.Sprintf(`{ERROR: %s}`, err.Error())
	}

	return string(b)
}

func (s *scenario) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID          string             `json:"id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Environment string             `json:"environment"`
		Cases       []restify.TestCase `json:"cases"`
	}{
		ID:          s.id,
		Name:        s.name,
		Description: s.description,
		Environment: s.environment,
		Cases:       s.cases,
	})
}

func (s *scenario) UnmarshalJSON(data []byte) error {
	alias := struct {
		ID          string             `json:"id"`
		Name        string             `json:"name"`
		Description string             `json:"description"`
		Environment string             `json:"environment"`
		Cases       []restify.TestCase `json:"cases"`
	}{}

	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}

	s.id = alias.ID
	s.name = alias.Name
	s.description = alias.Description
	s.environment = alias.Environment
	s.cases = alias.Cases
	return nil
}

//TODO: UGLY AF CODE! TOO EFFIN FAT! NEED TO REFACTOR!
func (s *scenario) Run(w io.Writer) []restify.TestResult {
	io.WriteString(w, fmt.Sprintf(
		"Start running test scenario: name=%s, env=%s, desc=%s, cases=%d\r\n",
		s.name, s.environment, s.description, len(s.cases)))

	testResults := []restify.TestResult{}
	httpClient := http.Client{}

loop:
	for i, tc := range s.cases {
		io.WriteString(w, fmt.Sprintf(
			"%d. Test case: name=%s desc=%s onfail=%s\r\n",
			(i+1), tc.Name, tc.Description, tc.Pipeline.OnFailure))

		tr := restify.NewTestResult(s, i)

		//Parse any cache needed
		tc.Request.Parse(s.cache)
		// tc.Expect.Parse(s.cache)
		payload, _ := json.Marshal(tc.Request.Payload)

		//Setup HTTP request
		req, err := http.NewRequest(tc.Request.Method, tc.Request.URL, bytes.NewBuffer(payload))
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to create request: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to create request: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			continue
		}

		//Add headers to request
		for key, head := range tc.Request.Headers {
			req.Header.Add(key, head)
		}

		//trace HTTP
		var start, connect, dns, tlsHandshake time.Time
		trace := &httptrace.ClientTrace{
			DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
			DNSDone: func(ddi httptrace.DNSDoneInfo) {
				tr.TimingDNS = time.Since(dns)
			},

			TLSHandshakeStart: func() { tlsHandshake = time.Now() },
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				tr.TimingHandshake = time.Since(tlsHandshake)
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				tr.TimingConnected = time.Since(connect)
			},

			GotFirstResponseByte: func() {
				tr.TimingFirstByte = time.Since(start)
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

		//Initiate HTTP request
		start = time.Now()
		res, err := httpClient.Do(req)
		tr.TimingTotal = time.Since(start)
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to execute request: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to execute request: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)
			continue
		}

		tr.ResponseCode = res.StatusCode
		tr.ResponseSize = res.ContentLength

		//Assert response body first so we print the response JSON before anything else
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to get response body: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to get response body: %s\r\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			continue
		} else {
			io.WriteString(w, fmt.Sprintf("%d. Got response body: %s\r\n", (i+1), string(body)))
		}

		//Assert status code
		if res.StatusCode != tc.Expect.StatusCode && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\r\n",
				(i + 1), tc.Expect.StatusCode, res.StatusCode)
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults
		} else if res.StatusCode != tc.Expect.StatusCode {
			msg := fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\r\n",
				(i + 1), tc.Expect.StatusCode, res.StatusCode)
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			continue
		}

		// TODO: Evaluate every rule
		var pair map[string]interface{}
		json.Unmarshal(body, &pair) //	convert []byte to map[string]interface{}

		for _, expr := range tc.Expect.Evaluate { //	foreach rule in evaluate
			isValid := expr.IsTrue(pair)
			if !isValid && tc.Pipeline.OnFailure == onfailure.Exit {
				msg := fmt.Sprintf("%d. Expression Failed : Status %t\r\n", (i + 1), isValid)
				io.WriteString(w, msg)

				tr.Message = msg
				testResults = append(testResults, tr)

				return testResults
			} else if !isValid {
				msg := fmt.Sprintf("%d. Expression Failed : Status %t\r\n", (i + 1), isValid)
				io.WriteString(w, msg)

				tr.Message = msg
				testResults = append(testResults, tr)

				continue loop
			} else {
				msg := fmt.Sprintf("%d. Expression Success: Status %t\r\n", (i + 1), isValid)
				io.WriteString(w, msg)

				continue
			}
		}

		//cache if needed
		if tc.Pipeline.Cache {
			s.cache[tc.Pipeline.CacheAs] = body
		}

		msg := fmt.Sprintf("%d. Success\r\n", (i + 1))
		io.WriteString(w, msg)

		tr.Success = true
		tr.Message = msg
		testResults = append(testResults, tr)
	}

	return testResults
}
