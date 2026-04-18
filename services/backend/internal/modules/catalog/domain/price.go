package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type Price struct {
	cents int64
}

func NewPrice(raw string) (Price, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Price{}, ErrPriceEmpty
	}
	if strings.HasPrefix(trimmed, "-") {
		return Price{}, ErrNegativePrice
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) > 2 {
		return Price{}, ErrInvalidPrice
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Price{}, ErrInvalidPrice
	}

	fraction := "00"
	if len(parts) == 2 {
		switch len(parts[1]) {
		case 1:
			fraction = parts[1] + "0"
		case 2:
			fraction = parts[1]
		default:
			return Price{}, ErrInvalidPrice
		}
	}

	frac, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return Price{}, ErrInvalidPrice
	}

	return Price{cents: whole*100 + frac}, nil
}

func (p Price) Cents() int64 {
	return p.cents
}

func (p Price) DecimalString() string {
	return fmt.Sprintf("%d.%02d", p.cents/100, p.cents%100)
}
