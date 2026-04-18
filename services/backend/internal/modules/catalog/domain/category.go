package domain

import (
	"strings"
	"time"
)

type Category struct {
	id        CategoryID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func RestoreCategory(id, name string, createdAt, updatedAt time.Time) (Category, error) {
	categoryID, err := NewCategoryID(id)
	if err != nil {
		return Category{}, err
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return Category{}, ErrCategoryNameEmpty
	}

	return Category{
		id:        categoryID,
		name:      normalizedName,
		createdAt: createdAt.UTC(),
		updatedAt: updatedAt.UTC(),
	}, nil
}

func (c Category) ID() CategoryID {
	return c.id
}

func (c Category) Name() string {
	return c.name
}

func (c Category) CreatedAt() time.Time {
	return c.createdAt
}

func (c Category) UpdatedAt() time.Time {
	return c.updatedAt
}
