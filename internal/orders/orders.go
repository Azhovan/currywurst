package orders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/azhovan/currywurst/internal/cashregister"
	"github.com/azhovan/currywurst/pkg"
)

var (
	// ErrInvalidOrderType is the error returned when the currywurst type is invalid or unknown.
	ErrInvalidOrderType = errors.New("invalid currywurst type")

	// ErrOrderWithInvalidCtx is the error returned when the order does not have a cancellable context.
	ErrOrderWithInvalidCtx = errors.New("order requires a cancellable context to handle timeouts and cancellations")

	// ErrOrderTimeout is the error returned when the order could not be completed within the given timeout.
	ErrOrderTimeout = errors.New("order could not be completed. please try again later")

	// ErrOrderCancelled is the error returned when the order is cancelled by the customer or the terminal
	ErrOrderCancelled = errors.New("order was cancelled")

	// ErrOrderNil is the error returned when the order is nil.
	ErrOrderNil = errors.New("order is nil")
)

// Order represents the struct to hold the input parameters for the order.
// Order represents a pizza order
type Order struct {
	OrderStatus

	// ctx is the context associated with the order, which can be used to cancel the order
	// or pass request-scoped values across API boundaries and between processes.
	//
	// The context should be created and canceled by the customer or the terminal when they
	// want to cancel the order. The terminal and the workers should check the context.Done()
	// channel before getting or processing the order, and return nil if they receive a signal.
	//
	// This context MUST NOT be stored in a global or package-level variable, used across
	// API boundaries or between processes, reused or modified by multiple goroutines,
	// persisted or serialized, exposed to the user or the client, nested or embedded in
	// another struct, passed to a function that does not need it, returned from a function
	// that creates it, or used for a long time or for multiple operations.
	//
	// Refer to this blog post for more details : https://go.dev/blog/context-and-structs
	ctx    context.Context
	cancel context.CancelFunc

	OrderType OrderType // The type of the currywurst, i.e. vegan, non-vegan
	Inserted  int       // The amount of money inserted by the customer in cents
}

// OrderStatus holds all the information related to the order status.
// The customer can listen to the Ready channel to get notified
// when the order is ready or when there is an error.
// The customer can also check the Error field to see the details of the error, if any.
// The Ready channel is buffered with a capacity of 1, so the
// cash register can send a value to it without blocking.
type OrderStatus struct {
	Ready    chan bool                   // A channel to signal the customer when the order is ready (true) or when there is an error (false).
	Error    error                       // Any error related to the order or payment.
	Returned cashregister.ReturnedAmount // The amount of money returned to the customer
}

// NewOrder returns a new instance of the Order with the given values.
func NewOrder(ctx context.Context, inserted int, orderType OrderType) *Order {
	ctx, cancel := context.WithCancel(ctx)
	return &Order{
		ctx:       ctx,
		cancel:    cancel,
		Inserted:  inserted,
		OrderType: orderType,
		OrderStatus: OrderStatus{
			Ready:    make(chan bool, 1),
			Error:    nil,
			Returned: cashregister.ReturnedAmount{},
		},
	}
}

// The Cancel method prevents the order from being processed by workers.
// it returns an error when context is not cancellable or missing.
func (o *Order) Cancel() error {
	if o == nil {
		return ErrOrderNil
	}
	if o.ctx == nil || o.cancel == nil {
		return ErrOrderWithInvalidCtx
	}

	o.cancel()
	return nil
}

// IsCancelled checks if the order has been canceled by a call to the cancel function.
// It returns true if the order context is canceled, false otherwise. It also returns an error if the order is nil.
func (o *Order) IsCancelled() (bool, error) {
	if o == nil {
		return false, ErrOrderNil
	}

	select {
	// There is only one way to cancel the order! and that is by the customer
	// so that why we don't check the ctx.Err() like:
	//  if !errors.Is(o.ctx.Err(), context.Canceled) {
	//			return false, nil
	//	}
	case <-o.ctx.Done():
		return true, nil
	default:
		return false, nil
	}
}

// WaitWithTimeout waits for the order to be ready or timeout within the given duration.
// It returns nil if the order is ready, ErrOrderCancelled if the order is cancelled by the customer or the terminal,
// or ErrOrderTimeout if the order could not be completed within the timeout.
func (o *Order) WaitWithTimeout(timeout time.Duration) error {
	if o == nil {
		return ErrOrderNil
	}

	for {
		select {
		case <-time.NewTimer(timeout).C:
			return ErrOrderTimeout
		case <-o.ctx.Done():
			return ErrOrderCancelled
		case <-o.Ready:
			return nil
		}
	}
}

// Validate validates the different aspects of the order
func (o *Order) Validate() error {
	if o == nil {
		//  A nil order pointer is not an unrecoverable error!
		// that is why I prefer to not panic!
		return ErrOrderNil
	}

	orderType := pkg.GetOrderType(o.OrderType.String())
	if orderType == nil {
		return ErrInvalidOrderType
	}

	if o.Inserted < orderType.Price() {
		return &ErrInvalidOrder{inserted: o.Inserted, price: orderType.Price()}
	}

	return nil
}

// ErrInvalidOrder is a custom error type that indicates that the order is invalid.
type ErrInvalidOrder struct {
	inserted, price int
}

// Error returns the error message for ErrInvalidOrder.
func (e *ErrInvalidOrder) Error() string {
	return fmt.Sprintf("paid:%d, expected:%d", e.inserted, e.price)
}
