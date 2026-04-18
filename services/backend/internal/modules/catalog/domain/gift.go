package domain

import (
	"net/url"
	"strings"
	"time"
)

type Gift struct {
	id             GiftID
	categoryID     *CategoryID
	categoryName   string
	name           string
	description    string
	price          Price
	storeLink      string
	image          *string
	ageRestriction *AgeRestriction
	createdAt      time.Time
	updatedAt      time.Time
}

func RestoreGift(
	id string,
	categoryID *string,
	categoryName *string,
	name string,
	description string,
	priceRaw string,
	storeLink string,
	image *string,
	ageRestriction *int,
	createdAt time.Time,
	updatedAt time.Time,
) (Gift, error) {
	giftID, err := NewGiftID(id)
	if err != nil {
		return Gift{}, err
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return Gift{}, ErrGiftNameEmpty
	}

	price, err := NewPrice(priceRaw)
	if err != nil {
		return Gift{}, err
	}

	normalizedStoreLink, err := normalizeURL(storeLink)
	if err != nil {
		return Gift{}, err
	}

	var normalizedCategoryID *CategoryID
	if categoryID != nil && strings.TrimSpace(*categoryID) != "" {
		value, err := NewCategoryID(strings.TrimSpace(*categoryID))
		if err != nil {
			return Gift{}, err
		}

		normalizedCategoryID = &value
	}

	var normalizedCategoryName string
	if categoryName != nil {
		normalizedCategoryName = strings.TrimSpace(*categoryName)
	}

	var normalizedImage *string
	if image != nil {
		trimmedImage := strings.TrimSpace(*image)
		if trimmedImage != "" {
			normalizedImage = &trimmedImage
		}
	}

	var normalizedAgeRestriction *AgeRestriction
	if ageRestriction != nil {
		value, err := NewAgeRestriction(*ageRestriction)
		if err != nil {
			return Gift{}, err
		}

		normalizedAgeRestriction = &value
	}

	return Gift{
		id:             giftID,
		categoryID:     normalizedCategoryID,
		categoryName:   normalizedCategoryName,
		name:           normalizedName,
		description:    strings.TrimSpace(description),
		price:          price,
		storeLink:      normalizedStoreLink,
		image:          normalizedImage,
		ageRestriction: normalizedAgeRestriction,
		createdAt:      createdAt.UTC(),
		updatedAt:      updatedAt.UTC(),
	}, nil
}

func (g Gift) ID() GiftID {
	return g.id
}

func (g Gift) CategoryID() *CategoryID {
	if g.categoryID == nil {
		return nil
	}

	value := *g.categoryID
	return &value
}

func (g Gift) CategoryName() string {
	return g.categoryName
}

func (g Gift) Name() string {
	return g.name
}

func (g Gift) Description() string {
	return g.description
}

func (g Gift) Price() Price {
	return g.price
}

func (g Gift) StoreLink() string {
	return g.storeLink
}

func (g Gift) Image() *string {
	if g.image == nil {
		return nil
	}

	value := *g.image
	return &value
}

func (g Gift) AgeRestriction() *AgeRestriction {
	if g.ageRestriction == nil {
		return nil
	}

	value := *g.ageRestriction
	return &value
}

func (g Gift) CreatedAt() time.Time {
	return g.createdAt
}

func (g Gift) UpdatedAt() time.Time {
	return g.updatedAt
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrStoreLinkEmpty
	}

	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidStoreLink
	}

	return trimmed, nil
}
