package restify

//Expression rule of expected response
type Expression struct {
	Object      string `json:"object"`
	Prop        string `json:"prop"`
	Operator    string `json:"operator"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
