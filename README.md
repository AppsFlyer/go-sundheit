# go-sundheit
[![Actions Status](https://github.com/AppsFlyer/go-sundheit/workflows/go-build/badge.svg)](https://github.com/AppsFlyer/go-sundheit/actions)
[![CircleCI](https://circleci.com/gh/AppsFlyer/go-sundheit.svg?style=svg)](https://circleci.com/gh/AppsFlyer/go-sundheit)
[![Coverage Status](https://coveralls.io/repos/github/AppsFlyer/go-sundheit/badge.svg?branch=master)](https://coveralls.io/github/AppsFlyer/go-sundheit?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/AppsFlyer/go-sundheit)](https://goreportcard.com/report/github.com/AppsFlyer/go-sundheit)
[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://godoc.org/github.com/AppsFlyer/go-sundheit)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

<img align="right" src="docs/go-sundheit.png" width="200">

A library built to provide support for defining service health for golang services.
It allows you to register async health checks for your dependencies and the service itself, 
and provides a health endpoint that exposes their status.

## What's go-sundheit?
The project is named after the German word `Gesundheit` which means ‘health’, and it is pronounced `/ɡəˈzʊntˌhaɪ̯t/`.

## Installation
Using go modules:
```
go get github.com/AppsFlyer/go-sundheit@v0.5.0
```

## Usage
```go
import (
	"net/http"
	"time"
	"log"

	"github.com/pkg/errors"
	"github.com/AppsFlyer/go-sundheit"

	healthhttp "github.com/AppsFlyer/go-sundheit/http"
	"github.com/AppsFlyer/go-sundheit/checks"
)

func main() {
	// create a new health instance
	h := gosundheit.New()
	
	// define an HTTP dependency check
	httpCheckConf := checks.HTTPCheckConfig{
		CheckName: "httpbin.url.check",
		Timeout:   1 * time.Second,
		// dependency you're checking - use your own URL here...
		// this URL will fail 50% of the times
		URL:       "http://httpbin.org/status/200,300",
	}
	// create the HTTP check for the dependency
	// fail fast when you misconfigured the URL. Don't ignore errors!!!
	httpCheck, err := checks.NewHTTPCheck(httpCheckConf)
	if err != nil {
		fmt.Println(err)
		return // your call...
	}

	// Alternatively panic when creating a check fails
	httpCheck = checks.Must(checks.NewHTTPCheck(httpCheckConf))

	err = h.RegisterCheck(
		httpCheck,
		gosundheit.InitialDelay(time.Second),         // the check will run once after 1 sec
		gosundheit.ExecutionPeriod(10 * time.Second), // the check will be executed every 10 sec
	)
	
	if err != nil {
		fmt.Println("Failed to register check: ", err)
		return // or whatever
	}

	// define more checks...
	
	// register a health endpoint
	http.Handle("/admin/health.json", healthhttp.HandleHealthJSON(h))
	
	// serve HTTP
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
### Using `Option` to Configure `Health` Service
To create a health service, it's simple as calling the following code:
```go
gosundheit.New(options ...Option)
```
The optional parameters of `options` allows the user to configure the Health Service by passing configuration functions (implementing `Option` signature).    
All options are marked with the prefix `WithX`. Available options:
- `WithCheckListeners` - enables you to act on check registration, start and completed events
- `WithHealthListeners` - enables you to act on changes in the health service results

### Built-in Checks
The library comes with a set of built-in checks.
Currently implemented checks are as follows:

#### HTTP built-in check
The HTTP check allows you to trigger an HTTP request to one of your dependencies, 
and verify the response status, and optionally the content of the response body.
Example was given above in the [usage](#usage) section

#### DNS built-in check(s)
The DNS checks allow you to perform lookup to a given hostname / domain name / CNAME / etc, 
and validate that it resolves to at least the minimum number of required results.

Creating a host lookup check is easy:
```go
// Schedule a host resolution check for `example.com`, requiring at least one results, and running every 10 sec
h.RegisterCheck(
	checks.NewHostResolveCheck("example.com", 1),
	gosundheit.ExecutionPeriod(10 * time.Second),
)
```

You may also use the low level `checks.NewResolveCheck` specifying a custom `LookupFunc` if you want to to perform other kinds of lookups.
For example you may register a reverse DNS lookup check like so:
```go
func ReverseDNLookup(ctx context.Context, addr string) (resolvedCount int, err error) {
	names, err := net.DefaultResolver.LookupAddr(ctx, addr)
	resolvedCount = len(names)
	return
}

//...

h.RegisterCheck(
	checks.NewResolveCheck(ReverseDNLookup, "127.0.0.1", 3),
	gosundheit.ExecutionPeriod(10 * time.Second),
	gosundheit.ExecutionTimeout(1*time.Second)
)
```

#### Ping built-in check(s)
The ping checks allow you to verifies that a resource is still alive and reachable.
For example, you can use it as a DB ping check (`sql.DB` implements the Pinger interface):
```go
	db, err := sql.Open(...)
	dbCheck, err := checks.NewPingCheck("db.check", db)
	_ = h.RegisterCheck(&gosundheit.Config{
		Check: dbCheck,
		// ...
	})
```

You can also use the ping check to test a generic connection like so:
```go
	pinger := checks.NewDialPinger("tcp", "example.com")
	pingCheck, err := checks.NewPingCheck("example.com.reachable", pinger)
	h.RegisterCheck(pingCheck)
``` 

The `NewDialPinger` function supports all the network/address parameters supported by the `net.Dial()` function(s)

### Custom Checks
The library provides 2 means of defining a custom check.
The bottom line is that you need an implementation of the `Check` interface:
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

Now we register the check to start running right away, and execute once per 2 minutes with a timeout of 5 seconds:
```go
h := gosundheit.New()
...

h.RegisterCheck(
	&checks.CustomCheck{
		CheckName: "lottery.check",
		CheckFunc: lotteryCheck,
	},
	gosundheit.InitialDelay(0),
	gosundheit.ExecutionPeriod(2 * time.Minute), 
	gosundheit.ExecutionTimeout(5 * time.Second)
)
```

#### Implement the Check interface
Sometimes you need to define a more elaborate custom check.
For example when you need to manage state.
For these cases it's best to implement the `Check` interface yourself.

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

And register our custom check, scheduling it to run every 30 seconds (after a 1 second initial delay) with a 5 seconds timeout:
```go
h := gosundheit.New()
...

h.RegisterCheck(
	Lottery{myname: "custom.lottery.check", probability:0.3},
	gosundheit.InitialDelay(1*time.Second),
	gosundheit.ExecutionPeriod(30*time.Second),
	gosundheit.ExecutionTimeout(5*time.Second),
)
```

#### Custom Checks Notes
1. If a check take longer than the specified rate period, then next execution will be delayed, 
but will not be concurrently executed.
1. Checks must complete within a reasonable time. If a check doesn't complete or gets hung, 
the next check execution will be delayed. Use proper time outs.
1. Checks must respect the provided context. Specifically, a check must abort its execution, and return an error, if the context has been cancelled.  
1. **A health-check name must be a metric name compatible string** 
  (i.e. no funky characters, and spaces allowed - just make it simple like `clicks-db-check`).
  See here: https://help.datadoghq.com/hc/en-us/articles/203764705-What-are-valid-metric-names-

### Expose Health Endpoint
The library provides an HTTP handler function for serving health stats in JSON format.
You can register it using your favorite HTTP implementation like so:
```go
http.Handle("/admin/health.json", healthhttp.HandleHealthJSON(h))
```
The endpoint can be called like so:
```text
~ $ curl -i http://localhost:8080/admin/health.json
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
~ $ curl -i http://localhost:8080/admin/health.json?type=short
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

The `short` response type is suitable for the consul health checks / LB heath checks.

The response code is `200` when the tests pass, and `503` when they fail.

### CheckListener
It is sometimes desired to keep track of checks execution and apply custom logic.
For example, you may want to add logging, or external metrics to your checks, 
or add some trigger some recovery logic when a check fails after 3 consecutive times.

The `gosundheit.CheckListener` interface allows you to hook this custom logic.

For example, lets add a logging listener to our health repository:
```go
type checkEventsLogger struct{}

func (l checkEventsLogger) OnCheckRegistered(name string, res gosundheit.Result) {
	log.Printf("Check %q registered with initial result: %v\n", name, res)
}

func (l checkEventsLogger) OnCheckStarted(name string) {
	log.Printf("Check %q started...\n", name)
}

func (l checkEventsLogger) OnCheckCompleted(name string, res gosundheit.Result) {
	log.Printf("Check %q completed with result: %v\n", name, res)
}
```

To register your listener:
```go
h := gosundheit.New(gosundheit.WithCheckListeners(&checkEventsLogger))
```

Please note that your `CheckListener` implementation must not block!

### HealthListener
It is something desired to track changes in registered checks results.
For example, you may want to log the amount of results monitored, or send metrics on these results.

The `gosundheit.HealthListener` interface allows you to hook this custom logic.

For example, lets add a logging listener:
```go
type healthLogger struct{}

func (l healthLogger) OnResultsUpdated(results map[string]Result) {
	log.Printf("There are %d results, general health is %t\n", len(results), allHealthy(results))
}
```

To register your listener:
```go
h := gosundheit.New(gosundheit.WithHealthListeners(&checkHealthLogger))
```

## Metrics
The library can expose metrics using a `CheckListener`. At the moment, OpenCensus is available and exposes the following metrics:
* `health/check_status_by_name` - An aggregated health status gauge (0/1 for fail/pass) at the time of sampling.
The aggregation uses the following tags:
  * `check=allChecks`     - all checks aggregation
  * `check=<check-name>`  - specific check aggregation
*  `health/check_count_by_name_and_status` - Aggregated pass/fail counts for checks, with the following tags: 
   * `check=allChecks`     - all checks aggregation
   * `check=<check-name>`  - specific check aggregation
   * `check-passing=[true|false]` 
* `health/executeTime` - The time it took to execute a checks. Using the following tag:
  * `check=<check-name>`  - specific check aggregation


The views can be registered like so:
```go
import (
	"github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/opencensus"
	"go.opencensus.io/stats/view"
)
// This listener can act both as check and health listener for reporting metrics
oc := opencensus.NewMetricsListener()
h := gosundheit.New(gosundheit.WithCheckListeners(oc), gosundheit.WithHealthListeners(oc))
// ...
view.Register(opencensus.DefaultHealthViews...)
// or register individual views. For example:
view.Register(opencensus.ViewCheckExecutionTime, opencensus.ViewCheckStatusByName, ...)
```

### Classification

It is sometimes required to report metrics for different check types (e.g. setup, liveness, readiness).
To report metrics using `classification` tag - it's possible to initialize the OpenCensus listener with classification:

```go
// startup
opencensus.NewMetricsListener(opencensus.WithStartupClassification())
// liveness
opencensus.NewMetricsListener(opencensus.WithLivenessClassification())
// readiness
opencensus.NewMetricsListener(opencensus.WithReadinessClassification())
// custom
opencensus.NewMetricsListener(opencensus.WithClassification("custom"))
```
