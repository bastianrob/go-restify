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

type getter struct {
	scenario *scenario
}

type setter struct {
	scenario *scenario
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

func (g *getter) ID() string {
	return g.scenario.id
}

func (s *setter) ID(id string) restify.ScenarioSetter {
	s.scenario.id = id
	return s
}

func (g *getter) Name() string {
	return g.scenario.name
}

func (s *setter) Name(name string) restify.ScenarioSetter {
	s.scenario.name = name
	return s
}

func (g *getter) Description() string {
	return g.scenario.description
}

func (s *setter) Description(desc string) restify.ScenarioSetter {
	s.scenario.description = desc
	return s
}

func (g *getter) Environment() string {
	return g.scenario.environment
}

func (s *setter) Environment(env string) restify.ScenarioSetter {
	s.scenario.environment = env
	return s
}

func (g *getter) Cases() []restify.TestCase {
	return g.scenario.cases
}

func (s *setter) AddCase(tcase restify.TestCase) restify.ScenarioSetter {
	s.scenario.cases = append(s.scenario.cases, tcase)
	sort.Slice(s.scenario.cases, func(i, j int) bool {
		return s.scenario.cases[i].Order < s.scenario.cases[j].Order
	})

	return s
}

func (s *setter) End() restify.Scenario {

	return s.scenario
}

func (s *scenario) Get() restify.ScenarioGetter {
	return s.getter
}

func (s *scenario) Set() restify.ScenarioSetter {
	return s.setter
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
