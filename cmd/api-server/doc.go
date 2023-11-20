// Package api_server creates and runs the server that handles the order requests from the customers.
// It uses the internal packages to create the workers, the terminals, and the cash register, and to process the orders.
// It also uses the pkg package to validate and format the order types.
// It exposes the /order endpoint that accepts POST requests with a JSON body containing the order details.
// It responds with a JSON body containing the amount of money returned to the customer, or an error message if any.
package api_server
