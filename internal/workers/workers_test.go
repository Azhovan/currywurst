package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/currywurst/internal/cashregister"
	"github.com/currywurst/internal/orders"
	"github.com/currywurst/internal/terminals"
)

func Test_Run(t *testing.T) {
	tm, err := terminals.NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	cr := cashregister.NewCashRegister()

	order := orders.NewOrder(context.TODO(), 40, orders.Vegan)
	err = tm.Put(order)
	if err != nil {
		t.Fatalf("failed to send new order to terminal, err:%v", err)
	}

	workers := NewWorker(tm, cr)
	go workers.Run()

	err = order.WaitWithTimeout(time.Second * 20)
	if err != nil {
		t.Errorf("expected nil error, got:%v", err)
	}

	if order.Error != nil {
		t.Errorf("expected nil error, got:%v", order.Error)
	}
}

func Test_CancelledOrderByCustomer(t *testing.T) {
	tm, err := terminals.NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	cr := cashregister.NewCashRegister()

	order := orders.NewOrder(context.TODO(), 40, orders.Vegan)
	err = tm.Put(order)
	if err != nil {
		t.Fatalf("failed to send new order to terminal, err:%v", err)
	}
	// cancelling order
	err = order.Cancel()
	if err != nil {
		t.Fatalf("expected nil order, got:%v", err)
	}

	workers := NewWorker(tm, cr)
	go workers.Run()

	err = order.WaitWithTimeout(time.Second * 20)
	if !errors.Is(err, orders.ErrOrderCancelled) {
		t.Errorf("expected error type %v, got %v", orders.ErrOrderCancelled, err)
	}
}

func Test_InvalidOrder(t *testing.T) {
	tm, err := terminals.NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	cr := cashregister.NewCashRegister()

	// an invalid order (invalid price, inserted price is less than expected price)
	order := orders.NewOrder(context.TODO(), 20, orders.Vegan)
	err = tm.Put(order)
	if err != nil {
		t.Fatalf("failed to send new order to terminal, err:%v", err)
	}

	workers := NewWorker(tm, cr)
	go workers.Run()

	err = order.WaitWithTimeout(time.Second * 20)
	if err != nil {
		t.Fatalf("expected nil order, got:%v", err)
	}

	var e *orders.ErrInvalidOrder
	if !errors.As(order.Error, &e) {
		t.Errorf("expected error type %v, got %v", e, order.Error)
	}
}
