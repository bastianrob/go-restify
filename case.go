package restify

import (
	"regexp"

	"github.com/Spacestock/go-restify/enum"
)

var (
	replacable = regexp.MustCompile("\\{(.*?)\\}")
)

//Pipeline test pipeline as what to do with the response object
type Pipeline struct {
	Cache     bool           `json:"cache"`
	CacheAs   string         `json:"cache_as"`
	OnFailure enum.OnFailure `json:"on_failure"`
}

//TestCase struct
type TestCase struct {
	Order       uint     `json:"order"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Request     Request  `json:"request"`
	Expect      Expect   `json:"expect"`
	Pipeline    Pipeline `json:"pipeline"`
}
