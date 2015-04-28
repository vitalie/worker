package worker

import (
	"encoding/json"
	"strconv"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/sqs"
)

const (
	sqsSizeKey = "ApproximateNumberOfMessages"
)

type SQSQueue struct {
	queue  *sqs.Queue
	region aws.Region
}

func NewSQSQueue(name string, opts ...func(*SQSQueue)) (Queue, error) {
	q := &SQSQueue{
		region: aws.USEast,
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}

	// Get a reference to SQS region.
	service := sqs.New(auth, q.region)
	queue, err := service.GetQueue(name)
	if err != nil {
		return nil, err
	}

	q.queue = queue

	return q, nil
}

func (q *SQSQueue) Put(j Job) error {
	typ, err := StructType(j)
	if err != nil {
		return err
	}

	job := &Payload{
		Type: typ,
		Args: j,
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = q.queue.SendMessage(string(payload))
	return err
}

func (q *SQSQueue) Get() (Message, error) {
	resp, err := q.queue.ReceiveMessage(1)
	if err != nil {
		return nil, err
	}

	if len(resp.Messages) == 0 {
		return nil, &Error{Err: "timeout", IsTimeout: true}
	}

	msg := resp.Messages[0]
	env, err := newEnvelope(msg.MessageId, []byte(msg.Body))
	if err != nil {
		return nil, err
	}
	env.extra = msg

	return env, nil
}

func (q *SQSQueue) Delete(msg Message) error {
	env, ok := msg.(*envelope)
	if !ok {
		return NewErrorFmt("bad envelope")
	}

	if sm, ok := env.extra.(sqs.Message); ok {
		_, err := q.queue.DeleteMessage(&sm)
		return err
	}

	return NewErrorFmt("bad sqs message")
}

func (q *SQSQueue) Reject(msg Message) error {
	return nil
}

func (q *SQSQueue) Size() (uint64, error) {
	resp, err := q.queue.GetQueueAttributes(sqsSizeKey)
	if err != nil {
		return 0, err
	}

	if len(resp.Attributes) == 0 {
		return 0, NewError("bad attribute size")
	}

	attr := resp.Attributes[0]
	if attr.Name != sqsSizeKey {
		return 0, NewError("bad attribute")
	}

	return strconv.ParseUint(attr.Value, 10, 64)
}
