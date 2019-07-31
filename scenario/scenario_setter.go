package scenario

import (
	"sort"

	restify "github.com/bastianrob/go-restify"
)

type setter struct {
	scenario *scenario
}

func (s *scenario) Set() restify.ScenarioSetter {
	return s.setter
}

func (s *setter) ID(id string) restify.ScenarioSetter {
	s.scenario.id = id
	return s
}

func (s *setter) Name(name string) restify.ScenarioSetter {
	s.scenario.name = name
	return s
}

func (s *setter) Description(desc string) restify.ScenarioSetter {
	s.scenario.description = desc
	return s
}

func (s *setter) Environment(env string) restify.ScenarioSetter {
	s.scenario.environment = env
	return s
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
