package parser

import (
	"strings"

	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
)

var headerAliases = map[string]string{
	"name":              "name",
	"gift_name":         "name",
	"title":             "name",
	"category":          "category",
	"category_name":     "category",
	"price":             "price",
	"cost":              "price",
	"currency":          "currency",
	"description":       "description",
	"short_description": "description",
	"store_link":        "store_link",
	"store link":        "store_link",
	"link":              "store_link",
	"url":               "store_link",
	"store_url":         "store_link",
	"image":             "image",
	"image_link":        "image",
	"image link":        "image",
	"image_url":         "image",
	"age_restriction":   "age_restriction",
	"age restriction":   "age_restriction",
	"age":               "age_restriction",
	"age_min":           "age_restriction",
	"source":            "source",
	"source_name":       "source",
	"store":             "source",
	// gift_id is silently ignored — we generate our own IDs
	"gift_id": "_ignored",
}

var requiredHeaders = map[string]struct{}{
	"name":        {},
	"category":    {},
	"price":       {},
	"description": {},
	"store_link":  {},
}

func normalizeHeader(raw string) string {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if value, ok := headerAliases[normalized]; ok {
		return value
	}

	return normalized
}

func indexHeaders(headers []string) (map[string]int, error) {
	index := make(map[string]int, len(headers))
	for position, header := range headers {
		normalized := normalizeHeader(header)
		if normalized == "" {
			continue
		}

		if _, exists := index[normalized]; !exists {
			index[normalized] = position
		}
	}

	for header := range requiredHeaders {
		if _, ok := index[header]; ok {
			continue
		}

		return nil, errMissingHeader(header)
	}

	return index, nil
}

func rowFromColumns(rowNumber int, index map[string]int, values []string) catalogimportusecase.ImportRowRaw {
	return catalogimportusecase.ImportRowRaw{
		RowNumber:         rowNumber,
		Name:              valueAt(values, index["name"]),
		Category:          valueAt(values, index["category"]),
		PriceRaw:          valueAt(values, index["price"]),
		Currency:          valueAt(values, index["currency"]),
		Description:       valueAt(values, index["description"]),
		StoreLink:         valueAt(values, index["store_link"]),
		Image:             valueAt(values, index["image"]),
		AgeRestrictionRaw: valueAt(values, index["age_restriction"]),
		SourceName:        valueAt(values, index["source"]),
	}
}

func rowFromMap(rowNumber int, values map[string]any) (catalogimportusecase.ImportRowRaw, error) {
	row := catalogimportusecase.ImportRowRaw{RowNumber: rowNumber}

	for key, value := range values {
		normalized := normalizeHeader(key)

		if normalized == "offers" {
			extraOffers, err := parseOffersArray(value)
			if err != nil {
				return catalogimportusecase.ImportRowRaw{}, err
			}
			row.ExtraOffers = extraOffers
			continue
		}

		if normalized == "_ignored" {
			continue
		}

		stringValue, err := scalarToString(value)
		if err != nil {
			return catalogimportusecase.ImportRowRaw{}, errUnsupportedJSONValue(normalized)
		}

		switch normalized {
		case "name":
			row.Name = stringValue
		case "category":
			row.Category = stringValue
		case "price":
			row.PriceRaw = stringValue
		case "currency":
			row.Currency = stringValue
		case "description":
			row.Description = stringValue
		case "store_link":
			row.StoreLink = stringValue
		case "image":
			row.Image = stringValue
		case "age_restriction":
			row.AgeRestrictionRaw = stringValue
		case "source":
			row.SourceName = stringValue
		}
	}

	return row, nil
}

func parseOffersArray(value any) ([]catalogimportusecase.ImportOfferRaw, error) {
	array, ok := value.([]any)
	if !ok {
		return nil, errUnsupportedJSONValue("offers")
	}

	result := make([]catalogimportusecase.ImportOfferRaw, 0, len(array))
	for _, item := range array {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, errUnsupportedJSONValue("offers[]")
		}

		offer := catalogimportusecase.ImportOfferRaw{}
		for k, v := range obj {
			s, err := scalarToString(v)
			if err != nil {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(k)) {
			case "store_name", "store":
				offer.StoreName = s
			case "store_url", "url", "link", "store_link":
				offer.StoreURL = s
			case "price":
				offer.PriceRaw = s
			case "currency":
				offer.Currency = s
			}
		}
		if offer.StoreURL != "" {
			result = append(result, offer)
		}
	}

	return result, nil
}

func valueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}

	return values[index]
}
