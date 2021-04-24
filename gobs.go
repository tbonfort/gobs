// Package gobs implements a simple job queue where each individual job is run
// concurrently in its own goroutine while ensuring that no more than a given number
// of jobs can be ran at a time. It provides methods to ensure all jobs have been
// completed, and to capture errors
package gobs

import (
	"fmt"
	"sync"
)

// Job is a unit of work that returns a non-nil error in case of failure
type Job func() error

// Status tracks the completion of a Job
type Status struct {
	done chan struct{}
	err  error
}

// Wait blocks until the associated Job has terminated. It returns the error
// returned by the Job
func (s *Status) Wait() error {
	<-s.done
	return s.err
}

// Pool is a worker pool that accepts a bounded number of Jobs
type Pool struct {
	concurrency int
	jobs        chan struct{}
}

// Stop blocks until all submitted jobs have completed, then frees all resources created
// by the pool.
//
// Submitting a new Job to the Pool once Stop has been called will deadlock and/or panic.
// Calling Stop more than once will deadlock and/or panic
func (p *Pool) Stop() {
	//aquire all job slots (i.e. make sure none are occupied by a running job)
	for i := 0; i < p.concurrency; i++ {
		p.jobs <- struct{}{}
	}
	close(p.jobs)
	for i := 0; i < p.concurrency; i++ {
		<-p.jobs
	}
}

// NewPool creates a worker pool garanteeing that no more than concurrency jobs will
// be running at a given instant
func NewPool(concurrency int) *Pool {
	if concurrency < 1 {
		panic("concurrency must be >= 1")
	}
	p := &Pool{
		concurrency: concurrency,
	}
	p.jobs = make(chan struct{}, concurrency)
	return p
}

// Submit adds a new job to the worker pool. Submit blocks until the pool's concurrency
// setting allows the job to start running, then launches the job in a new goroutine.
// It returns a Status that can be used to track the job completion and/or error
func (p *Pool) Submit(job Job) *Status {
	p.jobs <- struct{}{}
	s := &Status{}
	s.done = make(chan struct{})
	go func() {
		s.err = job()
		close(s.done)
		<-p.jobs
	}()
	return s
}

type multiErr struct {
	mu   sync.Mutex
	errs []error
}

var _ MultiError = &multiErr{} // static type check

func (me *multiErr) add(err error) {
	if err == nil {
		return
	}
	me.mu.Lock()
	me.errs = append(me.errs, err)
	me.mu.Unlock()
}

// Error is the standard error interface
func (me *multiErr) Error() string {
	errstring := me.errs[0].Error()
	if len(me.errs) > 1 {
		errstring += fmt.Sprintf(" (and %d more errors)", len(me.errs)-1)
	}
	return errstring
}

// Errors returns the list of all errors that occurred in a batch
func (me *multiErr) Errors() []error {
	return me.errs
}

// MultiError is an error that handles multiple unrelated errors
type MultiError interface {
	error
	Errors() []error
}

// Batch is a holder for a group of jobs to be run in the Pool. Batch
// exposes a Wait() function which allows code to block while waiting for all jobs
// of the batch to be completed
type Batch struct {
	p  *Pool
	wg sync.WaitGroup
	me *multiErr
}

// Batch creates a new holder for a group of jobs to be run in the Pool.
func (p *Pool) Batch() *Batch {
	b := &Batch{
		p:  p,
		me: &multiErr{},
	}
	return b
}

// Submit adds a new job to the batch in the worker pool. Submit blocks until the pool's concurrency
// setting allows the job to start running, then launches the job in a new goroutine.
// To track the job's completion, use Batch.Wait()
func (b *Batch) Submit(job Job) {
	st := b.p.Submit(job)
	b.wg.Add(1)
	go func() {
		b.me.add(st.Wait())
		b.wg.Done()
	}()
}

// Wait blocks until all job submitted to the batch have completed. Once Wait
// has been called, no further jobs should be submitted to the batch.
//
// Wait returns a MultiError that can be used to inspect individual job errors
func (b *Batch) Wait() error {
	b.wg.Wait()
	if len(b.me.errs) > 0 {
		return b.me
	}
	return nil
}
