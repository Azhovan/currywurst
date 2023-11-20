package cashregister

import (
	"fmt"
	"testing"
)

func TestStockToReadable(t *testing.T) {
	cents, formatted := stockToCentsAndReadable(map[int]int{
		oneHundredCents: 1,
		fiftyCents:      1,
	})

	expectedFormatted := "1 Euro and 50 Cent"
	if formatted != expectedFormatted {
		t.Errorf("invalid format, expected %s, got:%s", expectedFormatted, formatted)
	}

	expectedCents := 150
	if cents != expectedCents {
		t.Errorf("incorrect cents, expected %d, got:%d", expectedCents, cents)
	}
}

func TestPay(t *testing.T) {
	tests := []struct {
		price, inserted int
		returned        ReturnedAmount
		err             error
	}{
		{
			price:    0,
			inserted: 0,
			returned: ReturnedAmount{}, // empty
			err:      ErrInvalidPayment,
		},
		{
			price:    10,
			inserted: 5,
			returned: ReturnedAmount{},
			err:      ErrInvalidPayment,
		},
		{
			price:    -1,
			inserted: -10,
			returned: ReturnedAmount{},
			err:      ErrInvalidPayment,
		},
		{
			price:    5,
			inserted: 5,
			returned: ReturnedAmount{Cents: 0, Formatted: ""},
			err:      nil,
		},
		{
			price:    5,
			inserted: 10,
			returned: ReturnedAmount{Cents: 5, Formatted: "5 Cent"},
			err:      nil,
		},
	}

	cr := NewCashRegister()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test: %d", i), func(t *testing.T) {
			returned, err := cr.Pay(tt.price, tt.inserted)
			if err != tt.err {
				t.Errorf("expected to get error: %v, got:%v", tt.err, err)
			}
			if returned != tt.returned {
				t.Errorf("expected returned anount:%v got:%v", tt.returned, returned)
			}
		})
	}
}
