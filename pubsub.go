package gopubsub

import "log"

type pubSub struct {
	subscribers []chan interface{}

	// Operations
	newsub chan subscribeOp
	pub    chan interface{}
	delsub chan (<-chan interface{})
	close  chan struct{}

	lastPublication interface{}

	closed bool
}

type subscribeOp struct {
	getInitial bool
	channel    chan interface{}
}

// New creates a pointer to a pubSub. pubSub uses a []chan interface{} to keep
// the list of subscribers, and the `cap` parameter is the initial headroom, not
// the final number of subscribers possible. It can go as high as you need to.
func New(cap int) *pubSub {
	ps := &pubSub{
		subscribers:     make([]chan interface{}, 0, cap),
		newsub:          make(chan subscribeOp),
		pub:             make(chan interface{}),
		delsub:          make(chan (<-chan interface{})),
		close:           make(chan struct{}),
		lastPublication: nil,
	}

	go func() {
		defer log.Println("Loop closed")
		for {
			select {
			case op := <-ps.newsub:
				// New Subscriber
				//

				ps.subscribers = append(ps.subscribers, op.channel)
				if op.getInitial && ps.lastPublication != nil {
					// subscriber wants the last published data
					op.channel <- ps.lastPublication
				}
			case c := <-ps.delsub:
				// Remove Subscriber
				//

				l := len(ps.subscribers)
				for i := 0; i < l; i++ {
					if ps.subscribers[i] == c {
						// Uses the technique described here: https://github.com/golang/go/wiki/SliceTricks
						// This is used as a pointer set, so it's not like the ordinary case
						ps.subscribers, ps.subscribers[len(ps.subscribers)-1] = append(ps.subscribers[:i], ps.subscribers[i+1:]...), nil
						break
					}
				}
			case p := <-ps.pub:
				// Publish
				//

				ps.lastPublication = p
				for _, suber := range ps.subscribers {
					select {
					case suber <- p:
						// Message consumed by subscriber
					default:
						// Subscriber is not listening
						break
					}
				}
			case <-ps.close:
				// Closed
				//

				for _, suber := range ps.subscribers {
					close(suber)
				}
				break
			}
		}
	}()

	return ps
}

// Subscribe gets you subscribed to this pubSub channel and returns a chan on
// which you can listen for updates. The `getinitial` is used to get the very
// last entry on this channel prior to getting the subsequent updates.
func (ps *pubSub) Subscribe(getinitial bool) <-chan interface{} {
	// buffer set to one so that the pubSub loop will be able to send to this in
	// case of a getinitial == true without blocking
	c := make(chan interface{}, 1)

	r := subscribeOp{
		getInitial: getinitial,
		channel:    c,
	}

	ps.newsub <- r
	return c
}

// Publish is used to publish the given `data` to all subscribers. Not that the
// `data` must not be nil or there will be panic.
func (ps *pubSub) Publish(data interface{}) {
	if data == nil {
		panic("Cannot Publish() nil. That is equivalent to a channel closure which is invalid.")
	}
	ps.pub <- data
}

// Unsubscribe removes the given channel `c` from the list of subscribers. Note
// that `c` must have been returned by the Subscribe() API call.
func (ps *pubSub) Unsubscribe(c <-chan interface{}) {
	ps.delsub <- c
}

// Close is used to dispose of the pubSub. After this call, all the channels
// that have been distributed from the Subscribe() call will be closed and will
// thus return the default nil value of the type.
func (ps *pubSub) Close() {
	if !ps.closed {
		ps.closed = true
		ps.close <- struct{}{}
	}
}
