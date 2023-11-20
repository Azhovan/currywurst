package orders

// OrderType is a custom type that represents a type of order
type OrderType string

// Define the possible values for the order type
const (
	Vegan    OrderType = "vegan"
	NonVegan OrderType = "non-vegan"
)

// String returns the name of the order type as a string
func (ot OrderType) String() string {
	return string(ot)
}
