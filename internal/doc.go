// Package internal contains supporting packages for the currywurst project.
// These packages are not intended to be imported or used by other modules,
// and they may change or be removed without notice.
//
// The internal package has the following subpackages:
//
// - cashregister: provides a CashRegister type that can calculate and return
// the change for a given price and inserted amount of money.
//
// - orders: provides an Order type that represents a currywurst order with
// a cancellable context and a status channel.
//
// - terminals: provides a Terminal type that represents a queue of orders
// that customers can join and place their orders.
//
// - workers: provides a worker type that can process orders from a terminal
// and return change using a cash register.
package internal
