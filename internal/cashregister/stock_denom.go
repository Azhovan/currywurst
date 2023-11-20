package cashregister

// Define the denominations of notes and coins in cents
// There is no 100 euro in the list because it not common denomination.
const (
	fiveThousandCents = 5000 // A 50 euro note
	twoThousandCents  = 2000 // A 20 euro note
	oneThousandCents  = 1000 // A 10 euro note
	fiveHundredCents  = 500  // A 5 euro note
	twoHundredCents   = 200  // A 2 euro coin
	oneHundredCents   = 100  // A 1 euro coin
	fiftyCents        = 50   // A 50 cent coin
	twentyCents       = 20   // A 20 cent coin
	tenCents          = 10   // A 10 cent coin
	fiveCents         = 5    // A 5 cent coin
	twoCents          = 2    // A 2 cent coin
	oneCent           = 1    // A 1 cent coin
)

// denominations represents the denominations of notes and coins in cents
var denominations = []int{
	fiveThousandCents,
	twoThousandCents,
	oneThousandCents,
	fiveHundredCents,
	twoHundredCents,
	oneHundredCents,
	fiftyCents,
	twentyCents,
	tenCents,
	fiveCents,
	twoCents,
	oneCent,
}

// stocks represents the initial stock of notes and coins in the cash register
var stocks = map[int]int{
	fiveThousandCents: 10,
	twoThousandCents:  10,
	oneThousandCents:  10,
	fiveHundredCents:  10,
	twoHundredCents:   10,
	oneHundredCents:   10,
	fiftyCents:        10,
	twentyCents:       10,
	tenCents:          10,
	fiveCents:         10,
	twoCents:          10,
	oneCent:           10,
}
