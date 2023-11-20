package utils

import (
	"testing"
)

func TestCreateTerminalWorkers(t *testing.T) {
	// create the workers and the terminals
	terminals, cashRegister, err := CreateTerminalWorkers(3)
	if err != nil {
		t.Fatal(err)
	}

	// check the number of terminals
	if len(terminals) != 3 {
		t.Errorf("expected 5 terminals, got %d", len(terminals))
	}

	// check the cash register
	if cashRegister == nil {
		t.Error("expected a cash register, got nil")
	}
}
