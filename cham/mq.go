package cham

import (
	"fmt"
	"sync"
)

const (
	DEFAULT_QUEUE_SIZE = 64
)

type Queue struct {
	sync.Mutex
	head int
	tail int
	buf  []*Msg
}

func NewQueue() *Queue {
	q := new(Queue)
	q.buf = make([]*Msg, DEFAULT_QUEUE_SIZE)
	return q
}

func (q *Queue) String() string {
	return fmt.Sprintf("QUEUE[%d-%d]", q.head, q.tail)
}

func (q *Queue) Length() int {
	q.Lock()
	head := q.head
	tail := q.tail
	size := cap(q.buf)
	q.Unlock()

	if head <= tail {
		return tail - head
	}
	return tail + size - head
}

func (q *Queue) Push(msg *Msg) {
	q.Lock()
	q.buf[q.tail] = msg
	q.tail++

	if q.tail >= cap(q.buf) {
		q.tail = 0
	}

	if q.head == q.tail {
		q.expand()
	}
	q.Unlock()
}

func (q *Queue) Pop() (msg *Msg) {
	q.Lock()
	if q.head != q.tail {
		msg = q.buf[q.head]
		q.head++
		if q.head >= cap(q.buf) {
			q.head = 0
		}
	}
	q.Unlock()
	return
}

func (q *Queue) expand() {
	newbuf := make([]*Msg, cap(q.buf)*2)
	copy(newbuf, q.buf[q.head:cap(q.buf)])
	q.head = 0
	q.tail = cap(q.buf)
	q.buf = newbuf
}
