package restify

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/buger/jsonparser"
)

//Request test object
type Request struct {
	URL     string                 `json:"url" bson:"url"`
	Method  string                 `json:"method" bson:"method"`
	Headers map[string]string      `json:"headers" bson:"headers"`
	Payload map[string]interface{} `json:"payload" bson:"payload"`
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
