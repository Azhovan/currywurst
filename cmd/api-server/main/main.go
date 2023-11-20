package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	. "github.com/currywurst/cmd/api-server"
	"github.com/currywurst/internal/utils"
	"github.com/currywurst/internal/workers"
)

func main() {
	// terminalCount is the number of terminals that the handler can handle.
	// It represents the number of terminals that customers can join and place their orders.
	//
	// At the moment, this constant has been hardcoded, but in real world scenarios,
	// it could be injected from configuration files, configmaps, etc.
	const terminalCount = 3
	terminals, cashRegister, err := utils.CreateTerminalWorkers(terminalCount)
	if err != nil {
		log.Fatal(err)
	}

	// creates and run workers for each terminal.
	for _, terminal := range terminals {
		worker := workers.NewWorker(terminal, cashRegister)
		go worker.Run()
	}

	// create the handler a serve mux, and registers the handler
	handler := NewHandler(terminals)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// create a logger that uses the handler and sets the minimum level to error
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// create a server with the logger as the option
	server := NewServer(mux, WithAddr(":8080"), WithLogger(logger))

	err = server.Start()
	if err != nil {
		log.Fatal(err)
	}
}
