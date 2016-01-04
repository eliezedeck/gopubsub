package gopubsub

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestBasicOps(t *testing.T) {
	ps := New(4)

	ps.Publish("something")

	var allGotSomething sync.WaitGroup
	var allGotAnother sync.WaitGroup

	test := func() {
		c := ps.Subscribe(true)
	L:
		for {
			select {
			case d := <-c:
				log.Println("Got data", d)
				break L
			case <-time.After(200 * time.Millisecond):
				panic("timeout, didn't receive data")
			}
		}

		allGotSomething.Done()
		log.Println("Got data", <-c)
		allGotAnother.Done()
	}

	for i := 0; i < 5; i++ {
		allGotSomething.Add(1)
		allGotAnother.Add(1)
		go test()
	}

	allGotSomething.Wait()
	ps.Publish("another")
	allGotAnother.Wait()

	ps.Close()
	time.Sleep(100 * time.Millisecond)
}
