package restify

import (
	"encoding/json"
	"testing"
)

func TestExpression_IsTrue(t *testing.T) {
	type args struct {
		json string
		pair map[string]interface{}
	}
	tests := []struct {
		name string
		expr Expression
		args args
		want bool
	}{{
		name: "Simple positive case 1",
		args: args{json: `{"name": "Mr. Brother"}`}, //We receive this from API call
		expr: Expression(`name === "Mr. Brother"`),  //We want to test the response with this expression
		want: true,                                  //Given such response, against the expr rule, we want the expression result to be true
	}, {
		name: "Simple positive case 2",
		args: args{json: `{"age": 10}`},
		expr: Expression(`age >= 10`),
		want: true,
	}, {
		name: "Composite positive case 3",
		args: args{json: `{"name": "John", "age": 10}`},
		expr: Expression(`name === "John" && age >= 10`),
		want: true,
	}, {
		name: "Complex object positive case 3",
		args: args{json: `{"person": {"name": "John", "age": 10} }`},
		expr: Expression(`person.name === "John" && person.age >= 10`),
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := map[string]interface{}{}
			json.Unmarshal([]byte(tt.args.json), &pair)

			if got := tt.expr.IsTrue(pair); got != tt.want {
				t.Errorf("Expression.IsTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}
