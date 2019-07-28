package scenario

import restify "github.com/bastianrob/go-restify"

func (s *scenario) Get() restify.ScenarioGetter {
	return s.getter
}

func (g *getter) ID() string {
	return g.scenario.id
}

func (g *getter) Name() string {
	return g.scenario.name
}

func (g *getter) Description() string {
	return g.scenario.description
}

func (g *getter) Environment() string {
	return g.scenario.environment
}

func (g *getter) Cases() []restify.TestCase {
	return g.scenario.cases
}
