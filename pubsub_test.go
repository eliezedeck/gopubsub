package gopubsub

import (
	"sync"
	"testing"
	"time"
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

func TestIncreasedSubscribers(t *testing.T) {
	ps := New(2)
	defer ps.Close()

	count := 5
	for i := 0; i < count; i++ {
		ps.Subscribe(false)
	}

	if len(ps.subscribers) != count {
		t.Errorf("Expected to have %d subscribers, have %d", count, len(ps.subscribers))
	}
}

func TestUnsubscribe(t *testing.T) {
	ps := New(2)
	defer ps.Close()

	c := ps.Subscribe(false)
	ps.Unsubscribe(c)
	time.Sleep(10 * time.Millisecond) // wait for the op to complete

	if len(ps.subscribers) > 0 {
		t.Errorf("Expected to have 0 subscribers, have %d", len(ps.subscribers))
	}
}
