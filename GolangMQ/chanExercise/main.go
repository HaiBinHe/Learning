package main

import (
	"sync"
	"time"
)

func main() {
	var c = make(chan string, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c <- "Golang Channel"
		c <- "hhb"
	}()

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		println("text:", <-c)
		println("author:", <-c)
	}()
	wg.Wait()
}
