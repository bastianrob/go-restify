package restify

import "time"

//TestResult object, all properties must be flat (not nested / denormalized)
//Hopefully de-normalizing the structure can make analytic easier
//Generally immutable, so we don't need ID and just let DB auto gen for us
type TestResult struct {
	Timestamp     int64  `json:"timestamp"`
	ScenarioName  string `json:"scenario_name"`
	TestCaseOrder int    `json:"test_case_order"`
	TestCaseName  string `json:"test_case_name"`
	RequestMethod string `json:"request_method"`
	RequestURL    string `json:"request_url"`
	//RequestPayload?
	ResponseCode int   `json:"response_code"`
	ResponseSize int64 `json:"response_size"`
	//Response timing
	TimingDNS       time.Duration `json:"timing_dns"`
	TimingHandshake time.Duration `json:"timing_handshake"`
	TimingConnected time.Duration `json:"timing_connected"`
	TimingFirstByte time.Duration `json:"timing_first_byte"`
	TimingTotal     time.Duration `json:"timing_total"`
	//ResponsePayload?
	Success      bool   `json:"success"`
	ExpectedCode int    `json:"expected_code"`
	Message      string `json:"message"`
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
