package pkg

// OrderType is an interface that represents a type of order that can be processed
// by the cash register. It defines two methods: Name and Price, which return the
// name and the price of the order type, respectively. The cash register package uses
// the OrderType interface to calculate the change and return the lowest number of notes
// and coins possible.
//
// Please note: The OrderType interface is an abstraction that is only relevant to the
// cash register functionality. The order package does not need to know about
// the OrderType interface, because it only deals with the order struct and its fields.
// The reason for this design is to follow the principle of dependency inversion, which
// states that high-level modules should not depend on low-level modules, but both should
// depend on abstractions. The order package is a high-level module, because it contains
// the core domain model of the application. The cash register package is a low-level module,
// because it contains the specific implementation details of the cash register functionality.
type OrderType interface {
	// Name returns the name of the order type, such as "vegan" or "non-vegan".
	Name() string
	// Price returns the price of the order type, in euros.
	Price() int
}

// validOrderTypes is a map that stores the valid order types by their names
var validOrderTypes = map[string]OrderType{
	"vegan":     Vegan{},
	"non-vegan": NonVegan{},
}

// GetOrderType returns the order type by its name, or nil if not found.
func GetOrderType(name string) OrderType {
	order, ok := validOrderTypes[name]
	if !ok {
		return nil
	}

	return order
}

// Vegan is a type of order that is vegan
type Vegan struct{}

// Name returns the name of the vegan order type
func (v Vegan) Name() string {
	return "vegan"
}

// Price returns the price of vegan order type
func (v Vegan) Price() int {
	return 30
}

// NonVegan is a type of order that is not vegan
type NonVegan struct{}

// Name returns the name of the non-vegan order type
func (n NonVegan) Name() string {
	return "non-vegan"
}

// Price returns the price of the non-vegan order type
func (n NonVegan) Price() int {
	return 35
}
