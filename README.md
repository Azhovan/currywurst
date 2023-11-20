## CurryWurst

Currywurst is a Go app that simulates a server for ordering currywurst. It uses terminals, workers,
and a cash register to process the orders and return the change.

## Order Details

The application supports two types of orders: `vegan` and `non-vegan`. The price for these orders
are 30 and 35 cents repectively defined [here](./pkg/order_types.go#L27)

The application has three terminals: `terminal-1`, `terminal-2`, and `terminal-3`.
Each terminal can handle one order at a time. The customer can choose which terminal to send the order to
by specifying the terminalId in the request body. For example:

#### request body

Note insertedPrice is in `cents`

```json
{
  "terminalId": "terminal-1",
  "orderType": "vegan",
  "insertedPrice": 40
}
```

#### request header

```text 
X-Pin: 1234
```

The app will check the inserted price and compare it with the expected price for the order type. If the inserted price
is equal to or greater than the expected price, the app will calculate the change and return it to the customer.
The app uses the following denominations for the change: 1, 2, 5, 10, 20, and 50 cents. The denominations are
defined [here](./internal/cashregister/stock_denom.go)
The app will try to use the smallest number of coins possible to return the change. For example, if the inserted price
is 40 cents and the expected price is 30 cents, the app will return 10 cents as the change.

## Authentication

The application uses pins to authenticate the customers. The pins are four-digit codes that are sent in the `X-Pin`
header.
The application has some valid pins pre-defined [here](./cmd/api-server/api.go#L50)
The app will respond with an error if the pin is invalid or missing. For example:
```json 
{
  "error": "Invalid pin"
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

To run the application, you need to have Go `1.21.1` or higher installed on your system. You also need to clone or download this repo to your local machine.
To run the application, you need to use the main command that is located in the `cmd/api-server/main` folder. You can use 
`go run` or g`o build` to run the `main.go` file or build an executable file. For example:

```shell 
# run the main.go file
go run cmd/api-server/main/main.go


# or build and run the executable file
go build cmd/api-server/main/main.go -o order-server
./order-server
```

The application will start the server and listen on port `8080`.
You can use any HTTP client to send requests to the `/order` endpoint. The request must be a `POST` request with a JSON body and a X-Pin header. For example:

#### request body

``` json
{
  "terminalId": "terminal-1",
  "orderType": "vegan",
  "insertedPrice": 40
}
```

#### request header

```text 
X-Pin: 1234
```

The response will be a JSON body with the change or an error. For example:
```json 
{
  "returned": "10 Cent"
}

```
### Example using CURL

```shell
# curl request
curl -X POST -H "Content-Type: application/json" \
  -H "X-Pin: 1234" \
  -d '{"terminalId": "terminal-1", "insertedPrice":40, "orderType": "vegan"}' http://localhost:8080/order

# response body
{
  "returned": "10 Cent"
}

```

## Dependencies
There is no dependencies to any third party library.
