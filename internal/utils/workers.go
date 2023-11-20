package utils

import (
	"strconv"

	"github.com/currywurst/internal/cashregister"
	"github.com/currywurst/internal/terminals"
	"github.com/currywurst/internal/workers"
)

// CreateTerminalWorkers creates the workers and the terminals and returns them as a map and a cash register.
// It takes the number of terminals as an argument and creates a worker and a terminal for each one.
// It also creates a shared cash register for all the terminals.
// It returns a map of terminal ids to terminals, a cash register, and an error if any.
func CreateTerminalWorkers(terminalCount int) (map[string]*terminals.Terminal, *cashregister.CashRegister, error) {
	terminalsMap := map[string]*terminals.Terminal{}
	// cashRegister is the shared cash register between terminals
	cashRegister := cashregister.NewCashRegister()
	// this is just an arbitrary number! for demonstration purposes
	terminalCapacity := 1 << 10

	// create works and associated terminal
	// each worker only manages a specific terminal
	for i := 0; i < terminalCount; i++ {
		terminal, err := terminals.NewTerminal(terminalCapacity)
		if err != nil {
			return nil, nil, err
		}
		terminalId := "terminal-" + strconv.Itoa(i)
		worker := workers.NewWorker(terminal, cashRegister)
		// run the worker in the background
		go worker.Run()
		// update the terminals map
		terminalsMap[terminalId] = terminal
	}

	return terminalsMap, cashRegister, nil
}
