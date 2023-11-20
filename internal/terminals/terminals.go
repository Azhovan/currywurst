package terminals

import (
	"errors"
	"fmt"
	"sync"

	"github.com/azhovan/currywurst/internal/orders"
)

var (
	// ErrTerminalNil is the error returned when the terminal is nil.
	ErrTerminalNil = errors.New("terminal is nil")

	// ErrTerminalFull is the error returned when the terminal is full and cannot accept new orders.
	ErrTerminalFull = errors.New("terminal is full")

	// ErrTerminalEmpty is the error returned when the terminal is empty and has no orders to process.
	ErrTerminalEmpty = errors.New("terminal is empty")

	// ErrTerminalClosed is the error returned when the terminal is closed and cannot accept or process any orders.
	ErrTerminalClosed = errors.New("terminal is closed")
)

// A Terminal is a queue of orders that customers can join and place their orders.
// A terminal has a fixed capacity and can process one order at a time.
type Terminal struct {
	orders   chan *orders.Order // The list of customer's order
	done     chan bool          // The signal to indicate that the terminal is closed and no more orders can be added or processed
	capacity int                // The maximum number of orders that can be queued at a time before terminal is blocked.

	mu   sync.Mutex // A lock that terminal holds when signaling/waiting
	cond *sync.Cond // A conditional variable for signaling the worker when terminal is empty or full
}

// NewTerminal creates and returns a new Terminal with an empty queue of orders.
func NewTerminal(capacity int) (*Terminal, error) {
	if capacity < 0 {
		return nil, fmt.Errorf("invalid capacity: %d", capacity)
	}

	terminal := &Terminal{
		orders:   make(chan *orders.Order, capacity),
		done:     make(chan bool, 1),
		capacity: capacity,
	}
	terminal.cond = sync.NewCond(&terminal.mu)
	return terminal, nil
}

// Put adds an order to the terminal's queue of orders.
// It returns nil if successful, or an error if the
// terminal is nil, closed, full, or the order is nil.
func (t *Terminal) Put(order *orders.Order) error {
	if t == nil {
		return ErrTerminalNil
	}
	if order == nil {
		return orders.ErrOrderNil
	}

	// using conditional variable has a downside here
	t.mu.Lock()
	defer t.mu.Unlock()
	// terminal is full, wait for the worker to catch up
	for len(t.orders) == t.capacity {
		t.cond.Wait()
	}

	// defer the signal for the worker
	// if it has been waiting for orders to come in
	defer t.cond.Signal()

	// try to send the order to the queue
	select {
	case <-t.done:
		return ErrTerminalClosed
	case t.orders <- order:
		return nil
	default:
		return ErrTerminalFull
	}
}

// Get returns and removes the first order in the terminal's queue of orders.
// It returns the order and nil if successful, or nil and an error if the terminal is nil, closed, or the order is nil or cancelled.
func (t *Terminal) Get() (*orders.Order, error) {
	if t == nil {
		return nil, ErrTerminalNil
	}

	// terminal has been already closed
	select {
	case <-t.done:
		return nil, ErrTerminalClosed
	default:
	}

	// check if there are any orders in the queue
	// block until a new orders come in
	t.mu.Lock()
	defer t.mu.Unlock()
	// defer the wake-up signal, so new orders can come in
	defer t.cond.Signal()

	// block until the terminal has a new order to process
	// or the terminal is no longer open
	for len(t.orders) == 0 {
		closed, _ := t.IsClosed()
		if closed {
			return nil, ErrTerminalClosed
		}

		t.cond.Wait()
	}

	// there is only one goroutine that consumes this channel,
	// because the channel will not be closed and drained by
	// another goroutine we may not need to check the `ok` value
	// after receiving from the channel like this:
	// 	order, ok := <-t.orders
	//	 if !ok {
	// 		return nil, ErrTerminalClosed
	// 	}
	order := <-t.orders

	// this MUST not happen, but just in case there is curred data in the system
	// we add this safety check
	if order == nil {
		return nil, orders.ErrOrderNil
	}

	cancelled, err := order.IsCancelled()
	switch {
	// order has been cancelled, but it doesn't mean that it is nil!
	// so that we return the order too
	case cancelled && err == nil:
		return order, orders.ErrOrderCancelled
	case err != nil:
		return nil, err
	default:
		return order, nil
	}
}

// Close closes the terminal and stops accepting or processing new orders.
// It returns nil if successful, or an error if the terminal is nil or already closed.
func (t *Terminal) Close() error {
	// check if the terminal is valid and open
	if t == nil {
		return ErrTerminalNil
	}
	select {
	case <-t.done:
		return ErrTerminalClosed
	default:
	}

	// close the terminal
	close(t.done)
	// setting the orders to nil is safe, because the only goroutine
	// that sends values to this channel is the current terminal, which is closed
	// therefore, there is no risk of panic.
	t.orders = nil
	return nil
}

// IsClosed checks if the terminal is closed.
// It returns true if the terminal is closed, false otherwise.
// It also returns an error if the terminal is nil.
func (t *Terminal) IsClosed() (bool, error) {
	if t == nil {
		return false, ErrTerminalNil
	}

	select {
	case <-t.done:
		return true, nil
	default:
		return false, nil
	}
}
