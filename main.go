package main

import "github.com/aertje/gonice/priority"

func main() {
	p := priority.New(10)
	done := <-p.Wait(1)

	done()
}
