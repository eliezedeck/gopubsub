package gopubsub

import (
	"sync"
	"testing"
)

func TestBasicOps(t *testing.T) {
	ps := New(4)
	ps.Close()

	ps.Publish("something")

	var allGotSomething sync.WaitGroup
	var allGotAnother sync.WaitGroup

	test := func() {
		c := ps.Subscribe(true)

		<-c // getting "something" here
		allGotSomething.Done()

		<-c // getting "another" here
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
}
