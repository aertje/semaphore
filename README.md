# Semaphore

[![Go Reference](https://pkg.go.dev/badge/github.com/aertje/semaphore.svg)](https://pkg.go.dev/github.com/aertje/semaphore)
[![Go Report Card](https://goreportcard.com/badge/github.com/aertje/semaphore)](https://goreportcard.com/report/github.com/aertje/semaphore)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/dynamic/json?url=https://aertje.github.io/semaphore/coverage.json&label=coverage&query=$.coverage&color=brightgreen)](https://aertje.github.io/semaphore/coverage.html)

The `semaphore` package provides a priority-based concurrency control mechanism. It allows you to manage the execution of functions based on their priority while respecting a maximum concurrency limit. This is particularly useful in scenarios where certain tasks need to be prioritised over others, and there is a need to limit the number of concurrent tasks to avoid overloading the system.

The general use case is to prioritise certain CPU-bound tasks over others. For example, in a web service, it could be used for example to prioritise the alive endpoint over the metrics endpoint, or to serve bulk requests before real-time requests.

The implementation does not interfere with Go's runtime semaphore. It is opt-in and does not affect the behavior of other goroutines in the application.

## Features

- **Priority-based scheduling**: Tasks are executed based on their priority. High priority tasks are started before low priority tasks once the maximum concurrency limit is reached.
- **Configurable maximum concurrency limit**: The number of concurrent tasks is configurable, defaulting to [`GOMAXPROCS`](https://pkg.go.dev/runtime#GOMAXPROCS).
- **Context cancellation**: Waiting tasks can optionally be cancelled using a context.
- **Force acquire**: Tasks can bypass the maximum concurrency limit using force acquire. These tasks will execute immediately but still count towards the concurrency limit for regular tasks. This ensures critical tasks are never blocked while maintaining backpressure on non-critical tasks.

## Installation

To install the package, use the following command:

```sh
go get github.com/aertje/semaphore
```

## Simple example

The following minimal example demonstrates how to use the `semaphore` package to create a semaphore that starts tasks based on their priority. It illustrates the required steps to create a semaphore, register a task with a specific priority, and signal the completion of the task.

```go
package main

import (
    "fmt"
    "time"

    "github.com/aertje/semaphore/semaphore"
)

func main() {
    // Create a new prioritized semaphore with the default maximum concurrency limit.
    s := semaphore.NewPrioritized()

    // Register a task with the semaphore with a priority of 1.
    s.Acquire(1)
    // Ensure signalling the completion of the task.
    defer s.Release()

    // Simulate a long-running task.
    time.Sleep(1 * time.Second)
}
```

The steps are as follows:

- Create a new semaphore with an optional maximum concurrency limit.

Then, for each task to be prioritised:

- Register a task with the semaphore using the `Acquire` method. This will block until the task can be executed.
- Execute the task.
- Call the `Release` method to signal the completion of the task to the semaphore.

Note the importance of calling the `Release` method to signal the completion of the task. This is necessary to allow other tasks to be executed by the semaphore.

If the context needs to be taken into account in order to support cancellation, the `AcquireContext` method can be used instead. If a highly critical task needs to be executed, the `ForceAcquire` method can be used to bypass the maximum concurrency limit.

## Example use case: Prioritizing critical endpoints

This example demonstrates the key features of the semaphore:

- Priority-based task execution
- Force acquire for critical tasks
- Context cancellation support

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/aertje/semaphore/semaphore"
)

func main() {
    // Create a semaphore with max 2 concurrent tasks
    s := semaphore.NewPrioritized(semaphore.WithMaxConcurrency(2))

    // Critical health check - uses force acquire to bypass limits
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        s.ForceAcquire()
        defer s.Release()
        w.Write([]byte("OK"))
    })

    // High priority endpoint - uses priority 1
    http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        if err := s.AcquireContext(r.Context(), 1); err != nil {
            http.Error(w, "Request cancelled", 499)
            return
        }
        defer s.Release()

        time.Sleep(100 * time.Millisecond) // Simulate work
        w.Write([]byte("User data"))
    })

    // Low priority endpoint - uses priority 2
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        if err := s.AcquireContext(r.Context(), 2); err != nil {
            http.Error(w, "Request cancelled", 499)
            return
        }
        defer s.Release()

        time.Sleep(500 * time.Millisecond) // Simulate heavy work
        w.Write([]byte("Metrics data"))
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

This package does not accept contributions at the moment. If you have any suggestions, feedback, or issues, please open an issue on GitHub.
