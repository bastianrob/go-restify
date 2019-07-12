package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/bastianrob/go-restify/enum/onfailure"
	"github.com/bastianrob/go-restify/testcase"
	valuator "github.com/bastianrob/go-valuator"
)

//Scenario is the biggest scope of a test
//Can have multiple test cases
type Scenario interface {
	Name(name string) Scenario
	Description(desc string) Scenario
	Environment(env string) Scenario
	AddCase(tcase testcase.TestCase) Scenario
	Run(w io.Writer)
}

type scenario struct {
	cache map[string]json.RawMessage

	name        string
	description string
	environment string

	cases []testcase.TestCase
}

//New Scenario create a new test scenario
func New() Scenario {
	return &scenario{
		cache: make(map[string]json.RawMessage),
		cases: []testcase.TestCase{},
	}
}

func (s *scenario) Name(name string) Scenario {
	s.name = name
	return s
}

func (s *scenario) Description(desc string) Scenario {
	s.description = desc
	return s
}

func (s *scenario) Environment(env string) Scenario {
	s.environment = env
	return s
}

func (s *scenario) AddCase(tcase testcase.TestCase) Scenario {
	s.cases = append(s.cases, tcase)
	sort.Slice(s.cases, func(i, j int) bool {
		return s.cases[i].Order < s.cases[j].Order
	})

	return s
}

func (s *scenario) Run(w io.Writer) {
	io.WriteString(w, fmt.Sprintf(
		"Start running test scenario: name=%s env=%s desc=%s cases=%d\n",
		s.name, s.environment, s.description, len(s.cases)))

	httpClient := http.Client{}
	for i, tc := range s.cases {
		io.WriteString(w, fmt.Sprintf(
			"%d. Test case: name=%s desc=%s onfail=%s\n",
			(i+1), tc.Name, tc.Description, tc.Pipeline.OnFailure))

		//Parse any cache needed
		tc.Request.Parse(s.cache)

		//Setup HTTP request
		req, err := http.NewRequest(tc.Request.Method, tc.Request.URL, bytes.NewBuffer(tc.Request.Payload))
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			io.WriteString(w, fmt.Sprintf("%d. Failed to create request: %s\n", (i+1), err.Error()))
			return
		} else if err != nil {
			io.WriteString(w, fmt.Sprintf("%d. Failed to create request: %s\n", (i+1), err.Error()))
			continue
		}

		//Add headers to request
		for key, head := range tc.Request.Headers {
			req.Header.Add(key, head)
		}

		//Initiate HTTP request
		res, err := httpClient.Do(req)
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			io.WriteString(w, fmt.Sprintf("%d. Failed to execute request: %s\n", (i+1), err.Error()))
			return
		} else if err != nil {
			io.WriteString(w, fmt.Sprintf("%d. Failed to execute request: %s\n", (i+1), err.Error()))
			continue
		}

		//Assert status code
		if res.StatusCode != tc.Expect.StatusCode && tc.Pipeline.OnFailure == onfailure.Exit {
			io.WriteString(w, fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\n",
				(i+1), tc.Expect.StatusCode, res.StatusCode))
			return
		} else if res.StatusCode != tc.Expect.StatusCode {
			io.WriteString(w, fmt.Sprintf(
				"%d. Expectation failed. Expect status: %d, got: %d\n",
				(i+1), tc.Expect.StatusCode, res.StatusCode))
			continue
		}

		//Assert response body
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
			io.WriteString(w, fmt.Sprintf("%d. Failed to get response body: %s\n", (i+1), err.Error()))
			return
		} else if err != nil {
			io.WriteString(w, fmt.Sprintf("%d. Failed to get response body: %s\n", (i+1), err.Error()))
			continue
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
			io.WriteString(w, fmt.Sprintf("%d. Failed to parse response body into map: %s\n", (i+1), err.Error()))
			return
		} else if err != nil {
			io.WriteString(w, fmt.Sprintf("%d. Failed to parse response body into map: %s\n", (i+1), err.Error()))
			continue
		}

		//evaluate every rule
		valid := true
		for ri, rule := range tc.Expect.Evaluate {
			eval, err := valuator.NewValuator(rule.Prop, rule.Operator, rule.Value, rule.Description)
			if err != nil && tc.Pipeline.OnFailure == onfailure.Exit {
				io.WriteString(w, fmt.Sprintf("%d.%d. Failed to get evaluator: %s\n", (i+1), (ri+1), err.Error()))
				return
			} else if err != nil {
				io.WriteString(w, fmt.Sprintf("%d.%d. Failed to get evaluator: %s\n", (i+1), (ri+1), err.Error()))
				continue
			}

			valid = eval.Evaluate(obj)
			if !valid {
				io.WriteString(w, fmt.Sprintf("%d.%d. Expectation failed against rule=%+v\n",
					(i+1), (ri+1), rule))
				break
			}
		}

		if !valid && tc.Pipeline.OnFailure == onfailure.Exit {
			return
		}

		//cache if needed
		if tc.Pipeline.Cache {
			s.cache[tc.Pipeline.CacheAs] = body
		}

		io.WriteString(w, fmt.Sprintf("%d. Success\n", (i+1)))
	}
}
