package restify

import (
	"encoding/json"
	"strings"

	"github.com/buger/jsonparser"
)

//Expect response expectation
type Expect struct {
	StatusCode       int               `json:"status_code" bson:"status_code"`
	Headers          map[string]string `json:"headers" bson:"headers"`
	EvaluationObject string            `json:"evaluation_object" bson:"evaluation_object"`
	Evaluate         []Expression      `json:"evaluate" bson:"evaluate"`
}

//Parse cache into evaluation value
//This will replace {....} with existing value in cache
func (e *Expect) Parse(cache map[string]json.RawMessage) {
	for i, exp := range e.Evaluate {
		matches := replacable.FindAllStringSubmatch(exp.Value, -1)
		for _, match := range matches {
			param := match[0]
			keys := strings.Split(match[1], ".")
			cacheKey := keys[0]
			cacheProps := keys[1:]

			obj := cache[cacheKey]
			strval, _ := jsonparser.GetString(obj, cacheProps...)

			exp.Value = strings.Replace(exp.Value, param, strval, 1)
			e.Evaluate[i] = exp
		}
	}
}
