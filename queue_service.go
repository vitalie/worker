package worker

// response represents a queue response.
type response struct {
	Msg Message
	Err error
}

// queueService represents a Queue wrapper which allows
// polling the queue without blocking.
type queueService struct {
	queue Queue
}

// newQueueService returns a queueService instance.
func newQueueService(q Queue) *queueService {
	return &queueService{queue: q}
}

// get service peeks a message from queue and
// returns it through a channel without blocking
// the caller.
func (qs *queueService) get() <-chan *response {
	// Buffered channel to avoid goroutine leaking.
	out := make(chan *response, 1)

	// Perform the request in a separate
	// goroutine to avoid blocking.
	go func() {
		defer close(out)

		msg, err := qs.queue.Get()
		if err != nil {
			out <- &response{Err: err}
			return
		}

		out <- &response{Msg: msg}
	}()

	return out
}
