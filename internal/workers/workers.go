package workers

import (
	"github.com/azhovan/currywurst/internal/cashregister"
	"github.com/azhovan/currywurst/internal/orders"
	"github.com/azhovan/currywurst/internal/terminals"
	"github.com/azhovan/currywurst/pkg"
)

// Worker represents a worker that can process orders from a terminal and return change using a cash register.
// A worker has a reference to a terminal and a cash register, and can run a loop that reads orders
// from the terminal, validates them, pays them using the cash register, and returns the change to the customer.
type Worker struct {
	terminal *terminals.Terminal
	cr       *cashregister.CashRegister
}

// NewWorker returns a new worker instance with the given terminal and cash register.
// It does not start the worker loop; use the Run method for that.
func NewWorker(t *terminals.Terminal, cr *cashregister.CashRegister) *Worker {
	return &Worker{
		terminal: t,
		cr:       cr,
	}
}

// Run starts the worker loop that processes orders from the terminal and returns change using the cash register.
// It expects the terminal and the cash register to be non-nil and initialized.
// It sets the Error field of the order if any error occurs during the payment process.
func (w *Worker) Run() {
	// unrecoverable state
	// we may log here, for simplicity lets just exit
	if w.terminal == nil {
		panic("invalid worker initialization. terminal is nil")
	}

	// unrecoverable state
	// we may log here, for simplicity lets just exit
	if w.cr == nil {
		panic("invalid worker initialization. cash register is nil")
	}

	for {
		// Get() blocks until there is a new order
		// it also has internal check for order cancellation and terminal closing
		order, err := w.terminal.Get()

		switch err {
		// there is nothing to do here. we may log it as well
		// for simplicity we just exit
		case terminals.ErrTerminalNil:
		case terminals.ErrTerminalClosed:
			return
		// we may log it as well
		// for simplicity lets just move to next
		case orders.ErrOrderNil:
		case orders.ErrOrderCancelled:
		case terminals.ErrTerminalEmpty:
			continue
		default:
		}

		// The order validation can also happen in the HTTP handler or in the Put() method of terminal.
		// but having it here is more concise, because we don't have to
		// expose the internal implementation details in the HTTP handler nor
		// complicate sending order to terminal. besides this is very inexpensive operation.
		if err = order.Validate(); err != nil {
			order.Error = err
			order.Ready <- false
			continue
		}

		// we do check if at this stage either order has been
		// cancelled or the terminal is still open
		orderCancelled, orderErr := order.IsCancelled()
		if orderCancelled && orderErr == nil {
			continue
		}

		// if terminal has been closed, we won't proceed with the order and exit
		terminalClosed, terminalErr := w.terminal.IsClosed()
		if terminalClosed && terminalErr == nil {
			return
		}

		// There is no need to check whether the orderType is nil or not
		// this has happened already in the validation step, so we are confident that
		// order at this stage is valid and has proper price
		price := pkg.GetOrderType(order.OrderType.String()).
			Price()

		// calculate the returned price
		returned, err := w.cr.Pay(price, order.Inserted)
		if err != nil {
			order.Error = err
			order.Ready <- false

			continue
		}

		order.Ready <- false
		order.Returned = returned
	}
}
