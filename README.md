# go-validator

This package helps you automate API testing

## Limitation

Only works for JSON API

## Dependencies

> "github.com/bastianrob/go-valuator"
> "github.com/buger/jsonparser"

## Example

### Setting up test scenario

```go
buffer := bytes.Buffer{}
scenario.New().
    Name("Scenario One").
    Environment("Local").
    Description("").
    AddCase(testcase.TestCase{
        Order:       1,
        Name:        "Setup Auth",
        Description: "Auth to firebase",
        Request: testcase.Request{
            URL:    "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key={YOUR_FIREBASE_KEY}",
            Method: "POST",
            Payload: json.RawMessage(`{
                "email": "your.user@email.com",
                "password": "your.password",
                "returnSecureToken": true
            }`),
        },
        Expect: testcase.Expect{
            StatusCode: 200,
            Evaluate: []testcase.Expression{{
                Prop:        "idToken",
                Operator:    "!=",
                Value:       "",
                Description: "ID Token must be returned from firebase",
            }},
        },
        Pipeline: testcase.Pipeline{
            Cache:     true,
            CacheAs:   "auth",
            OnFailure: onfailure.Exit,
        },
    }).
    AddCase(testcase.TestCase{
        Order:       2,
        Name:        "Get List",
        Description: "Get a list of resource that returns {data: [{resource1, resource2}]}",
        Request: testcase.Request{
            URL:    "http://localhost:3000/resources",
            Method: "GET",
            Headers: map[string]string{
                "Authorization": "Bearer {auth.idToken}",
            },
            Payload: nil,
        },
        Expect: testcase.Expect{
            StatusCode:       200,
            EvaluationObject: "data.[0]",
            Evaluate: []testcase.Expression{{
                Prop:        "id",
                Operator:    "!=",
                Value:       "",
                Description: "Returned data is not empty",
            }},
        },
        Pipeline: testcase.Pipeline{
            Cache:     true,
            CacheAs:   "R1",
            OnFailure: onfailure.Exit,
        },
    }).
    AddCase(testcase.TestCase{
        Order:       3,
        Name:        "Get One",
        Description: "Get one resource from previous case",
        Request: testcase.Request{
            URL:    "http://localhost:3000/resources/{R1.data.[0].id}",
            Method: "GET",
            Headers: map[string]string{
                "Authorization": "Bearer {auth.idToken}",
            },
            Payload: nil,
        },
        Expect: testcase.Expect{
            StatusCode:       200,
            EvaluationObject: "data",
            Evaluate: []testcase.Expression{{
                Prop:        "id",
                Operator:    "!=",
                Value:       "",
                Description: "Returned data is not empty",
            }},
        },
        Pipeline: testcase.Pipeline{
            Cache:     true,
            CacheAs:   "R2",
            OnFailure: onfailure.Exit,
        },
    }).
    Run(&buffer)
```
