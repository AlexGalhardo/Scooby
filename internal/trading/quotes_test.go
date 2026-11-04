package trading

import "testing"

func quoteInput() QuoteInput {
	return QuoteInput{
		Mid: D(
			"100",
		),
		Inventory:    Zero,
		MaxInventory: D("10"),
		Gamma:        0.1,
		Sigma:        0.5,
		Horizon:      1,
		Kappa:        1.5,
		Tick:         D("0.01"),
	}
}

func TestQuotesAreSymmetricWithoutInventory(t *testing.T) {
	q, err := AvellanedaStoikov(quoteInput())
	if err != nil {
		t.Fatal(err)
	}
	if !q.Bid.IsPositive() || !q.Ask.GreaterThan(q.Bid) {
		t.Fatalf("a bid below an ask is required: %+v", q)
	}
	below, above := D("100").Sub(q.Bid), q.Ask.Sub(D("100"))
	if below.Sub(above).Abs().GreaterThan(D("0.02")) {
		t.Fatalf("without inventory the quotes straddle the mid within a tick: bid %s ask %s", q.Bid, q.Ask)
	}
}

func TestInventorySkewsTheQuotesAgainstTheInventory(t *testing.T) {
	flat, _ := AvellanedaStoikov(quoteInput())
	long := quoteInput()
	long.Inventory = D("5")
	heavy, _ := AvellanedaStoikov(long)
	if !heavy.Reservation.LessThan(D("100")) || !heavy.Ask.LessThan(flat.Ask) || !heavy.Bid.LessThan(flat.Bid) {
		t.Fatalf("being long must lower the quotes to encourage selling: flat %+v long %+v", flat, heavy)
	}
	short := quoteInput()
	short.Inventory = D("-5")
	light, _ := AvellanedaStoikov(short)
	if !light.Reservation.GreaterThan(D("100")) || !light.Bid.GreaterThan(flat.Bid) {
		t.Fatalf("being short must raise the quotes to encourage buying: %+v", light)
	}
}

func TestInventoryCapStopsQuotingTheGrowingSide(t *testing.T) {
	long := quoteInput()
	long.Inventory = D("10")
	q, _ := AvellanedaStoikov(long)
	if !q.Bid.IsZero() || !q.Ask.IsPositive() {
		t.Fatalf("at the long cap only the ask may rest: %+v", q)
	}
	short := quoteInput()
	short.Inventory = D("-10")
	q, _ = AvellanedaStoikov(short)
	if !q.Ask.IsZero() || !q.Bid.IsPositive() {
		t.Fatalf("at the short cap only the bid may rest: %+v", q)
	}
}

func TestQuotesRefuseNonsenseInputs(t *testing.T) {
	for name, mutate := range map[string]func(*QuoteInput){
		"no gamma":    func(q *QuoteInput) { q.Gamma = 0 },
		"no kappa":    func(q *QuoteInput) { q.Kappa = -1 },
		"no tick":     func(q *QuoteInput) { q.Tick = Zero },
		"no cap":      func(q *QuoteInput) { q.MaxInventory = Zero },
		"bad horizon": func(q *QuoteInput) { q.Horizon = 2 },
		"no mid":      func(q *QuoteInput) { q.Mid = Zero },
	} {
		in := quoteInput()
		mutate(&in)
		if _, err := AvellanedaStoikov(in); err == nil {
			t.Fatalf("%s must be refused", name)
		}
	}
}
