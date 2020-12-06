# go-restify

This package helps you automate API testing

## Limitation

Only works for JSON API

## Dependencies

```bash
"github.com/robertkrimen/otto"
"github.com/buger/jsonparser"
```

## Example

### Setting up test scenario

```go
buffer := bytes.Buffer{}
scenario.New().Set().
    Name("Scenario One").
    Environment("Local").
    Description("").
    AddCase(restify.TestCase{
        Order:       1,
        Name:        "Setup Auth",
        Description: "Auth to firebase",
        Request: restify.Request{
            URL:    "https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key={YOUR_FIREBASE_KEY}",
            Method: "POST",
            Payload: json.RawMessage(`{
                "email": "your.user@email.com",
                "password": "your.password",
                "returnSecureToken": true
            }`),
        },
        Expect: restify.Expect{
            StatusCode: 200,
            Evaluate: []restify.Expression{{
                Prop:        "idToken",
                Operator:    "!=",
                Value:       "",
                Description: "ID Token must be returned from firebase",
            }},
        },
        Pipeline: restify.Pipeline{
            Cache:     true,
            CacheAs:   "auth",
            OnFailure: onfailure.Exit,
        },
    }).
    AddCase(restify.TestCase{
        Order:       2,
        Name:        "Get List",
        Description: "Get a list of resource that returns {data: [{resource1, resource2}]}",
        Request: restify.Request{
            URL:    "http://localhost:3000/resources",
            Method: "GET",
            Headers: map[string]string{
                "Authorization": "Bearer {auth.idToken}",
            },
            Payload: nil,
        },
        Expect: restify.Expect{
            StatusCode:       200,
            Evaluate: []restify.Expression{{
                Object:      "data.[0]",
                Prop:        "id",
                Operator:    "!=",
                Value:       "",
                Description: "Returned data is not empty",
            }},
        },
        Pipeline: restify.Pipeline{
            Cache:     true,
            CacheAs:   "R1",
            OnFailure: onfailure.Exit,
        },
    }).
    AddCase(restify.TestCase{
        Order:       3,
        Name:        "Get One",
        Description: "Get one resource from previous case",
        Request: restify.Request{
            URL:    "http://localhost:3000/resources/{R1.data.[0].id}",
            Method: "GET",
            Headers: map[string]string{
                "Authorization": "Bearer {auth.idToken}",
            },
            Payload: nil,
        },
        Expect: restify.Expect{
            StatusCode:       200,
            Evaluate: []restify.Expression{{
                Object:      "data",
                Prop:        "id",
                Operator:    "!=",
                Value:       "",
                Description: "Returned data is not empty",
            }},
        },
        Pipeline: restify.Pipeline{
            Cache:     true,
            CacheAs:   "R2",
            OnFailure: onfailure.Exit,
        },
    }).End().
    Run(&buffer)
```
