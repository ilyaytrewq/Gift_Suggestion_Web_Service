package domain

import (
	"strings"
	"time"
)

type Offer struct {
	id        string
	giftID    GiftID
	storeName string
	storeURL  string
	price     Price
	currency  string
	available bool
	createdAt time.Time
}

func RestoreOffer(
	id string,
	giftID GiftID,
	storeName string,
	storeURL string,
	priceCents int64,
	currency string,
	available bool,
	createdAt time.Time,
) (Offer, error) {
	trimmedStore := strings.TrimSpace(storeName)
	if trimmedStore == "" {
		return Offer{}, ErrOfferStoreNameEmpty
	}

	normalizedURL, err := normalizeURL(storeURL)
	if err != nil {
		return Offer{}, err
	}

	trimmedCurrency := strings.TrimSpace(currency)
	if trimmedCurrency == "" {
		return Offer{}, ErrOfferCurrencyEmpty
	}

	if priceCents < 0 {
		return Offer{}, ErrNegativePrice
	}

	return Offer{
		id:        strings.TrimSpace(id),
		giftID:    giftID,
		storeName: trimmedStore,
		storeURL:  normalizedURL,
		price:     Price{cents: priceCents},
		currency:  trimmedCurrency,
		available: available,
		createdAt: createdAt.UTC(),
	}, nil
}

func (o Offer) ID() string {
	return o.id
}

func (o Offer) GiftID() GiftID {
	return o.giftID
}

func (o Offer) StoreName() string {
	return o.storeName
}

func (o Offer) StoreURL() string {
	return o.storeURL
}

func (o Offer) Price() Price {
	return o.price
}

func (o Offer) Currency() string {
	return o.currency
}

func (o Offer) Available() bool {
	return o.available
}

func (o Offer) CreatedAt() time.Time {
	return o.createdAt
}
