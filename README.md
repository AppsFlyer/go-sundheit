# af-go-health
A library built to provide support for defining service health for golang services.
It allows you to register async health checks for your dependencies and the service itself, 
and provides a health endpoint that exposes their status.

## Installation

```
dep ensure -add gitlab.appsflyer.com/Architecture/af-go-health@v0.0.1
```

Or add this to your Gopkg.toml:
```go
[[constraint]]
  name = "gitlab.appsflyer.com/Architecture/af-go-health"
  version = "0.0.1"
```

## Usage
```go
import (
	"net/http"
	"time"
	"log"

	"github.com/pkg/errors"
	"gitlab.appsflyer.com/Architecture/af-go-health"

	healthhttp "gitlab.appsflyer.com/Architecture/af-go-health/http"
	"gitlab.appsflyer.com/Architecture/af-go-health/checks"
)

func main() {
	// create a new health instance
	h := health.New()
	
	// define an HTTP dependency check
	httpCheckConf := &checks.HTTPCheckConfig{
		CheckName: "httpbin.url.check",
		Timeout:   1 * time.Second,
		// dependency you're checking - use your own URL here...
		// this URL will fail 50% of the times
		URL:       "http://httpbin.org/status/200,300",
	}
	// create the HTTP check for the dependency
	// fail fast when you misconfigured the URL. Don't ignore errors!!!
	httpCheck, err := checks.NewHTTPCheck(httpCheckConf)
	if err == nil {
		err = h.RegisterCheck(&health.Config{
			Check:           httpCheck,
			InitialDelay:    time.Second,      // the check will run once after 1 sec
			ExecutionPeriod: 10 * time.Second, // the check will be executed every 10 sec
		})
		
		if (err != nil) {
			fmt.Println("Failed to register check: ", err)
			return // or whatever
		}
	} else {
		fmt.Println(err)
		return // your call...
	}

  // define more checks...
  
  // register a health endpoint
  http.Handle("/_/health.json", healthhttp.HandleHealthJSON(h))
	
	// serve HTTP
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Custom Checks
The library provides 2 means of defining a custom check.
The bottom line is that you need an implementation of the `checks.Check` interface:
```go
// Check is the API for defining health checks.
// A valid check has a non empty Name() and a check (Execute()) function.
type Check interface {
	// Name is the name of the check.
	// Check names must be metric compatible.
	Name() string
	// Execute runs a single time check, and returns an error when the check fails, and an optional details object.
	Execute() (details interface{}, err error)
}
```
See examples in the following 2 sections below.

#### Use the CustomCheck struct
The `checksCustomCheck` struct implements the `checks.Check` interface,
and is the simplest way to implement a check if all you need is to define a check function.

Let's define a check function that fails 50% of the times:
```go
func lotteryCheck() (details interface{}, err error) {
	lottery := rand.Float32()
	details = fmt.Sprintf("lottery=%f", lottery)
	if lottery < 0.5 {
		err = errors.New("Sorry, I failed")
	}
	return
}
```

Now we register the check to start running right away, and execute once per 2 minutes:
```go
h := health.New()
...

h.RegisterCheck(&health.Config{
  Check: &checks.CustomCheck{
    CheckName: "lottery.check",
    CheckFunc: lotteryCheck,
  },
  InitialDelay:    0,
  ExecutionPeriod: 2 * time.Minute,
})
```

#### Implement the Check interface
Sometimes you need to define a more elaborate custom check.
For example when you need to manage state.
For these cases it's best to implement the `checks.Check` interface yourself.

Let's define a flexible example of the lottery check, that allows you to define a fail probability:
```go
type Lottery struct {
	myname string
	probability float32
}

func (l Lottery) Execute() (details interface{}, err error) {
	lottery := rand.Float32()
	details = fmt.Sprintf("lottery=%f", lottery)
	if lottery < l.probability {
		err = errors.New("Sorry, I failed")
	}
	return
}

func (l Lottery) Name() string {
	return l.myname
}
```

And register our custom check, scheduling it to run after 1 sec, and every 30 sec:
```go
h := health.New()
...

h.RegisterCheck(&health.Config{
  Check: Lottery{myname: "custom.lottery.check", probability:0.3,},
  InitialDelay: 1*time.Second,
  ExecutionPeriod: 30*time.Second,
})
```

#### Custom Checks Notes
1. If a check take longer than the specified rate period, then next execution will be delayed, 
but will not be concurrently executed.
1. Checks must complete within a reasonable time. If a check doesn't complete or gets hung, 
the next check execution will be delayed. Use proper time outs.
1. **A health-check name must be a metric name compatible string** 
  (i.e. no funky characters, and spaces allowed - just make it simple like `clicks-db-check`).
  See here: https://help.datadoghq.com/hc/en-us/articles/203764705-What-are-valid-metric-names-

### Expose Health Endpoint
The library provides an HTTP handler function for serving health stats in JSON format.
You can register it using your favorite HTTP implementation like so:
```go
http.Handle("/_/health.json", healthhttp.HandleHealthJSON(h))
```
The endpoint can be called like so:
```text
~ $ curl -i http://localhost:8080/_/health.json
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
Date: Tue, 22 Jan 2019 09:31:46 GMT
Content-Length: 701

{
	"custom.lottery.check": {
		"message": "lottery=0.206583",
		"error": {
			"message": "Sorry, I failed"
		},
		"timestamp": "2019-01-22T11:31:44.632415432+02:00",
		"num_failures": 2,
		"first_failure_time": "2019-01-22T11:31:41.632400256+02:00"
	},
	"lottery.check": {
		"message": "lottery=0.865335",
		"timestamp": "2019-01-22T11:31:44.63244047+02:00",
		"num_failures": 0,
		"first_failure_time": null
	},
	"url.check": {
		"message": "http://httpbin.org/status/200,300",
		"error": {
			"message": "unexpected status code: '300' expected: '200'"
		},
		"timestamp": "2019-01-22T11:31:44.632442937+02:00",
		"num_failures": 4,
		"first_failure_time": "2019-01-22T11:31:38.632485339+02:00"
	}
}
```
Or for the shorter version:
```text
~ $ curl -i http://localhost:8080/_/health.json?type=short
HTTP/1.1 503 Service Unavailable
Content-Type: application/json
Date: Tue, 22 Jan 2019 09:40:19 GMT
Content-Length: 105

{
	"custom.lottery.check": "PASS",
	"lottery.check": "PASS",
	"my.check": "FAIL",
	"url.check": "PASS"
}
```

The `short` reposonse type is suitable for the consul health checks / LB heath checks.

The response code is `200` when the tests pass, and `503` when they fail.

## Metrics
***Not implemented yet!!!***

The library exposes the following metrics for your convenience:
* `health.status.allChecks` - An aggregated health status gauge (0/1 for fail/pass) at the time of sampling. 
This value may be using the cached checks results for async checks. 
* `health.executeTime.allChecks` - The time it took to execute all checks. 
Note that async checks aren't really being run when the health API is queried.
* `health.executeTime.<check-name>` - The time it took to execute a checks.
* `health.failures.<check-name>` - Fail rate. This value depends on the scheduling rate for async checks, or the querying rate for sync checks.
* `health.status.<check-name>` - The health status gauge (0/1 for fail/pass) at the time of sampling. 

## Migration Process
Don't forget to remove the previous health endpoint, and to fix the health check in OneBar to point to the new API, 
e.g. `http://localhost:8080/_/health.json`