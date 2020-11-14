package types

import (
	"sync"
)

type ErrorSwitchChannel interface {
	Send(err error)
	Receive() error
}

type errorSwitchChannel struct {
	done  bool
	err   error
	mux   sync.Mutex
	errCh chan error
}

func (ch *errorSwitchChannel) Send(err error) {
	ch.mux.Lock()
	if !ch.done {
		ch.errCh <- err
		close(ch.errCh)
		ch.done = true
	}
	ch.mux.Unlock()
}

func (ch *errorSwitchChannel) Receive() error {
	return <-ch.errCh
}

func NewErrorSwitchChannel() ErrorSwitchChannel {
	return &errorSwitchChannel{
		done:  false,
		err:   nil,
		mux:   sync.Mutex{},
		errCh: make(chan error, 1),
	}
}
