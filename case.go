package restify

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/bastianrob/go-restify/enum"
)

//Request test object
type Request struct {
	URL     string            `json:"url" bson:"url"`
	Method  string            `json:"method" bson:"method"`
	Headers map[string]string `json:"headers" bson:"headers"`
	Payload json.RawMessage   `json:"payload" bson:"payload"`
}

//Parse cache into request parameter
//This will replace {....} with existing value in cache
func (r *Request) Parse(cache map[string]json.RawMessage) {
	regex := regexp.MustCompile("\\{(.*?)\\}")

	//URL regex
	{
		matches := regex.FindAllStringSubmatch(r.URL, -1)
		for _, match := range matches {
			param := match[0]
			keys := strings.Split(match[1], ".")
			cacheKey := keys[0]
			cacheProps := keys[1:]

			obj := cache[cacheKey]
			strval, _ := jsonparser.GetString(obj, cacheProps...)

			r.URL = strings.Replace(r.URL, param, strval, 1)
		}
	}

	//Headers regex
	for key, head := range r.Headers {
		matches := regex.FindAllStringSubmatch(head, -1)
		for _, match := range matches {
			param := match[0]
			keys := strings.Split(match[1], ".")
			cacheKey := keys[0]
			cacheProps := keys[1:]

			obj := cache[cacheKey]
			strval, _ := jsonparser.GetString(obj, cacheProps...)

			replaced := strings.Replace(head, param, strval, 1)
			r.Headers[key] = replaced
		}
	}

}

//Expression rule of expected response
type Expression struct {
	Prop        string `json:"prop" bson:"prop"`
	Operator    string `json:"operator" bson:"operator"`
	Value       string `json:"value" bson:"value"`
	Description string `json:"description" bson:"description"`
}

//Expect response expectation
type Expect struct {
	StatusCode       int               `json:"status_code" bson:"status_code"`
	Headers          map[string]string `json:"headers" bson:"headers"`
	EvaluationObject string            `json:"evaluation_object" bson:"evaluation_object"`
	Evaluate         []Expression      `json:"evaluate" bson:"evaluate"`
}

//Pipeline test pipeline as what to do with the response object
type Pipeline struct {
	Cache     bool           `json:"cache" bson:"cache"`
	CacheAs   string         `json:"cache_as" bson:"cache_as"`
	OnFailure enum.OnFailure `json:"on_failure" bson:"on_failure"`
}

//TestCase struct
type TestCase struct {
	Order       uint     `json:"order" bson:"order"`
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description" bson:"description"`
	Request     Request  `json:"request" bson:"request"`
	Expect      Expect   `json:"expect" bson:"expect"`
	Pipeline    Pipeline `json:"pipeline" bson:"pipeline"`
}
