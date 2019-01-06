package pool

import (
	"fmt"
	"time"
)

type Task func() error

type Pool struct {
	Size  int
	tasks chan Task
	syncc chan struct{}
}

func (p *Pool) run() {
	for {
		select {
		case t, ok := <-p.tasks:
			if !ok {
				return
			}
			go func() {
				err := t()
				if err != nil {
					fmt.Println("worker: " + err.Error())
				}
			}()
		case <-time.After(1 * time.Second):
		}
	}
}

func (p *Pool) Close() {
	close(p.tasks)
	p.syncc <- struct{}{}
}

func (p *Pool) Schedule(t Task) {
	p.tasks <- t
}

func New(size int, s chan struct{}) *Pool {
	p := &Pool{
		Size:  size,
		tasks: make(chan Task, size),
		syncc: s,
	}
	go p.run()
	return p
}
