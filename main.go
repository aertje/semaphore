package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/aertje/gonice/priority"
)

func main() {
	p := priority.New()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			fnDone := <-p.Wait(10 - i)

			time.Sleep(1 * time.Second)

			fnDone()

			fmt.Printf("done %d", i)

			defer wg.Done()
		}()
	}
	wg.Wait()
}
