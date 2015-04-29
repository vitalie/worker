package worker

type response struct {
	Msg Message
	Err error
}

type queueService struct {
	queue Queue
}

func newQueueService(q Queue) *queueService {
	return &queueService{queue: q}
}

// get service peeks a message from queue and
// returns it through a channel without blocking
// the caller.
func (qs *queueService) get() <-chan *response {
	// Buffered channel to avoid goroutine leaking.
	out := make(chan *response, 1)

	// Start a gorouting to avoid blocking.
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
