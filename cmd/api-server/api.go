package api_server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/azhovan/currywurst/internal/orders"
	"github.com/azhovan/currywurst/internal/terminals"
)

// Handler is a struct that handles HTTP requests.
type Handler struct {
	// pins is a map of valid pins.
	// This is for demonstration only and not practical in production.
	// More realistic options would be:
	// - mutual-TLS
	// - user/password combinations saved in external storage (db, etc.)
	// - API access token with expiry date
	// ...
	pins map[string]bool

	// terminals is a map of terminals that receive the customer orders.
	// The key is the name of the terminal (by default it is terminal-1, terminal-2, terminal-3).
	// terminals discover the terminal that customer's request should be sent to.
	terminals map[string]*terminals.Terminal
}

// OrderRequest is a struct type that represents an order request from a customer.
type OrderRequest struct {
	// TerminalId is a string that specifies the id of the terminal that will process the order.
	TerminalId string `json:"terminalId"`
	// OrderType is a string that specifies the type of the order, such as `vegan` or `non-vegan`.
	OrderType string `json:"orderType"`
	// Price specifies the inserted price of the order in cents sent by customer.
	InsertedPrice int `json:"insertedPrice"`
}

// OrderResponse is a struct type that represents an order response to the customer.
type OrderResponse struct {
	// Returned is the amount of money returned to the customer in a human-readable format.
	Returned string `json:"returned"`
}

// NewHandler creates a new Handler with some hardcoded pins.
func NewHandler(terminals map[string]*terminals.Terminal) *Handler {
	return &Handler{
		pins: map[string]bool{
			"1234": true,
			"5678": true,
			"9012": true,
		},
		terminals: terminals,
	}
}

// RegisterRoutes registers the routes for the handler
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/order", h.orderHandler)
}

// orderHandler handles the /order endpoint
func (h *Handler) orderHandler(w http.ResponseWriter, r *http.Request) {
	// check the method and the pin
	if err := h.validateRequest(r); err != nil {
		h.writeJSONError(w, err)
		return
	}

	// process the HTTP request body
	orderRequest, err := h.parseOrderRequest(r)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	// get the terminal by id
	terminal, err := h.getTerminal(orderRequest.TerminalId)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	// send the order to the terminal and wait for the response
	response, err := h.sendOrder(r.Context(), terminal, orderRequest)
	if err != nil {
		h.writeJSONError(w, err)
		return
	}

	// write the response to the client as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(OrderResponse{Returned: response})
}

// validateRequest checks the method and the pin of the request
func (h *Handler) validateRequest(r *http.Request) *httpError {
	// check the method
	if r.Method != http.MethodPost {
		return &httpError{"Method not allowed", http.StatusMethodNotAllowed}
	}

	// check the pin
	pin := r.Header.Get("X-Pin")
	if !h.pins[pin] {
		return &httpError{"Invalid pin", http.StatusUnauthorized}
	}

	return nil
}

// parseOrderRequest decodes the request body into an Order struct
func (h *Handler) parseOrderRequest(r *http.Request) (*OrderRequest, *httpError) {
	orderRequest := OrderRequest{}
	err := json.NewDecoder(r.Body).Decode(&orderRequest)
	if err != nil {
		return nil, &httpError{"Bad request", http.StatusBadRequest}
	}

	if orderRequest.TerminalId == "" {
		return nil, &httpError{"terminalId is missing", http.StatusBadRequest}
	}

	return &orderRequest, nil
}

// getTerminal returns the terminal by id or an error if not found
func (h *Handler) getTerminal(id string) (*terminals.Terminal, *httpError) {
	terminal, ok := h.terminals[id]
	if !ok {
		return nil, &httpError{"invalid terminalId", http.StatusBadRequest}
	}

	return terminal, nil
}

// sendOrder sends the order to the terminal and waits for the response
func (h *Handler) sendOrder(ctx context.Context, terminal *terminals.Terminal, orderRequest *OrderRequest) (string, *httpError) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// build order object out of customers request to send to the terminal
	// and send it to the terminal queue
	order := orders.NewOrder(ctx, orderRequest.InsertedPrice, orders.OrderType(orderRequest.OrderType))
	err := terminal.Put(order)
	if err != nil {
		return "", &httpError{err.Error(), http.StatusUnprocessableEntity}
	}

	// wait until order is ready, or gave up after 10 minutes
	orderTimeout := time.Minute * 10
	err = order.WaitWithTimeout(orderTimeout)
	// order has been processed
	if err == nil {
		er := order.OrderStatus.Error

		if er == nil {
			// the amount of money returned
			// to the customer in a human-readable format
			returned := order.OrderStatus.Returned.Formatted
			return string(returned), nil
		}
		// there was an issue with order, like:
		// - invalid price
		// - no not enough cash in cash register
		// - invalid currywurst type
		//
		// case 1: invalid price
		invalidOrder, ok := er.(*orders.ErrInvalidOrder)
		if ok {
			return "", &httpError{invalidOrder.Error(), http.StatusBadRequest}
		}

		// case 2: invalid order type
		if errors.Is(er, orders.ErrInvalidOrderType) {
			return "", &httpError{er.Error(), http.StatusBadRequest}
		}

		// case 3: not enough cash in the cash register
		return "", &httpError{er.Error(), http.StatusInternalServerError}

	}

	// order has been cancelled by customer, return
	if errors.Is(err, orders.ErrOrderCancelled) {
		return "", &httpError{err.Error(), http.StatusBadRequest}
	}

	// this error indicates that worker is so busy
	// and can't complete order in the given orderTimeout as defined in above
	if errors.Is(err, orders.ErrOrderTimeout) {
		return "", &httpError{err.Error(), http.StatusUnprocessableEntity}
	}

	return "", &httpError{err.Error(), http.StatusInternalServerError}
}

// httpError is a custom error type that contains a message and a status code
type httpError struct {
	Message    string
	StatusCode int
}

// Error implements the error interface
func (e *httpError) Error() string {
	return e.Message
}

// writeJSONError writes an error message and status code to the response as JSON
func (h *Handler) writeJSONError(w http.ResponseWriter, err *httpError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Message})
}
