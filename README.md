# Nice

[![Go Reference](https://pkg.go.dev/badge/github.com/aertje/gonice.svg)](https://pkg.go.dev/github.com/aertje/gonice)
[![Go Report Card](https://goreportcard.com/badge/github.com/aertje/gonice)](https://goreportcard.com/report/github.com/aertje/gonice)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The `nice` package provides a priority-based concurrency control mechanism. It allows you to manage the execution of functions based on their priority while respecting a maximum concurrency limit. This is particularly useful in scenarios where certain tasks need to be prioritised over others, and there is a need to limit the number of concurrent tasks to avoid overloading the system.

The general use case is to prioritise certain CPU-bound tasks over others. For example, in a web service, it could be used for example to prioritise the alive endpoint over the metrics endpoint, or to serve bulk requests before real-time requests.

The implementation does not interfere with Go's runtime scheduler. It is opt-in and does not affect the behavior of other goroutines in the application.

## Features

- **Priority-based scheduling**: Tasks are executed based on their priority. High priority tasks are started before low priority tasks once the maximum concurrency limit is reached.
- **Configurable maximum concurrency limit**: The number of concurrent tasks is configurable, defaulting to [`GOMAXPROCS`](https://pkg.go.dev/runtime#GOMAXPROCS).
- **Context cancellation**: Waiting tasks can optionally be cancelled using a context.

## Installation

To install the package, use the following command:

```sh
go get github.com/aertje/gonice
```

## Simple example

The following minimal example demonstrates how to use the `nice` package to create a scheduler that starts tasks based on their priority. It illustrates the required steps to create a scheduler, register a task with a specific priority, and signal the completion of the task.

```go
package main

import (
    "fmt"
    "time"

    "github.com/aertje/gonice/nice"
)

func main() {
    // Create a new scheduler with the default maximum concurrency limit.
    scheduler := nice.NewScheduler()

    // Register a task with the scheduler with a priority of 1.
    fnDone := scheduler.Wait(1)
    // Signal the completion of the task.
    defer fnDone()

    // Simulate a long-running task.
    time.Sleep(1 * time.Second)
}
```

The steps are as follows:

- Create a new scheduler with an optional maximum concurrency limit.

Then, for each task to be prioritised:

- Register a task with the scheduler using the `Wait` method. This will block until the task can be executed. It returns a function that should be called to signal the completion of the task.
- Execute the task.
- Call the function returned from the call to `Wait` to signal the completion of the task to the scheduler.

Note the importance of calling the function returned by `Wait` to signal the completion of the task. This is necessary to allow other tasks to be executed by the scheduler.

If the context needs to be taken into account in order to support cancellation, the `WaitContext` method can be used instead.

## Example use case: Prioritizing `/alive` endpoint

In this example, we will create a scheduler that prioritises an `/alive` endpoint over other endpoints. This is useful in scenarios where the `/alive` endpoint is critical and needs to be executed before other endpoints.

It also demonstrates use of the `WaitContext` method to support context cancellation. This is useful in scenarios where the client cancels the request, and the server should dispose of the task.

```go
package main

import (
    "context"
    "errors"
    "net/http"
    "time"

    "github.com/aertje/gonice/nice"
)

func main() {
    // Create a new scheduler with a maximum concurrency limit of 10.
    scheduler := nice.NewScheduler(nice.WithMaxConcurrency(10))

    http.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
        // Register a task with the scheduler with a higher priority of 1.
        fnDone, err := scheduler.WaitContext(r.Context(), 1)
        if err != nil {
            if errors.Is(err, context.Canceled) {
                http.Error(w, context.Cause(r.Context()).Error(), 499)
                return
            }

            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer fnDone()

        w.Write([]byte("I'm alive!"))
    })

    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        // Register a task with the scheduler with a lower priority of 2.
        fnDone, err := scheduler.WaitContext(r.Context(), 2)
        if err != nil {
            if errors.Is(err, context.Canceled) {
                http.Error(w, context.Cause(r.Context()).Error(), 499)
                return
            }

            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        defer fnDone()

        time.Sleep(1 * time.Second)

        w.Write([]byte("Metrics are here!"))
    })

    http.ListenAndServe(":8080", nil)
}
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

This package does not accept contributions at the moment. If you have any suggestions, feedback, or issues, please open an issue on GitHub.
