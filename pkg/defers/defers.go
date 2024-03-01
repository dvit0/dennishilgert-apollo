package defers

import "sync"

// Defers maintains an ordered lifo list of functions to handle on the defer call.
type Defers interface {
	// Add adds a new function to the beginning of the list.
	Add(func())

	// CallAll invokes all deffered functions in reverse order.
	CallAll()

	// Trigger tells the instance to (true) or not to (false) process the defers.
	Trigger(bool)
}

type defaultDefers struct {
	// A mutex is used to lock the resources while they are used by the defers.
	sync.Mutex

	fs      []func()
	trigger bool
}

// NewDefers returns a new instance of Defers.
func NewDefers() Defers {
	return &defaultDefers{
		fs:      []func(){},
		trigger: true,
	}
}

func (df *defaultDefers) Add(fn func()) {
	df.fs = append([]func(){fn}, df.fs...)
}

func (df *defaultDefers) CallAll() {
	df.Lock()
	defer df.Unlock()
	if !df.trigger {
		return
	}
	for _, fn := range df.fs {
		fn()
	}
}

func (df *defaultDefers) Trigger(input bool) {
	df.Lock()
	defer df.Unlock()
	df.trigger = input
}
