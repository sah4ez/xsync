package pool

import "time"

type Task func() error

type Pool struct {
	Size  int
	tasks chan Task
}

func (p *Pool) run() {
	for {
		select {
		case t, ok := <-p.tasks:
			if !ok {
				return
			}
			go t()
		case <-time.After(1 * time.Second):
		}
	}
}

func (p *Pool) Close() {
	close(p.tasks)
}

func (p *Pool) Schedule(t Task) {
	p.tasks <- t
}

func New(size int) *Pool {
	p := &Pool{
		Size:  size,
		tasks: make(chan Task, size),
	}
	go p.run()
	return p
}
