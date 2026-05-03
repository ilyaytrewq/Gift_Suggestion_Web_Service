package usecase

type ImportRowRaw struct {
	RowNumber         int
	Name              string
	Category          string
	PriceRaw          string
	Currency          string
	Description       string
	StoreLink         string
	Image             string
	AgeRestrictionRaw string
	SourceName        string
	ExtraOffers       []ImportOfferRaw
}

// ImportOfferRaw holds additional store offers parsed from JSON "offers" array.
type ImportOfferRaw struct {
	StoreName string
	StoreURL  string
	PriceRaw  string
	Currency  string
}
