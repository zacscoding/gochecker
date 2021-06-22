:heavy_check_mark: Go Health Checker
===  

Simple health check library for ur Go application :)

# Installation  

```shell
$ go get github.com/zacscoding/gochecker
```

# Usage

> Initialize new health checker

```go
package main

import (
	"github.com/zacscoding/gochecker"
	"time"
)

func main() {
	// This health checker uses cache health result with given TTL
	checker := gochecker.NewHealthChecker(gochecker.WithCacheTTL(time.Minute))

	// This health checker checks health status with given time interval in a new goroutine
	checker = gochecker.NewHealthChecker(gochecker.WithBackground(time.Minute))
}
```  

> Add checkers and observers

```go
package main

import (
	"github.com/zacscoding/gochecker"
	"github.com/zacscoding/gochecker/database"
	"time"
)

func main() {
	// Add health check components
	// add database component
	checker.AddChecker("MyDatabase", database.NewMySQLIndicator(db))
	// add remote service component
	checker.AddChecker("RemoteService-1", gochecker.NewUrlIndicator("RemoteService-1", "GET", "http://anotherservice.com", nil, time.Second*10))

	// Add observers
	checker.AddObserver("BatchProcessor", batchProcessor)
	checker.AddChecker("RemoteService-2", gochecker.NewUrlIndicator("RemoteService-2", "GET", "http://localhost:8890", nil, time.Second*10))
}
```  

> Getting health check result  

```go
status := checker.Health(context.Background())
```

```json
{
  "status": "UP",
  "components": {
    "BatchProcessor": {
      "details": {
        "checkpoint": "100"
      },
      "status": "UP"
    },
    "LocalDatabase": {
      "details": {
        "database": "mysql",
        "validationQuery": "SELECT 1",
        "version": "5.6.1"
      },
      "status": "UP"
    },
    "RemoteService-1": {
      "details": {
        "status": 200
      },
      "status": "UP"
    },
    "RemoteService-2": {
      "details": {
        "err": "Get \"http://localhost:8990\": dial tcp [::1]:8990: connect: connection refused"
      },
      "status": "DOWN"
    }
  }
}
```