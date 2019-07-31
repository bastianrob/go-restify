package restify

//Expression rule of expected response
type Expression struct {
	Prop        string `json:"prop" bson:"prop"`
	Operator    string `json:"operator" bson:"operator"`
	Value       string `json:"value" bson:"value"`
	Description string `json:"description" bson:"description"`
}
