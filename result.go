package restify

import "time"

//TestResult object, all properties must be flat (not nested / denormalized)
//Hopefully de-normalizing the structure can make analytic easier
//Generally immutable, so we don't need ID and just let DB auto gen for us
type TestResult struct {
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
	ScenarioName  string `json:"scenario_name" bson:"scenario_name"`
	TestCaseOrder int    `json:"test_case_order" bson:"test_case_order"`
	TestCaseName  string `json:"test_case_name" bson:"test_case_name"`
	RequestMethod string `json:"request_method" bson:"request_method"`
	RequestURL    string `json:"request_url" bson:"request_url"`
	//RequestPayload?
	ResponseCode int   `json:"response_code" bson:"response_code"`
	ResponseSize int64 `json:"response_size" bson:"response_size"`
	//Response timing
	TimingDNS       float64 `json:"timing_dns" bson:"timing_dns"`
	TimingHandshake float64 `json:"timing_handshake" bson:"timing_handshake"`
	TimingConnected float64 `json:"timing_connected" bson:"timing_connected"`
	TimingFirstByte float64 `json:"timing_first_byte" bson:"timing_first_byte"`
	TimingTotal     float64 `json:"timing_total" bson:"timing_total"`
	//ResponsePayload?
	ExpectedCode int    `json:"expected_code" bson:"expected_code"`
	Message      string `json:"message" bson:"message"`
}

//NewTestResult from scenario
func NewTestResult(scn Scenario, tcn int) TestResult {
	tc := scn.Get().Cases()[tcn]
	return TestResult{
		Timestamp:     time.Now().UnixNano(),
		ScenarioName:  scn.Get().Name(),
		TestCaseOrder: tcn,
		TestCaseName:  tc.Name,
		RequestMethod: tc.Request.Method,
		RequestURL:    tc.Request.URL,
		ExpectedCode:  tc.Expect.StatusCode,
	}
}
