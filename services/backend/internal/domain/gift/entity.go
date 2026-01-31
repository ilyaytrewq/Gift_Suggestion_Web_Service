package gift

import (
	"net"
	"net/url"
	"strings"

	shared2 "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/shared"
)

type GiftID string

type Gift struct {
	id       GiftID
	title    string
	category shared2.CategoryID
	price    shared2.Money
	shopURLs []string
	tags     []shared2.TagID
	ageLimit shared2.AgeLimit
	active   bool
}

func NewGift(id GiftID, title string, cat shared2.CategoryID, price shared2.Money, urls []string, tags []shared2.TagID, limit shared2.AgeLimit) (*Gift, error) {
	if !id.IsValid() {
		return nil, ErrInvalidGiftID
	}
	if title == "" {
		return nil, ErrInvalidTitle
	}
	if len(urls) == 0 {
		return nil, ErrInvalidShopURL
	}
	if !cat.IsValid() {
		return nil, ErrInvalidCategory
	}
	normalizedURLs, err := normalizeShopURLs(urls)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		if !t.IsValid() {
			return nil, ErrInvalidTag
		}
	}
	if !price.IsNonNegative() {
		return nil, shared2.ErrInvalidPrice
	}
	if !limit.IsValid() {
		return nil, shared2.ErrInvalidAgeLimit
	}
	return &Gift{
		id:       id,
		title:    title,
		category: cat,
		price:    price,
		shopURLs: normalizedURLs,
		tags:     shared2.UniqTags(tags),
		ageLimit: limit,
		active:   true,
	}, nil
}

func normalizeShopURLs(raws []string) ([]string, error) {
	if len(raws) == 0 {
		return nil, ErrInvalidShopURL
	}

	uniq := make([]string, 0, len(raws))
	seen := make(map[string]struct{}, len(raws))

	for _, raw := range raws {
		raw = strings.TrimSpace(raw)
		u, err := url.Parse(raw)
		if err != nil {
			return nil, ErrInvalidShopURL
		}
		if u.Scheme == "" || u.Host == "" {
			return nil, ErrInvalidShopURL
		}
		u.Scheme = strings.ToLower(u.Scheme)
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, ErrInvalidShopURL
		}

		u.Fragment = ""
		u.RawFragment = ""

		host := strings.ToLower(u.Hostname())
		port := u.Port()
		if port == "" || (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
			if strings.Contains(host, ":") {
				u.Host = "[" + host + "]"
			} else {
				u.Host = host
			}
		} else {
			u.Host = net.JoinHostPort(host, port)
		}

		normalized := u.String()
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		uniq = append(uniq, normalized)
	}

	return uniq, nil
}
