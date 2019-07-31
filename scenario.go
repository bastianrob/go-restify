package restify

import (
	"io"
)

//ScenarioSetter access to set scenario properties
type ScenarioSetter interface {
	ID(id string) ScenarioSetter
	Name(name string) ScenarioSetter
	Description(desc string) ScenarioSetter
	Environment(env string) ScenarioSetter
	AddCase(tcase TestCase) ScenarioSetter
	End() Scenario
}

//ScenarioGetter access to get scenario properties
type ScenarioGetter interface {
	ID() string
	Name() string
	Description() string
	Environment() string
	Cases() []TestCase
}

//Scenario is the biggest scope of a test
//Can have multiple test cases
type Scenario interface {
	Get() ScenarioGetter
	Set() ScenarioSetter
	Run(w io.Writer) ([]TestResult, error)
}
