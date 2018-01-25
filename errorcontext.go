package main

import (
	"errors"
	"sync"
)

var StoppedErr error = errors.New("Visitor stopped reading.")

type ErrorContext struct {
	wg       sync.WaitGroup
	Canceled chan struct{}

	Err    error
	errMut sync.Mutex

	latest int
}

func NewErrorContext() *ErrorContext {
	ec := &ErrorContext{Canceled: make(chan struct{})}
	return ec
}

// Add(n) adds n workers to the loadContext.
func (ec *ErrorContext) Add(n int) {
	ec.wg.Add(n)
	ec.latest = ec.latest + n
}

// Start() adds a worker and returns an ID for it.
func (ec *ErrorContext) Start() int {
	ec.wg.Add(1)

	ec.latest++
	return ec.latest
}

// Done() marks a worker as done.
func (ec *ErrorContext) Done() {
	ec.wg.Done()
}

// Cancel(err) cancels the ErrorContext with an error, closing the Canceled channel.
// Subsequent calls to Cancel() have no effect.
func (ec *ErrorContext) Cancel(err error) {
	ec.errMut.Lock()
	select {
	case <-ec.Canceled:
		// Do nothing, already canceled.
	default:
		// Not yet canceled.
		ec.Err = err
		close(ec.Canceled) // cancel the context.
	}
	ec.errMut.Unlock()
}

// Wait() blocks until either all worker goroutines call Done(), or the ErrorContext is canceled. It returns nil on success, and the error if one occurs.
func (ec *ErrorContext) Wait() error {
	success := make(chan struct{})

	go func() {
		ec.wg.Wait()
		close(success)
	}()

	select {
	case <-success:
		return nil
	case <-ec.Canceled:
		return ec.Err
	}
}
