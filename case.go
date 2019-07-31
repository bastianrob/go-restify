package restify

import (
	"regexp"

	"github.com/bastianrob/go-restify/enum"
)

var (
	replacable = regexp.MustCompile("\\{(.*?)\\}")
)

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
