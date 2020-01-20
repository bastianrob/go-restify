package restify

import (
	"github.com/robertkrimen/otto"
)

//Expression rule of expected response
type Expression string

// IsTrue evaluate given boolean expression whether it is true, or false
// Return false on invalid expression
func (expr Expression) IsTrue(input map[string]interface{}) bool {
	jsvm := otto.New()
	for key, val := range input {
		jsvm.Set(key, val)
	}

	// jsvm.Set("res", input)
	// jsvm.Run("res = JSON.parse(res)")
	val, err := jsvm.Run(string(expr))
	if err != nil {
		return false
	}

	result, err := val.ToBoolean()
	if err != nil {
		return false
	}

	return result
}
