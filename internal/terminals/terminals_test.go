package terminals

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/currywurst/internal/orders"
)

func Test_PutGetOrder(t *testing.T) {
	terminal, err := NewTerminal(5)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	tests := []struct {
		order *orders.Order
	}{
		{order: orders.NewOrder(context.TODO(), 45, orders.Vegan)},
		{order: orders.NewOrder(context.TODO(), 50, orders.NonVegan)},
		{order: orders.NewOrder(context.TODO(), 10, orders.NonVegan)},
	}
	// add some orders to the terminal
	for _, v := range tests {
		terminal.Put(v.order)
	}

	// close the terminal to signal no more order
	close(terminal.orders)

	expectedInsertedPrices := []int{45, 50, 10}
	for i, price := range expectedInsertedPrices {
		t.Run(fmt.Sprintf("test: %d", i), func(t *testing.T) {
			o, err := terminal.Get()
			if o == nil {
				t.Fatal("expected a non-nil order")
			}
			if err != nil {
				t.Fatalf("expected nil error, got:%v", err)
			}
			if o.Inserted != price {
				t.Errorf("expected: %d, got: %d", price, o.Inserted)
			}
		})
	}
}

func Test_CancelledOrder(t *testing.T) {
	order := orders.NewOrder(context.Background(), 10, orders.Vegan)

	terminal, err := NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	terminal.Put(order)

	// cancel the order
	err = order.Cancel()
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	o, err := terminal.Get()
	if o == nil {
		t.Errorf("expected a non-nil order")
	}
	if err == nil {
		t.Errorf("expected non-nil error")
	}
	if !errors.Is(err, orders.ErrOrderCancelled) {
		t.Errorf("expected error type %v, got %v", orders.ErrOrderCancelled, err)
	}
}

func Test_EmptyTerminal(t *testing.T) {
	t.Parallel() // run the test in parallel
	terminal, err := NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	// terminal is empty and terminal.Get() blocks
	// unless terminal is no longer open or new order comes in
	var err1 error // go routine1 error
	var err2 error // go routine2 error
	var gotOrder, wantOrder *orders.Order

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		gotOrder, err1 = terminal.Get()
	}()

	go func() {
		defer wg.Done()
		wantOrder = orders.NewOrder(context.TODO(), 45, orders.Vegan)
		err2 = terminal.Put(wantOrder)
	}()

	wg.Wait()
	if err1 != nil {
		t.Fatalf("expected nil error, got:%v", err1)
	}
	if err2 != nil {
		t.Fatalf("expected nil error, got:%v", err2)
	}
	// assert that sent order has the same attributes as got order
	if wantOrder.OrderType != gotOrder.OrderType {
		t.Fatalf("expected order type %v, got:%v", wantOrder.OrderType, gotOrder.OrderType)
	}
	if wantOrder.Inserted != gotOrder.Inserted {
		t.Fatalf("expected order Inserted price %v, got:%v", wantOrder.Inserted, gotOrder.Inserted)
	}
}

func Test_ClosedTerminal(t *testing.T) {
	order := orders.NewOrder(context.Background(), 10, orders.Vegan)

	terminal, err := NewTerminal(1)
	if err != nil {
		t.Fatalf("expected error to be nil, got:%v", err)
	}

	terminal.Put(order)

	err = terminal.Close()
	if err != nil {
		t.Fatalf("expected nil error, got:%v", err)
	}

	order, err = terminal.Get()
	if order != nil {
		t.Errorf("expected to nothing from the terminal")
	}
	if !errors.Is(err, ErrTerminalClosed) {
		t.Errorf("expected error type %v, got %v", orders.ErrOrderNil, err)
	}
}
