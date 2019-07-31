package scenario

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"go.mongodb.org/mongo-driver/bson"

	restify "github.com/bastianrob/go-restify"
	"github.com/bastianrob/go-restify/enum/onfailure"
	valuator "github.com/bastianrob/go-valuator"
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

func (s *scenario) MarshalBSON() ([]byte, error) {
	return bson.Marshal(struct {
		ID          string             `bson:"_id"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Environment string             `bson:"environment"`
		Cases       []restify.TestCase `bson:"cases"`
	}{
		ID:          s.id,
		Name:        s.name,
		Description: s.description,
		Environment: s.environment,
		Cases:       s.cases,
	})
}

func (s *scenario) UnmarshalBSON(data []byte) error {
	alias := struct {
		ID          string             `bson:"_id"`
		Name        string             `bson:"name"`
		Description string             `bson:"description"`
		Environment string             `bson:"environment"`
		Cases       []restify.TestCase `bson:"cases"`
	}{}

	err := bson.Unmarshal(data, &alias)
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

func (s *scenario) Run(w io.Writer) ([]restify.TestResult, error) {
	io.WriteString(w, fmt.Sprintf(
		"Start running test scenario: name=%s env=%s desc=%s cases=%d\n",
		s.name, s.environment, s.description, len(s.cases)))

	testResults := []restify.TestResult{}
	httpClient := http.Client{}
	for i, tc := range s.cases {
		io.WriteString(w, fmt.Sprintf(
			"%d. Test case: name=%s desc=%s onfail=%s\n",
			(i+1), tc.Name, tc.Description, tc.Pipeline.OnFailure))

		tr := restify.NewTestResult(s, i)

		//Parse any cache needed
		tc.Request.Parse(s.cache)
		tc.Expect.Parse(s.cache)
		payload, _ := json.Marshal(tc.Request.Payload)

		//Setup HTTP request
		req, err := http.NewRequest(tc.Request.Method, tc.Request.URL, bytes.NewBuffer(payload))
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to create request: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults, errors.New(msg)
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to create request: %s\n", (i + 1), err.Error())
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
				tr.TimingDNS = float64(time.Since(dns)) / float64(time.Millisecond)
			},

			TLSHandshakeStart: func() { tlsHandshake = time.Now() },
			TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
				tr.TimingHandshake = float64(time.Since(tlsHandshake)) / float64(time.Millisecond)
			},

			ConnectStart: func(network, addr string) { connect = time.Now() },
			ConnectDone: func(network, addr string, err error) {
				tr.TimingConnected = float64(time.Since(connect)) / float64(time.Millisecond)
			},

			GotFirstResponseByte: func() {
				tr.TimingFirstByte = float64(time.Since(start)) / float64(time.Millisecond)
			},
		}

		req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		//Initiate HTTP request
		res, err := httpClient.Do(req)
		tr.TimingTotal = float64(time.Since(start)) / float64(time.Millisecond)
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to execute request: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults, errors.New(msg)
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to execute request: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)
			continue
		}

		tr.ResponseCode = res.StatusCode
		tr.ResponseSize = res.ContentLength

		//Assert status code
		if res.StatusCode != tc.Expect.StatusCode && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\n",
				(i + 1), tc.Expect.StatusCode, res.StatusCode)
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults, errors.New(msg)
		} else if res.StatusCode != tc.Expect.StatusCode {
			msg := fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\n",
				(i + 1), tc.Expect.StatusCode, res.StatusCode)
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			continue
		}

		//Assert response body
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to get response body: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			return testResults, errors.New(msg)
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to get response body: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)

			continue
		} else {
			io.WriteString(w, fmt.Sprintf("%d. Got response body: %s\n", (i+1), string(body)))
		}

		//parse response body into map
		obj := make(map[string]interface{})
		if tc.Expect.EvaluationObject != "" {
			paths := strings.Split(tc.Expect.EvaluationObject, ".")
			val, _, _, _ := jsonparser.Get(body, paths...)
			err = json.Unmarshal(val, &obj)
		} else {
			err = json.Unmarshal(body, &obj)
		}

		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			msg := fmt.Sprintf("%d. Failed to parse response body into map: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)
			return testResults, errors.New(msg)
		} else if err != nil {
			msg := fmt.Sprintf("%d. Failed to parse response body into map: %s\n", (i + 1), err.Error())
			io.WriteString(w, msg)

			tr.Message = msg
			testResults = append(testResults, tr)
			continue
		}

		//evaluate every rule
		valid := true
		invalidMsg := ""
		for ri, rule := range tc.Expect.Evaluate {
			eval, err := valuator.NewValuator(rule.Prop, rule.Operator, rule.Value, rule.Description)
			if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
				msg := fmt.Sprintf("%d.%d. Failed to get evaluator: %s\n", (i + 1), (ri + 1), err.Error())
				io.WriteString(w, msg)

				tr.Message = msg
				testResults = append(testResults, tr)
				return testResults, errors.New(msg)
			} else if err != nil {
				msg := fmt.Sprintf("%d.%d. Failed to get evaluator: %s\n", (i + 1), (ri + 1), err.Error())
				io.WriteString(w, msg)

				tr.Message = msg
				testResults = append(testResults, tr)
				continue
			}

			valid = eval.Evaluate(obj)
			if !valid {
				invalidMsg = fmt.Sprintf("%d.%d. Expectation failed against rule=%+v\n", (i + 1), (ri + 1), rule)
				io.WriteString(w, invalidMsg)
				break
			}
		}

		if !valid && tc.Pipeline.OnFailure == onfailure.Exit {
			tr.Message = invalidMsg
			testResults = append(testResults, tr)
			return testResults, errors.New(invalidMsg)
		}

		//cache if needed
		if tc.Pipeline.Cache {
			s.cache[tc.Pipeline.CacheAs] = body
		}

		msg := fmt.Sprintf("%d. Success\n", (i + 1))
		io.WriteString(w, msg)
		tr.Message = invalidMsg
		testResults = append(testResults, tr)
	}

	return testResults, nil
}
