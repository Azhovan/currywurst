package cashregister

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	// ErrNotEnoughChange is the error returned when the cash register does not have enough notes and coins to return the change to the customer.
	ErrNotEnoughChange = errors.New("not enough change")

	// ErrInvalidPayment is the error returned when the customer inserts an invalid or insufficient amount of money for the order.
	ErrInvalidPayment = errors.New("invalid payment")
)

// CashRegister represents a cash register that can calculate and return
// the change for a given price and inserted amount of money.
// It has a mutex to lock the access to the stock.
type CashRegister struct {
	mu sync.Mutex
}

// NewCashRegister returns an instance of CashRegister
func NewCashRegister() *CashRegister {
	return &CashRegister{mu: sync.Mutex{}}
}

// ReturnedAmount contains the amount of money returned to the customer.
type ReturnedAmount struct {
	Cents     int
	Formatted string // human-readable format
}

// Pay calculates the change and returns the lowest number of notes and coins possible
// it takes the price and the inserted amount of money as arguments, in cents
func (cr *CashRegister) Pay(price, inserted int) (ReturnedAmount, error) {
	// fast fail
	if (price <= 0 || inserted <= 0) || inserted < price {
		return ReturnedAmount{}, ErrInvalidPayment
	}

	returned := make(map[int]int, 0) // returned change
	amount := inserted - price       // remaining amount of money that needs to be given as change to the customer

	// if the amount is zero, the customer paid the exact price and no change is needed
	if amount == 0 {
		return ReturnedAmount{Cents: 0, Formatted: ""}, nil
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()
	// Loop through the denominations from highest to lowest
	// skip the ones that are out of stock
	for _, denom := range denominations {
		if amount == 0 {
			break
		}

		// calculate how many notes/coins of the current denomination are needed to give the change
		// If the stock has enough of them, add them to the change map and update the stock and amount accordingly
		// If the stock has less than needed, use all of them and set the stock to zero for that denomination
		quantity := amount / denom
		if quantity <= stocks[denom] {
			returned[denom] = quantity
			stocks[denom] -= quantity
			amount -= quantity * denom
		} else {
			returned[denom] = stocks[denom]
			amount -= stocks[denom] * denom
			stocks[denom] = 0
		}
	}

	// there was no enough money to give the change to the customer!
	if amount != 0 {
		return ReturnedAmount{}, ErrNotEnoughChange
	}

	cents, formatted := stockToCentsAndReadable(returned)
	return ReturnedAmount{
		Cents:     cents,
		Formatted: formatted,
	}, nil
}

// stockToCentsAndReadable converts the given stock which is a map of stocks into cents and a human-readable format
func stockToCentsAndReadable(stock map[int]int) (int, string) {
	var cents int
	var result strings.Builder
	for denom, quantity := range stock {
		cents += denom * quantity
	}

	euros := cents / 100
	remainingCents := cents % 100

	if euros > 0 {
		result.WriteString(fmt.Sprintf("%d Euro", euros))
	}
	if remainingCents > 0 {
		if euros > 0 {
			result.WriteString(" and ")
		}
		result.WriteString(fmt.Sprintf("%d Cent", remainingCents))
	}

	return cents, result.String()
}
