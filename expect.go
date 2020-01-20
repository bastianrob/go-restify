package restify

//Expect response expectation
type Expect struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Evaluate   []Expression      `json:"evaluate"`
}

// TODO: NO LONGER NEEDED
// Parse cache into evaluation value
// This will replace {....} with existing value in cache
// func (e *Expect) Parse(cache map[string]json.RawMessage) {
// 	for i, exp := range e.Evaluate {
// 		matches := replacable.FindAllStringSubmatch(exp.Value, -1)
// 		for _, match := range matches {
// 			param := match[0]
// 			keys := strings.Split(match[1], ".")
// 			cacheKey := keys[0]
// 			cacheProps := keys[1:]

// 			obj := cache[cacheKey]
// 			strval, _ := jsonparser.GetString(obj, cacheProps...)

// 			exp.Value = strings.Replace(exp.Value, param, strval, 1)
// 			e.Evaluate[i] = exp
// 		}
// 	}
// }
