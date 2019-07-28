package restify

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/bastianrob/go-restify/enum"
)

var (
	replacable = regexp.MustCompile("\\{(.*?)\\}")
)

//Request test object
type Request struct {
	URL     string                 `json:"url" bson:"url"`
	Method  string                 `json:"method" bson:"method"`
	Headers map[string]string      `json:"headers" bson:"headers"`
	Payload map[string]interface{} `json:"payload" bson:"payload"`
}

//Parse cache into request parameter
//This will replace {....} with existing value in cache
func (r *Request) Parse(cache map[string]json.RawMessage) {
	//URL regex
	{
		matches := replacable.FindAllStringSubmatch(r.URL, -1)
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
		matches := replacable.FindAllStringSubmatch(head, -1)
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

	//Payload Regex
	r.Payload = recursiveMapParser(r.Payload, cache)
}

func recursiveMapParser(obj map[string]interface{}, cache map[string]json.RawMessage) map[string]interface{} {
	for key, val := range obj {
		if val == nil {
			continue
		}

		valKind := reflect.TypeOf(val).Kind()
		if valKind == reflect.Map {
			obj[key] = recursiveMapParser(val.(map[string]interface{}), cache)
			continue
		} else if valKind == reflect.String {
			str := val.(string)
			matches := replacable.FindAllStringSubmatch(str, -1)
			for _, match := range matches {
				param := match[0]
				keys := strings.Split(match[1], ".")
				cacheKey := keys[0]
				cacheProps := keys[1:]

				cacheObj := cache[cacheKey]
				strval, _ := jsonparser.GetString(cacheObj, cacheProps...)

				replaced := strings.Replace(str, param, strval, 1)
				obj[key] = replaced
			}
			continue
		}

	}
	return obj
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
