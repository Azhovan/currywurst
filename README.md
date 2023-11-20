## Assumptions

In addition to the existing assumptions, I also consider these assumptions:

- These automated currywurst machine only allows customers to order one item at a time to avoid confusion or errors
  in the order. This means that a customer can’t order two or more currywursts in a single order.
- There are two types of currywurst, `vegan` and `non-vegan` with a **fixed** price at the moment. This is important to
  make
  sure that no one can temper with currywurst price.

## Order Details 
The application supports two types of orders: `vegan` and `non-vegan`. The price for these orders 
are 30 and 35 cents repectively defined [pkg/order_types.go:27](here)

The application has three terminals: `terminal-1`, `terminal-2`, and `terminal-3`. 
Each terminal can handle one order at a time. The customer can choose which terminal to send the order to 
by specifying the terminalId in the request body. For example:
```json
// request body
{
  "terminalId": "terminal-1",
  "orderType": "vegan",
  "price": 500
}
```

## How it works

The application consists of four main components:

- The HTTP handler
- The workers
- The terminals
- The cash register.

#### HTTP Handler

The handler is responsible for handling the HTTP requests from the clients. It validates the request method, the pin,
and the order details. It then sends the order to the corresponding terminal and waits for the response. It finally
writes the response to the client as JSON.

#### Workers

The workers are responsible for processing the orders from the terminals. They use the cash register to
check the inserted price and return the change. They also validate the order type and handle any errors.
They update the order status and notify the handler when the order is ready.

#### Terminals

The terminals are responsible for receiving the orders from the handler and putting them in a queue.
They also provide a way for the workers to access the orders and update their status. They use channels and syn.Cond
to synchronize the communication between the handler and the workers.

#### Cash Register

The cash register is responsible for storing and managing the cash in the system. It provides methods for checking
the inserted price, calculating the change, and updating the stock. It uses a map to store the denominations and their
counts.

## How to run it
To run the application, you need to have Go `1.21.1` or higher installed on your system.
To run the application, you need to use the main command that is located in the cmd/api-server/main folder

```shell 
# run the main.go file
go run cmd/api-server/main/main.go


# or build and run the executable file
go build cmd/api-server/main/main.go -o order-server
./order-server
```

The application will start the server and listen on port `8080`.
The request must be a POST request with a JSON body containing the order details.

``` json
// request body
{
  "terminalId": "terminal-1",
  "orderType": "vegan",
  "insertedPrice": 40
}

// request header
X-Pin: 1234
```

The response will be a JSON body containing the amount of money returned to the customer, or an error message if any.
For example:

```shell
# curl request 
curl -X POST -H "Content-Type: application/json" \
-H "X-Pin: 1234" \
-d '{"terminalId": "terminal-1", "insertedPrice":40, "orderType": "vegan"}' http://localhost:8080/order 

# response 
"10 Cent"
```

## Authentication
The application uses pins to authenticate the customers. The pins are four-digit codes that are sent in the `X-Pin` header.
The application has some valid pins pre-defined [cmd/api-server/api.go:51](here)
