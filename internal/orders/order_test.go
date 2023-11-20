package orders

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOrder_Validate(t *testing.T) {
	tests := []struct {
		name    string
		order   *Order
		wantErr error
	}{
		{
			name:    "valid order",
			order:   NewOrder(context.TODO(), 50, Vegan),
			wantErr: nil,
		},
		{
			name:    "invalid order type",
			order:   NewOrder(context.TODO(), 50, OrderType("burger")),
			wantErr: ErrInvalidOrderType,
		},
		{
			name:    "invalid price",
			order:   NewOrder(context.TODO(), int(10), NonVegan),
			wantErr: &ErrInvalidOrder{inserted: int(10), price: int(35)},
		},
		{
			name:    "nil order",
			order:   (*Order)(nil),
			wantErr: ErrOrderNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			// test the custom error type
			er, ok := err.(*ErrInvalidOrder)
			if ok {
				if !errors.As(err, &er) {
					t.Errorf("order.validate() got error:%v, want:%v", err, tt.wantErr)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Errorf("order.validate() got error:%v, want:%v", err, tt.wantErr)
			}
		})
	}
}

func TestOrder_WaitWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*1)
	defer cancel()

	order := NewOrder(ctx, 50, Vegan)
	err := order.WaitWithTimeout(time.Millisecond * 2)
	if !errors.Is(err, ErrOrderCancelled) {
		t.Errorf("order.WaitWithTimeout() got error :%v, want:%v", err, ErrOrderCancelled)
	}
}

func TestNewOrder(t *testing.T) {
	order := NewOrder(context.TODO(), 50, Vegan)
	if order.OrderType != Vegan {
		t.Errorf("order.NewOrder() got orderType:%v, want:%v", order.OrderType, Vegan)
	}
	if order.Inserted != 50 {
		t.Errorf("order.NewOrder() got Inserted:%d, want:%d", order.Inserted, 50)
	}
	if order.Error != nil {
		t.Errorf("order.NewOrder() got error :%v, expected nil", order.Error)
	}
}

func TestOrder_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		order   *Order
		wantErr error
	}{
		{
			name:    "valid order with cancellable context",
			order:   NewOrder(context.WithoutCancel(context.TODO()), 50, Vegan),
			wantErr: nil,
		},
		{
			name: "order with nil cancellable function",
			order: &Order{
				ctx:    context.Background(),
				cancel: nil,
			},
			wantErr: ErrOrderWithInvalidCtx,
		},
		{
			name:    "nil order",
			order:   (*Order)(nil),
			wantErr: ErrOrderNil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Cancel()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("order.Cancel() got error:%v, want:%v", err, tt.wantErr)
			}

			// assert that valid order with cancellable context
			// and cancellable function has been cancelled
			if tt.order != nil &&
				tt.order.ctx != nil &&
				tt.order.cancel != nil {
				select {
				case <-tt.order.ctx.Done():
				// do nothing all good!
				default:
					t.Errorf("order.cancel() didn't cancel the context")
				}
			}
		})
	}
}
